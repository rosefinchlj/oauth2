// 使用go-oauth2实现的Oauth2授权服务

package main

import (
	"fmt"
	"net/http"
	"oauth_server/server"
)

func main() {
	server.Startup()

	// auth_server 授权入口, 这里生成code
	http.HandleFunc("/oauth2/authorize", server.AuthorizeHandler)

	// auth_server 发现未登录状态, 跳转到的登录handler
	http.HandleFunc("/oauth2/login", server.LoginHandler)

	// auth_server拿到 client以后重定向到的地址, 也就是 auth_client 获取到了code, 准备用code换取accessToken
	//http.HandleFunc("/oauth2/code_to_token", server.CodeToToken)

	// auth_server 处理由code换取accessToken的handler
	// 由第三方应用的后台服务来请求(demo是在前端直接请求获取)
	http.HandleFunc("/oauth2/token", server.TokenHandler)

	// 登录完成, 同意授权的页面，然后再次进入授权 -> /oauth2/authorize -> 服务检测到已经登录过，生成code后重定向
	http.HandleFunc("/oauth2/agree-auth", server.AgreeAuthHandler)

	// accessToken换取用户信息的handler
	http.HandleFunc("/oauth2/getuserinfo", server.GetUserInfoHandler)

	http.Handle("/", http.FileServer(http.Dir("./static/"))) // http://localhost:9000/youWebApp.html

	errChan := make(chan error)

	go func() {
		errChan <- http.ListenAndServe(":9000", nil)
	}()

	err := <-errChan
	if err != nil {
		fmt.Println("Hello server stop running.")
	}
}
