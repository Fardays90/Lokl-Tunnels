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
var myId string

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

func tunnelClient() {
	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://localhost:8080/connect", nil)
	if err != nil {
		fmt.Println("Error trying to construct the http request err:" + err.Error())
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error trying to reach aea err:" + err.Error())
		return
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error trying to read the response body")
		return
	}
	defer resp.Body.Close()
	myId = string(bodyBytes)
	fmt.Println("Tunnel has been set up, your link is http://localhost:8080/" + myId)
	fmt.Println("Make requests to this as the starter to reach your endpoint.")
}
func listenForMessages(conn *websocket.Conn) {
	fmt.Println("Waiting for requests")
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
		fmt.Printf("Got the request %s %s", localRequest.Method, localRequest.URL.Path)
	}
}

func main() {
	fmt.Scanf("http --port %s", &myport)
	fmt.Printf("Port: %s", myport)
	// tunnelClient()
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/connect", nil)
	if err != nil {
		fmt.Println("Websocket connection failed err: " + err.Error())
		return
	}
	defer conn.Close()
	listenForMessages(conn)
}
