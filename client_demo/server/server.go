package server

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var accessTokenEntity *getAccessTokenRsp = new(getAccessTokenRsp)

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

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	outputHTML(w, r, "static/index.html")
}

// CallBackHandler 这里获取授权码(code)
func CallBackHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("github redirect")

	if err := r.ParseForm(); err != nil {
		log.Println(err)
		return
	}

	code := r.Form.Get("code")
	log.Println("receive code:", code)

	var err error
	accessTokenEntity, err = getAccessToken(code)
	if err != nil {
		log.Println(err)
	}

	log.Printf("%+v\n", accessTokenEntity)
}

func getAccessToken(code string) (*getAccessTokenRsp, error) {
	req := getAccessTokenReq{
		ClientId:     clientId,
		ClientSecret: clientSecrets,
		Code:         code,
		RedirectUri:  "",
	}
	body, _ := json.Marshal(req)

	client := http.Client{Timeout: time.Second * 8}
	request, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("accept", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	rsp := &getAccessTokenRsp{}
	err = json.NewDecoder(response.Body).Decode(rsp)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

func GetUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	client := http.Client{Timeout: time.Second * 8}
	request, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return
	}
	request.Header.Add("accept", "application/json")
	request.Header.Add("Authorization", accessTokenEntity.TokenType+" "+accessTokenEntity.AccessToken)

	response, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return
	}

	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	w.Write(body)
}
