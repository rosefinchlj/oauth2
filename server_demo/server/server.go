package server

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/go-session/session"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	manager *manage.Manager
	srv     *server.Server
)

// UserInfo 用户信息结构体
type UserInfo struct {
	Username string `json:"username"`
	Gender   string `json:"gender"`
}

// 用一个 map 存储用户信息
var userInfoMap = make(map[string]UserInfo)

func Startup() {
	// 设置 client 信息
	clientStore := store.NewClientStore()
	err := clientStore.Set("youWebApp", &models.Client{ID: "youWebApp", Secret: "xxxxxx", Domain: "http://youWebApp.com"})
	if err != nil {
		log.Println(err)
		return
	}

	// 设置 manager, manager 参与校验 code/accessToken 请求
	manager = manage.NewDefaultManager()

	// 校验 redirect_uri 和 client 的 Domain, 简单起见, 不做校验
	manager.SetValidateURIHandler(func(baseURI, redirectURI string) error {
		return nil
	})

	manager.MustTokenStorage(store.NewMemoryTokenStore())

	// manger 包含 client 信息
	manager.MapClientStorage(clientStore)

	// server 也包含 manger, client 信息
	srv = server.NewServer(server.NewConfig(), manager)

	// 根据 client id 从 manager 中获取 client info, 在获取 accessToken 校验过程中会被用到
	srv.SetClientInfoHandler(func(r *http.Request) (clientID, clientSecret string, err error) {
		clientInfo, err := srv.Manager.GetClient(r.Context(), r.URL.Query().Get("client_id")) //r.URL.Query().Get("client_id")
		if err != nil {
			log.Println(err)
			return "", "", err
		}
		return clientInfo.GetID(), clientInfo.GetSecret(), nil
	})

	// 设置为 authorization code 模式
	srv.SetAllowedGrantType(oauth2.AuthorizationCode)

	// authorization code 模式,  第一步获取code,然后再用code换取 accessToken, 而不是直接获取 accessToken
	srv.SetAllowedResponseType(oauth2.Code)

	// 校验授权请求用户的handler, 会重定向到 登陆页面, 返回"", nil
	srv.SetUserAuthorizationHandler(callBackUserAuthorizationHandler)

	// 校验授权请求的用户的账号密码, 给 LoginHandler 使用, 简单起见, 只对特定用户admin授权
	srv.SetPasswordAuthorizationHandler(callBackPasswordAuthorizationHandler)

	// 允许使用 get 方法请求授权
	srv.SetAllowGetAccessRequest(true)

	// TODO 用户正常注册进来的的信息
	// 模拟有一个注册用户admin, 性别:Male
	userInfoMap["0001"] = UserInfo{
		"admin", "Male",
	}
}

func callBackPasswordAuthorizationHandler(ctx context.Context, clientID, username, password string) (userID string, err error) {
	// TODO 正常的用户登录校验逻辑
	if username == "admin" && password == "123456" {
		return "0001", nil
	}
	return "", errors.New("username or password error")
}

// 下面的方法AuthorizeHandler中srv.HandleAuthorizeRequest内部使用, 用于查看是否有登陆状态
func callBackUserAuthorizationHandler(w http.ResponseWriter, r *http.Request) (userID string, err error) {
	cache, err := session.Start(r.Context(), w, r)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	uid, ok := cache.Get("LoginedUserId")
	// 如果没有查询到登陆状态, 则跳转到登陆页面
	if !ok {
		if r.Form == nil {
			err := r.ParseForm()
			if err != nil {
				log.Println(err)
				return "", err
			}
		}

		// 重定向
		w.Header().Set("Location", "/oauth2/login")
		w.WriteHeader(http.StatusFound)
		return "", nil
	}
	// 若有登录状态, 返回 user id
	userID = uid.(string)
	return userID, nil
}

// AuthorizeHandler 授权入口, index.html 和 agree-auth.html 按下 button 后
func AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
	err := srv.HandleAuthorizeRequest(w, r)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// LoginHandler 登录页面的handler
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	cache, err := session.Start(r.Context(), w, r)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
			return
		}

		userID, err := srv.PasswordAuthorizationHandler(r.Context(), "youWebApp", r.Form.Get("username"), r.Form.Get("password"))
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		cache.Set("LoginedUserId", userID) // 保存登录状态
		err = cache.Save()
		if err != nil {
			log.Println(err)
			return
		}

		// 登录完成后，跳转到 -> 同意授权页面
		w.Header().Set("Location", "/oauth2/agree-auth")
		w.WriteHeader(http.StatusFound)
		return
	}

	// 若请求方法错误, 提供login.html页面
	outputHTML(w, r, "static/login.html")
}

// AgreeAuthHandler 若发现登录状态则提供 agree-auth.html, 否则跳转到 登陆页面
func AgreeAuthHandler(w http.ResponseWriter, r *http.Request) {
	cache, err := session.Start(r.Context(), w, r)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 如果没有查询到登陆状态, 则跳转到 -> 登陆页面
	if _, ok := cache.Get("LoginedUserId"); !ok {
		w.Header().Set("Location", "/oauth2/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	// 如果有登陆状态, 会跳转到 确认授权页面
	outputHTML(w, r, "static/agree-auth.html")
}

// TokenHandler code 换取 accessToken
func TokenHandler(w http.ResponseWriter, r *http.Request) {
	err := srv.HandleTokenRequest(w, r)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// GetUserInfoHandler accessToken 换取用户信息
func GetUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// 获取 accessToken
	accessToken, ok := srv.BearerAuth(r)
	if !ok {
		log.Println("Failed to get accessToken from request")
		return
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
	defer cancelFunc()

	// 从 accessToken 中获取 信息
	tokenInfo, err := srv.Manager.LoadAccessToken(ctx, accessToken)
	if err != nil {
		log.Println(err)
		return
	}

	// 获取 user id
	userID := tokenInfo.GetUserID()
	grantScope := tokenInfo.GetScope()

	// 根据 grant scope 决定获取哪些用户信息
	if grantScope != "read_user_info" {
		log.Println("invalid grant scope")
		size, err := w.Write([]byte("invalid grant scope"))
		if err != nil {
			log.Println("write byte size: ", size, " err:", err)
			return
		}
		return
	}

	userInfo := userInfoMap[userID]
	resp, _ := json.Marshal(userInfo)
	w.Write(resp) //nolint:errcheck
}

// 提供 HTML 文件显示
func outputHTML(w http.ResponseWriter, req *http.Request, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}
	defer file.Close()
	fi, _ := file.Stat()
	http.ServeContent(w, req, file.Name(), fi.ModTime(), file)
}
