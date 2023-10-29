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

const PORT = 3333

type User struct {
	Name string
	Id   string
}

// Name -> socketId
var users map[string]string
var connections map[uuid.UUID]*websocket.Conn
var upgrader = websocket.Upgrader{CheckOrigin: allowCorsSocket}

func allowCorsSocket(r *http.Request) bool {
	return true
}

func cleanupUserOnClose(id uuid.UUID) {
	var nameToRemove string
	for name, socketId := range users {
		if socketId == id.URN() {
			nameToRemove = name
			break
		}
	}
	if nameToRemove != "" {
		delete(users, nameToRemove)
		for _, conn := range connections {
			conn.WriteJSON(struct {
				Name string `json:"name"`
				Type string `json:"type"`
			}{Name: nameToRemove, Type: "leave"})
		}
	}
}

func handleSocket(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handling socket connection")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	id := uuid.New()
	connections[id] = conn
	defer delete(connections, id)
	defer cleanupUserOnClose(id)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("socket closing: ", err)
			return
		}
		conn.WriteJSON(struct {
			Id   string `json:"id"`
			Type string `json:"type"`
		}{Id: id.URN(), Type: "id"})
	}
}

func addUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("add user request")

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil && err != io.EOF {
		fmt.Println(err)
		return
	}

	users[user.Name] = user.Id
	for _, conn := range connections {
		fourUsersReady := len(users) == 4
		conn.WriteJSON(struct {
			Name      string `json:"name"`
			Type      string `json:"type"`
			GameReady bool   `json:"gameReady"`
		}{Name: user.Name, Type: "join", GameReady: fourUsersReady})
	}
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("get user request")

	var userKeys = make([]string, 0)
	for u := range users {
		userKeys = append(userKeys, u)
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(userKeys)
}

func getCorsOptions() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})
}

func main() {
	connections = make(map[uuid.UUID]*websocket.Conn)
	users = make(map[string]string)

	mux := http.NewServeMux()
	handler := getCorsOptions().Handler(mux)

	mux.HandleFunc("/ws", handleSocket)
	mux.HandleFunc("/addUser", addUser)
	mux.HandleFunc("/getUsers", getUsers)

	fmt.Printf("listening on port %v\n", PORT)
	if err := http.ListenAndServe(":3333", handler); err != nil {
		log.Fatal(err)
	}
}
