package main

import (
	"bytes"
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
var channels = make(map[string] chan HTTPRes)
var pendingResponses = sync.Map{};
type HTTPReq struct{
	ID string `json:"id"`
	Method string `json:"method"`
	Body []byte `json:"body"`
	Path string `json:"path"`
	Headers map[string][]string `json:"headers"`
}

type HTTPRes struct{
	ID string `json:"id"`
	StatusCode string `json:"status_code"`
	Body []byte `json:"body"`
	Headers map[string][]string `json:"headers"`
}

func generateRandomId(length int) string {
	id := make([]byte, length)
	for i := range id {
		id[i] = charset[rand.Intn(len(charset))]
	}
	return string(id)
}

func waitForResponse(id string){
	ch, ok := pendingResponses.Load(id);
	if !ok{
		log.Println("Internal Sever Error occured")
		return
	}
	select{

	}
}

func handleTunnelClient(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w,r,nil);
	if err != nil{
		fmt.Println("Error trying to upgrade http connection to websocket err:"+err.Error());
		return;
	}
	uniqueId := generateRandomId(5);
	connections[uniqueId] = conn;
	fmt.Println("Tunnel has been made between aea and http://localhost:"+targetPort);
	fmt.Println("Your path is http://localhost"+serverPort+"/"+uniqueId+"/");
	for{
		var req HTTPReq;
		err := conn.ReadJSON(&req);
		if err != nil{
			fmt.Println("Error trying to read the json from websocket server err:"+err.Error());
			break;
		}
		path := req.Path;
		client := &http.Client{};
		convertedToHTTP, err := http.NewRequest(req.Method, "http://localhost:"+targetPort+path, bytes.NewReader(req.Body));
		if err != nil{
			fmt.Println("Error trying to convert json to http request err:"+err.Error());
			continue
		}
		resp, err := client.Do(convertedToHTTP);
		if err != nil{
			fmt.Println("Error trying to make the http request to tunnel client err:"+err.Error());
		}
		bodyBytes, err := io.ReadAll(resp.Body);
		if err != nil{
			fmt.Println("Error trying to read response body from the client tunnel err:"+err.Error());
			continue;
		}
		defer resp.Body.Close();
		localResponse := HTTPRes{ID: req.ID, StatusCode: resp.StatusCode, Body: bodyBytes, Headers: resp.Header};
		err = conn.WriteJSON(localResponse);
		if err != nil{
			fmt.Println("Error while trying to write json to websocket server from tunnel client err:"+err.Error());
			break;
		}
	}
}
func handleExternalReqs(w http.ResponseWriter, r *http.Request) {
	id := strings.Split(r.URL.Path, "/")[1]
	Path := r.URL.Path
	log.Println("Making req to -> " + id);
	tunnelToFind, found := connections[id];
	if !found{
		log.Println("Tunnel does not exist or wrong id");
		return
	}
	method := r.Method;
	bodyBytes, err := io.ReadAll(r.Body);
	if err != nil{
		log.Println("Error reading request body err:"+err.Error())
		return
	}
	defer r.Body.Close()
	requestHeaders := r.Header;
	clientReq := HTTPReq{ID: id, Method: method, Body:bodyBytes, Path: Path, Headers: requestHeaders}
	responseChan := make(chan HTTPRes);
	pendingResponses.Store(clientReq.ID, responseChan);
	tunnelToFind.WriteJSON(clientReq);
	response :=
}

func main() {
	fmt.Scanf("http --port %s", &targetPort)
	fmt.Println("Will create tunnel to localhost:" + targetPort)
	http.HandleFunc("/connect", handleTunnelClient)
	serverPort = ":8080"
	fmt.Println("Server started at localhost:" + serverPort)
	err := http.ListenAndServe(serverPort, nil)
	if err != nil {
		fmt.Println("Error starting server at localhost" + serverPort)
		return
	}
}
