package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var myport string

type HTTPReq struct {
	ID      string              `json:"id"`
	Method  string              `json:"method"`
	Body    []byte              `json:"body"`
	Path    string              `json:"path"`
	Headers map[string][]string `json:"headers"`
}

type HTTPRes struct {
	ID         string              `json:"id"`
	StatusCode int                 `json:"status_code"`
	Body       []byte              `json:"body"`
	Headers    map[string][]string `json:"headers"`
}

type idJson struct {
	Id string `json:"id"`
}

func listenForMessages(conn *websocket.Conn) {
	fmt.Println("Waiting for requests")
	var sentId idJson
	err := conn.ReadJSON(&sentId)
	if err != nil {
		fmt.Println("Error trying to read id")
		return
	}
	fmt.Println("Tunnel has been set up.")
	fmt.Println("Your link: http://localhost:8080/" + sentId.Id)
	for {
		var req HTTPReq
		err := conn.ReadJSON(&req)
		if err != nil {
			fmt.Println("Error trying to read the json sent from server")
			break
		}
		pathWithoutId := strings.Split(req.Path, "/")[2]
		URL := "http://localhost:" + myport + "/" + pathWithoutId
		fmt.Println(URL)
		localRequest, err := http.NewRequest(req.Method, URL, bytes.NewReader(req.Body))
		if err != nil {
			fmt.Println("Error trying to construct the http request err:" + err.Error())
			return
		}
		localRequest.Header = req.Headers
		client := &http.Client{}
		resp, err := client.Do(localRequest)
		if err != nil {
			fmt.Println("Error trying to do the http reques to the origin err:" + err.Error())
			break
		}
		responseBodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error trying to read the response body from the origin err:" + err.Error())
			break
		}
		defer resp.Body.Close()
		localResponse := HTTPRes{ID: req.ID, StatusCode: resp.StatusCode, Body: responseBodyBytes, Headers: resp.Header}
		err = conn.WriteJSON(localResponse)
		if err != nil {
			fmt.Println("Error trying to write json to server err: " + err.Error())
			break
		}
		fmt.Printf("Got the request %s %s", localRequest.Method, URL)
	}
}

func main() {
	fmt.Scanf("http --port %s", &myport)
	fmt.Printf("Port: %s \n", myport)
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/connect", nil)

	if err != nil {
		fmt.Println("Websocket connection failed err: " + err.Error())
		return
	}
	defer conn.Close()
	listenForMessages(conn)
}
