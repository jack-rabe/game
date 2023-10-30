package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
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

var lobbyConnections map[uuid.UUID]UserConnection
var upgrader = websocket.Upgrader{CheckOrigin: allowCorsSocket}
var games []Game

func allowCorsSocket(r *http.Request) bool {
	return true
}

func cleanupUserOnClose(id uuid.UUID) {
	removedConn := lobbyConnections[id]
	delete(lobbyConnections, id)
	for _, userConn := range lobbyConnections {
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
	lobbyConnections[id] = UserConnection{user: &User{Id: id.URN()}, conn: conn}
	defer delete(lobbyConnections, id)
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
	for _, userConn := range lobbyConnections {
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
		createGame(gameId)
	}
	// broadcast info that user has joined game
	for _, userConn := range lobbyConnections {
		userConn.conn.WriteJSON(struct {
			Name      string `json:"name"`
			Type      string `json:"type"`
			GameReady bool   `json:"gameReady"`
			GameId    string `json:"gameId"`
		}{Name: user.Name, Type: "join", GameReady: fourUsersReady, GameId: gameId})
	}
	// clear lobby if new game has been created
	if fourUsersReady {
		for c := range lobbyConnections {
			delete(lobbyConnections, c)
		}
	}
}

func getLobbyUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GET /getUsers")

	var userNames = make([]string, 0)
	for _, userConn := range lobbyConnections {
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

// ----------------------------------- Individual Games

type Game struct {
	Id          string
	Connections map[string]*websocket.Conn
}

// TODO figure out what I want as strings and what i want as UUIDs
func createGame(id string) {
	game := Game{Id: id}
	game.Connections = make(map[string]*websocket.Conn)

	for _, userConn := range lobbyConnections {
		name := userConn.user.Name
		if name != "" {
			game.Connections[name] = nil
		}
	}
	fmt.Printf("New game created with ID: %v\n", id)
	games = append(games, game)
}

func joinGame(w http.ResponseWriter, r *http.Request) {
	gameId := mux.Vars(r)["gameId"]
	name := r.URL.Query().Get("name")
	fmt.Printf("CONNECT /joinGame/%v name: %v\n", gameId, name)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	gameFound := false
	for _, game := range games {
		if gameId == game.Id {
			game.Connections[name] = conn
			gameFound = true
		}
	}
	if !gameFound {
		fmt.Println(err)
		return
	}

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("socket closing: ", err)
			return
		}
	}
}

// ----------------------------------- Individual Games

func main() {
	lobbyConnections = make(map[uuid.UUID]UserConnection)
	games = make([]Game, 0)

	mux := mux.NewRouter()
	handler := getCorsOptions().Handler(mux)

	mux.HandleFunc("/ws", handleSocket)
	mux.HandleFunc("/addUser", addUser).
		Methods("POST")
	mux.HandleFunc("/getLobbyUsers", getLobbyUsers).
		Methods("GET")
	mux.HandleFunc("/joinGame/{gameId}", joinGame)

	fmt.Printf("listening on port %v\n", PORT)
	if err := http.ListenAndServe(":3333", handler); err != nil {
		log.Fatal(err)
	}
}
