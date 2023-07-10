package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

var wsChan = make(chan WsPayload)

var clients = make(map[WebSocketConnection]string)

var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode(),
)

var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func Home(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "home.jet", nil)
	if err != nil {
		log.Println(err)
	}
}

type WebSocketConnection struct {
	*websocket.Conn
}

// WsJsonResponse defines the response sent back from websocket
type WsJsonResponse struct {
	Action         string   `json:"action"`
	Message        string   `json:"message"`
	MessageType    string   `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
}

type WsPayload struct {
	Action   string              `json:"action"`
	Username string              `json:"username"`
	Message  string              `json:"message"`
	Conn     WebSocketConnection `json:"-"`
}

// Upgrade connection to websocket
func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("Client connected to endpoint")

	var response WsJsonResponse
	response.Message = `<em><small>Connected to Server</small></em>`

	conn := WebSocketConnection{Conn: ws}
	clients[conn] = ""

	err = ws.WriteJSON(response)
	if err != nil {
		log.Println(err)
	}

	go ListenForWas(&conn)

}

func ListenForWas(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error", fmt.Sprintf("%v", r))
		}
	}()

	var payload WsPayload

	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			//do nothing
		} else {
			payload.Conn = *conn
			wsChan <- payload
		}
	}
}

func ListenToWsChannel() {
	var response WsJsonResponse

	for {
		e := <-wsChan

		switch e.Action {
		case "username":
			//get a list of all users and sent it back via broadcast
			clients[e.Conn] = e.Username
			users := getUserList()
			response.Action = "list_users"
			response.ConnectedUsers = users
			broadcastToAll(response)
		case "left":
			response.Action = "list_users"
			delete(clients, e.Conn)
			users := getUserList()
			response.ConnectedUsers = users
			broadcastToAll(response)
		case "broadcast":
			response.Action = "broadcast"
			response.Message = fmt.Sprintf("<strong>%s</string>: %s", e.Username, e.Message)
			broadcastToAll(response)
		}

	}
}

func getUserList() []string {
	var userList []string
	for k, x := range clients {
		fmt.Println(k, "userConn")
		if x != "" {
			userList = append(userList, x)
		}
	}
	sort.Strings(userList)
	return userList
}

func broadcastToAll(response WsJsonResponse) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			log.Println("websocket err")
			_ = client.Close()
			delete(clients, client)
		}
	}
}

func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		log.Println(err)
		return err
	}

	err = view.Execute(w, data, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Yes, you can store the client's WebSocket connection information in a database or Redis to make your application stateless and enable scaling across multiple pods in Kubernetes. Storing the WebSocket connections allows you to maintain the state of connected clients and communicate with them even if the client connects to a different server pod.

// Here are the general steps to achieve this:

// 1. Choose a database or Redis as your storage solution.
// Redis is often a good choice for real-time applications due to its in-memory data structure store
// and pub/sub capabilities.

// 2. Modify your code to store the WebSocket connection
// information in the chosen storage solution.
// Instead of using the `clients` map, you can store
// the connection details in the database or Redis using a suitable data structure.

// 3. Update the `WsEndpoint` function to retrieve the WebSocket
// connection details from the storage solution instead of the
// `clients` map. This ensures that each WebSocket connection can be associated with the correct client.

// 4. Modify the `ListenToWsChannel` function to read and
// write WebSocket messages using the stored connection details
// from the database or Redis.

// By storing the WebSocket connections in a database or Redis, you can
// horizontally scale your application across multiple pods in Kubernetes.
// Each pod can then retrieve the necessary connection details from the shared
// storage to communicate with the connected clients.
// This approach ensures that client information is not tied to a specific server
// and allows for a stateless architecture.

// When the front-end client refreshes the page, the WebSocket connection is indeed disconnected and a new connection is established. In this scenario, you need a mechanism to identify and recognize the user as the same user who had the previous WebSocket connection.

// One common approach is to use authentication and session management. When the user initially connects to the WebSocket, you can authenticate the user and generate a unique session identifier. This session identifier can be stored in a cookie or local storage on the client-side.

// When the client refreshes the page and establishes a new WebSocket connection, the client can send the session identifier as part of the connection request. On the server-side, you can validate the session identifier to associate the new WebSocket connection with the same user.

// Here's a high-level overview of how this can be implemented:

// During the initial WebSocket connection, authenticate the user and generate a session identifier. Store this identifier on the server-side along with the associated WebSocket connection details.

// Send the session identifier to the client and store it in a cookie or local storage.

// When the client refreshes the page and establishes a new WebSocket connection, include the session identifier in the connection request headers or query parameters.

// On the server-side, validate the session identifier. If it is valid, associate the new WebSocket connection with the user identified by the session identifier.

// By using session management and associating WebSocket connections with user sessions, you can recognize the same user even after a page refresh or when establishing a new WebSocket connection.
