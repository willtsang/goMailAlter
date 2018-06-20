package main

import (
	"golang.org/x/net/websocket"
	"fmt"
)

func main() {

	var origin = "http://127.0.0.1:9999/"
	var url = "ws://127.0.0.1:9999/sub"

	for i := 0; i < 4; i++ {
		go sendRequest(url, origin)
	}

	var exit chan bool

	 <-exit
}

func sendRequest(url, origin string) {
	ws, err := websocket.Dial(url, "", origin)

	if err != nil {
		panic(err.Error())
	}

	msg := []byte("hello world!")

	_, err = ws.Write(msg)

	if err != nil {
		panic(err.Error())
	}

	for {
		readMsg := make([]byte, 512)

		m, err := ws.Read(readMsg)

		if err != nil {
			panic(err.Error())
		}

		fmt.Printf("Receive %s\n", readMsg[:m])
	}

	ws.Close()
}
