package turbostream

import (
	"errors"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Send individual message.
	message chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}



func NewHub() *Hub {
	return &Hub{
		message:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Broadcast(msg []byte)  {
	for client := range h.clients {
		select {
		case client.send <- msg:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

func (h *Hub) Clients() (map[*Client]bool)  {
	return h.clients
}


func (h *Hub) Subscribe(session_id string,channel_id string) (error) {


	for client := range h.clients {

		if(client.id == session_id){

			client.SubscribeChannel(channel_id)
			return nil

		}
	}


	return errors.New("NOT FOUND")

}

func (h *Hub) Unsubscribe(session_id string,channel_id string) (error) {


	for client := range h.clients {

		if(client.id == session_id){

			client.SubscribeChannel(channel_id)
			return nil

		}
	}

	return errors.New("NOT FOUND")

}


func (h *Hub) SendChannel(channel_id string,message []byte) (error) {

	for client := range h.clients {

			if (client.HasChannel(channel_id)) {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}

	}

	return nil

}

/*
Sends to specific client id
*/
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
		}
	}
}