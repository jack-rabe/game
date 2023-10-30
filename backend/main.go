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

type User struct {
	Name string
	Id   string
}

type UserConnection struct {
	user *User
	conn *websocket.Conn
}

const PORT = 3333

var connections map[uuid.UUID]UserConnection
var upgrader = websocket.Upgrader{CheckOrigin: allowCorsSocket}

func allowCorsSocket(r *http.Request) bool {
	return true
}

func cleanupUserOnClose(id uuid.UUID) {
	removedConn := connections[id]
	delete(connections, id)
	for _, userConn := range connections {
		userConn.conn.WriteJSON(struct {
			Name string `json:"name"`
			Type string `json:"type"`
		}{Name: removedConn.user.Name, Type: "leave"})
	}
}

func handleSocket(w http.ResponseWriter, r *http.Request) {
	fmt.Println("CONNECT /ws")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	id := uuid.New()
	connections[id] = UserConnection{user: &User{Id: id.URN()}, conn: conn}
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
	fmt.Println("POST /addUser")

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil && err != io.EOF {
		// TODO give appropriate errors for failed requests
		fmt.Println(err)
		return
	}

	playersJoined := 0
	for _, userConn := range connections {
		// TODO give appropriate errors for failed requests
		if user.Name == userConn.user.Name {
			fmt.Println("duplicate name")
			return
		}
		if userConn.user.Id == user.Id {
			userConn.user.Name = user.Name
		}
		if userConn.user.Name != "" {
			playersJoined++
		}
	}
	fourUsersReady := playersJoined == 4
	var gameId string
	if fourUsersReady {
		gameId = uuid.New().URN()[9:]
	}
	for _, userConn := range connections {
		userConn.conn.WriteJSON(struct {
			Name      string `json:"name"`
			Type      string `json:"type"`
			GameReady bool   `json:"gameReady"`
			GameId    string `json:"gameId"`
		}{Name: user.Name, Type: "join", GameReady: fourUsersReady, GameId: gameId})
	}
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GET /getUsers")

	var userNames = make([]string, 0)
	for _, userConn := range connections {
		if userConn.user.Name != "" {
			userNames = append(userNames, userConn.user.Name)

		}
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(userNames)
}

func getCorsOptions() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})
}

func main() {
	connections = make(map[uuid.UUID]UserConnection)

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
