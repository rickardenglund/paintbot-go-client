package basebot

import (
	"net/url"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	"paintbot-client/models"
)

func getWebsocketConnection(gameMode models.GameMode) *websocket.Conn {
	var u = url.URL{
		Scheme: "ws",
		Host:   "server.paintbot.cygni.se:80",
		//Host: "localhost:8080",
		Path: string(gameMode),
	}

	log.Debugf("connecting to: %s\n", u.String())
	conn, _, connectionError := websocket.DefaultDialer.Dial(u.String(), nil)
	if connectionError != nil {
		panic(connectionError)
	}
	return conn
}

func send(conn *websocket.Conn, msg interface{}) {
	mux.Lock()
	if err := conn.WriteJSON(msg); err != nil {
		panic(err)
	}
	mux.Unlock()
}
