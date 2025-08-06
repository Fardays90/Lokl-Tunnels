package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

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

var token string

func listenForMessages(conn *websocket.Conn) {
	fmt.Println("Waiting for requests")
	var sentId idJson
	err := conn.ReadJSON(&sentId)
	if err != nil {
		fmt.Println("Error trying to read id")
		return
	}
	fmt.Println("Tunnel has been set up.")
	fmt.Println("Your link: http://" + sentId.Id + ".tunnels." + "fardays.com/")
	for {
		var req HTTPReq
		err := conn.ReadJSON(&req)
		if err != nil {
			fmt.Println("Error trying to read the json sent from server err:" + err.Error())
			break
		}
		path := req.Path
		URL := "http://localhost:" + myport + path
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
		fmt.Printf("Got the request %s %s \n", localRequest.Method, URL)
	}
}

func main() {
	fmt.Println("Welcome to lokl cli please enter the command http --port <port> to create a tunnel")
	fmt.Scanf("http --port %s", &myport)
	fmt.Printf("Port: %s \n", myport)
	headers := http.Header{}
	headers.Add("Authorization", `Bearer `+token)
	conn, _, err := websocket.DefaultDialer.Dial("ws://tunnels.fardays.com/connect", headers)
	if err != nil {
		fmt.Println("Websocket connection failed err: " + err.Error())
		return
	}
	defer conn.Close()
	listenForMessages(conn)
}
