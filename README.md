# Golang Turbo Stream Package

For use with hotwire turbo, this package allows you to quickly get up and running with responding to websockets with turbo-streams

## Basic Examples

### Attach a standard logger
```
turbostream.Logger(log.New(os.Stdout, "[turbostream]", log.Lshortfile))
```

### Create a message hub
```
	hub = turbostream.NewHub()
	go hub.Run() 
```
### Add to your websocket handler (gorilla in this case)
```
    
    r.HandleFunc("/ws", websocketHandler)
    
    <snip>
    
    func websocketHandler(w http.ResponseWriter, r *http.Request) {
    	turbostream.HandleWs(hub, session_id, w, r)
    }
```

### Broadcast a Message to all clients

```
	hub.Broadcast(turbostream.Message("append","example_div_id",fmt.Sprint(time.Now().Unix(),"<br>")))
```

### Subscribe to a channel

```
  hub.Subscribe(session_id,channel_name)
```

```
  hub.Unsubscribe(session_id,channel_name)
```

### Send a message on a channel
```
 hub.SendChannel(channel_id,message)
```

The session id can be any unique value for the cleint

### Send a message to a specific client

```
    err:= hub.Send(session_id,turbostream.Message("append","example_div_id","Hello!"))
	if(err!=nil){
		fmt.Fprintf(w, err.Error())
	}
```