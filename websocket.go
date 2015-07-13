package main

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type wsHandler struct {
	rw   http.ResponseWriter
	r    *http.Request
	conn *websocket.Conn
}

func (ws *wsHandler) open() error {
	var err error
	ws.conn, err = upgrader.Upgrade(ws.rw, ws.r, nil)
	return err
}

func (ws *wsHandler) readJSON() (interface{}, error) {
	var data interface{}
	err := ws.conn.ReadJSON(&data)
	return data, err
}

func (ws *wsHandler) writeError(err error, code int) {
	ws.conn.WriteJSON(map[string]string{"error": err.Error()})
}

func (ws *wsHandler) writeJSON(data interface{}) {
	ws.conn.WriteJSON(data)
}
