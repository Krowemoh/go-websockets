package main

import (
    "fmt"
    "net/http"
    "github.com/gorilla/websocket"
)

type Message struct {
    Room string `json:"room"`
    Type string `json:"type"`
    Username string `json:"username"`
    Message string `json:"message"`
}

type Client struct {
    Room string
    Username string
}

var clients = make(map[*websocket.Conn]Client)
var processMessage = make(chan Message)

var upgrader = websocket.Upgrader {
    ReadBufferSize: 0,
    WriteBufferSize: 0,
}

func wsConnections(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer conn.Close();

    var msg Message
    err = conn.ReadJSON(&msg)

    if err != nil {
        fmt.Println(err)
        return
    }

    if msg.Type == "join" {
        clients[conn] = Client {
            Room: msg.Room,
            Username: msg.Username,
        }
    }

    for {
        err = conn.ReadJSON(&msg)
        if err != nil {
            fmt.Println(err)
            delete(clients, conn)
            return
        }

        processMessage <- msg
    }
}

func processWsMessages() {
    for {
        msg := <-processMessage

        for conn, client := range clients {
            if client.Room == msg.Room {
                err := conn.WriteJSON(msg)
                if err != nil {
                    fmt.Println(err)
                    conn.Close()
                    delete(clients, conn)
                }
            }
        }
    }
}

func index(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "static/chat.html")
}

func main() {
    var port = ":8082"

    http.HandleFunc("/",index)
    http.HandleFunc("/ws",wsConnections)

    go processWsMessages()

    fmt.Println("Started server on: " + port)
    err := http.ListenAndServe(port,nil)

    if err != nil {
        panic("Error starting server: " + err.Error())
    }
}
