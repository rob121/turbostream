// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package turbostream



import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"strings"
	"sync"
	/*"strings"*/
	"time"
)

var mutex = &sync.RWMutex{}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub
    //unique identifier like a session
	id string
	// The websocket connection.
	conn *websocket.Conn
	//list of channels to subscribe to
	channels []string
	// Buffered channel of outbound messages.
	send chan []byte
}

func (c *Client) SubscribeChannel(id string){

	mutex.Lock()
	c.channels = append(c.channels,id)
	mutex.Unlock()

}

func (c *Client) UnsubscribeChannel(id string){

	var nc []string
	for _,ch := range c.channels {

		if ( ch!=id ){

	      nc = append(nc,ch)
		}

	}

	mutex.Lock()
	c.channels = nc //overwrite
	mutex.Unlock()

}

func (c *Client) HasChannel(channel_id string)(bool){

	 for _,ch := range c.channels {

	 	if(ch==channel_id){
	 		return true
		}

	 }

	 return false

}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := c.conn.ReadMessage() //do something with messages from client?
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Printf("error: %v", err)
			}
			break
		}
		//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		//c.hub.broadcast <- message
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.

type Response struct{
   Identifier ResponseIdentifier `json:"identifier"`
   Message string `json:"message"`
}

type ResponseIdentifier struct {
  Channel string `json:"channel"`
  StreamName string `json:"signed_stream_name"`
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			resp := Response{Identifier: ResponseIdentifier{Channel: "Turbo::StreamsChannel",StreamName: fmt.Sprint(time.Now().Unix()) },Message: string(message)}

			msg_full,err := jsonMarshal(resp)

			if(err!=nil){

				logger.Println(err)
			}

			//remove the new lines and tabs from the response
		    r := strings.NewReplacer("\\n","","\\t","")

		    msg := []byte(r.Replace(string(msg_full)))

			w.Write(msg)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func jsonMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}


// serveWs handles websocket requests from the peer.
func HandleWs(hub *Hub,session_id string, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Println(err)
		return
	}

	client := &Client{hub: hub,id: session_id, conn: conn, send: make(chan []byte, 256)}
	client.SubscribeChannel(defaultChannel)//this is a default channel
	client.hub.register <- client
	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}