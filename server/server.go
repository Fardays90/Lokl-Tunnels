package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var targetPort string
var connections = make(map[string]*websocket.Conn)
var charset = "abcdefghijklmnoqprstyxz1234567890"
var serverPort string
var upgrader = websocket.Upgrader{}
var pendingResponses = sync.Map{}

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

func generateRandomId(length int) string {
	id := make([]byte, length)
	for i := range id {
		id[i] = charset[rand.Intn(len(charset))]
	}
	return string(id)
}
func handleTunnelClient(w http.ResponseWriter, r *http.Request) {
	// if r.Method != "POST" {
	// 	log.Println("Please send a post request ")
	// 	return
	// }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error trying to upgrade http connection to websocket err:" + err.Error())
		return
	}
	uniqueId := generateRandomId(5)
	connections[uniqueId] = conn
	log.Println("Tunnel has been made between aea and http://localhost:" + targetPort)
	log.Println("Your path is http://localhost" + serverPort + "/" + uniqueId + "/")
	// uniqueIdBytes := []byte(uniqueId)
	// w.Write(uniqueIdBytes)
	for {
		var response HTTPRes
		err := conn.ReadJSON(&response)
		if err != nil {
			fmt.Println("Error trying to read the json err: " + err.Error())
			break
		}
		val, ok := pendingResponses.Load(response.ID)
		if !ok {
			fmt.Println("No pending response channel found for ID: ", response.ID)
			continue
		}
		responseChannel, ok := val.(chan HTTPRes)
		if !ok {
			fmt.Println("Pending response is not a channel for ID: ", response.ID)
			continue
		}
		responseChannel <- response
		pendingResponses.Delete(response.ID)
	}
}
func handleExternalReqs(w http.ResponseWriter, r *http.Request) {
	id := strings.Split(r.URL.Path, "/")[1]
	Path := r.URL.Path
	log.Println("Making req to -> " + id)
	tunnelToFind, found := connections[id]
	if !found {
		log.Println("Tunnel does not exist or wrong id")
		return
	}
	method := r.Method
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading request body err:" + err.Error())
		return
	}
	defer r.Body.Close()
	requestHeaders := r.Header
	reqId := generateRandomId(6)
	clientReq := HTTPReq{ID: reqId, Method: method, Body: bodyBytes, Path: Path, Headers: requestHeaders}
	responseChan := make(chan HTTPRes)
	pendingResponses.Store(clientReq.ID, responseChan)
	tunnelToFind.WriteJSON(clientReq)
	response := <-responseChan
	for k, vv := range response.Headers {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(response.StatusCode)
	w.Write(response.Body)
}

func main() {
	http.HandleFunc("/connect", handleTunnelClient)
	http.HandleFunc("/", handleExternalReqs)
	serverPort = ":8080"
	fmt.Println("Server started at localhost:" + serverPort)
	err := http.ListenAndServe(serverPort, nil)
	if err != nil {
		fmt.Println("Error starting server at localhost" + serverPort)
		return
	}
}
