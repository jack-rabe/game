package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

var upgrader = websocket.Upgrader{CheckOrigin: allowCorsSocket}
var users []string
var connections map[uuid.UUID]*websocket.Conn

type User struct {
	Name string
}

func allowCorsSocket(r *http.Request) bool {
	return true
}

func handleSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	id := uuid.New()
	connections[id] = conn
	defer delete(connections, id)

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("socket closing: ", err)
			return
		}

		fmt.Println("message type", messageType)
		fmt.Println(string(message))
	}
}

func addUser(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("add user request\n")

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil && err != io.EOF {
		fmt.Println(err)
		return
	}

	users = append(users, user.Name)
	for _, conn := range connections {
		conn.WriteJSON(user)
	}
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("get user request\n")

	encoder := json.NewEncoder(w)
	encoder.Encode(users)
}

func main() {
	connections = make(map[uuid.UUID]*websocket.Conn)

	mux := http.NewServeMux()

	corsOptions := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})
	handler := corsOptions.Handler(mux)

	mux.HandleFunc("/ws", handleSocket)
	mux.HandleFunc("/addUser", addUser)
	mux.HandleFunc("/getUsers", getUsers)

	if err := http.ListenAndServe(":3333", handler); err != nil {
		log.Fatal(err)
	}
}
