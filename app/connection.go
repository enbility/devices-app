package app

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait        = 10 * time.Second
	maxMessageSize   = 8192
	pongWait         = 60 * time.Second
	pingPeriod       = (pongWait * 9) / 10
	closeGracePeriod = 15 * time.Second
)

type Connection struct {
	cem *Cem

	conn *websocket.Conn

	sendChannel  chan []byte
	closeChannel chan struct{}
}

func NewConnection(cem *Cem, ws *websocket.Conn) *Connection {
	conn := &Connection{
		cem:          cem,
		conn:         ws,
		sendChannel:  make(chan []byte, 1),
		closeChannel: make(chan struct{}, 1),
	}

	go conn.readPump(ws)
	go conn.writePump(ws)

	return conn
}

func (c *Connection) sendMessage(msg Message) {
	message, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error json marshal:", err)
		return
	}

	c.sendChannel <- message
}

func (c *Connection) readPump(ws *websocket.Conn) {
	defer func() {
		c.cem.RemoveConnection(c)
		c.conn.Close()
	}()

	_ = ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { _ = ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			break
		}
		log.Println("Received", string(message))

		c.cem.handleMessage(c, message)
	}
}

func (c *Connection) writePump(ws *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case <-c.closeChannel:
			return
		case message, ok := <-c.sendChannel:
			_ = ws.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := ws.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println("Error sending:", err)
				ws.Close()
				return
			}

			log.Println("Sent: ", string(message))
		case <-ticker.C:
			_ = ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println("Error sending:", err)
				return
			}
		}
	}
}
