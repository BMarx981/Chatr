package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool) //connected clients
var broadcast = make(chan string)            //broadcast channel
var upgrader = websocket.Upgrader{}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	//Upgrade inital get request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	//Make sure we close the connection when the function returns
	defer ws.Close()

	//regiser our new Client
	clients[ws] = true

	for {
		_, bArr, erar := ws.ReadMessage()
		var msg string = string(bArr)
		if erar != nil {
			log.Fatal(erar)
			delete(clients, ws)
			break
		}
		//Send the newly received message to the broadcast channel
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		//Grab the next message from the broadcast channel
		msg := <-broadcast

		for client := range clients {
			err := client.WriteMessage(1, []byte(msg))
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func main() {
	//Configure websocket route
	http.HandleFunc("/", handleConnections)
	//Start listening for incoming chat messages
	go handleMessages()

	//Start the server on localhost port 8080 and log any errors
	log.Println("http server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
