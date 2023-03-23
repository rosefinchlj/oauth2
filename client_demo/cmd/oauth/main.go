package main

import (
	"client_demo/server"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", server.IndexHandler)
	http.HandleFunc("/oauth/redirect", server.CallBackHandler)
	http.HandleFunc("/oauth/getUserInfo", server.GetUserInfoHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
