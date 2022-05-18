package turbostream

import (
	"errors"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Send individual message.
	message chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		message:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Broadcast(msg []byte)  {
	h.broadcast <- msg
}

func (h *Hub) Send(client_id string,message []byte) (error) {

	client,err := h.ClientFetch(client_id)

	if(err!=nil){
		return err
	}

	select {
	case client.send <- message:
	default:
		close(client.send)
		delete(h.clients, client)
		return errors.New("CLOSED_CONN")
	}

	return nil
}

func (h *Hub) ClientFetch(id string) (*Client,error) {

	 for c,_ := range h.clients  {
	 	if(c.id == id){
	 		return c,nil
		}
	 }

	 return &Client{},errors.New("NOT_FOUND")
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}