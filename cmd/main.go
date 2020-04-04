package main

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/gorilla/websocket"

	"paintbot-client/models"
	"paintbot-client/utilities/time"
)

func main() {
	u := url.URL{
		Scheme: "ws",
		Host:   "server.paintbot.cygni.se:80",
		Path:   "/training",
	}
	fmt.Printf("connecting to: %s\n", u.String())
	conn, _, connectionError := websocket.DefaultDialer.Dial(u.String(), nil)
	if connectionError != nil {
		panic(connectionError)
	}
	defer conn.Close()

	registerPlayer(conn)

	more := true
	for more {
		more = !recv(conn)
	}
}

func recv(conn *websocket.Conn) (done bool) {
	if _, msg, err := conn.ReadMessage(); err != nil {
		panic(err)
	} else {
		gameMSG := models.GameMessage{}
		if err := json.Unmarshal(msg, &gameMSG); err != nil {
			panic(err)
		}

		switch gameMSG.Type {
		case "se.cygni.paintbot.api.exception.InvalidMessage":
			panic("invalid message: " + string(msg))
		case "se.cygni.paintbot.api.response.PlayerRegistered":
			fmt.Printf("Player Registered\n")
			sendClientInfo(conn, gameMSG)
			StartGame(conn)
		case "se.cygni.paintbot.api.event.GameLinkEvent", "se.cygni.paintbot.api.event.GameStartingEvent":
			fmt.Printf("Received: %s\n", msg)
		case "se.cygni.paintbot.api.event.MapUpdateEvent":
			updateEvent := models.MapUpdateEvent{}
			if err := json.Unmarshal(msg, &updateEvent); err != nil {
				panic(err)
			}
			fmt.Printf("Map update: %+v\n", updateEvent)
			calculateMove(conn, updateEvent)
		case "se.cygni.paintbot.api.event.GameEndedEvent":
			fmt.Printf("Game ended: %s\n", msg)
			return true
		}
	}
	return false
}

func calculateMove(conn *websocket.Conn, updateEvent models.MapUpdateEvent) {
	move := models.RegisterMoveEvent{
		Type:              "se.cygni.paintbot.api.request.RegisterMove",
		GameID:            updateEvent.GameID,
		GameTick:          updateEvent.GameTick,
		Direction:         "LEFT",
		ReceivingPlayerID: updateEvent.ReceivingPlayerID,
		Timestamp:         time.Now(),
	}

	sendMove(move, conn)
}

func sendMove(move models.RegisterMoveEvent, conn *websocket.Conn) {
	if marshal, err := json.Marshal(move); err != nil {
		panic(err)
	} else {
		fmt.Printf("send move: %s\n", marshal)
	}

	if err := conn.WriteJSON(move); err != nil {
		panic(err)
	}
}

func registerPlayer(conn *websocket.Conn) {
	registerMSG := models.RegisterPlayerEvent{
		Type:       "se.cygni.paintbot.api.request.RegisterPlayer",
		PlayerName: "Simple Go Bot",
		GameSettings: models.GameSettings{
			MaxNOOFPlayers:                 5,
			TimeInMSPerTick:                250,
			ObstaclesEnabled:               true,
			PowerUpsEnabled:                true,
			AddPowerUpLikelihood:           15,
			RemovePowerUpLikelihood:        5,
			TrainingGame:                   true,
			PointsPerTileOwned:             1,
			PointsPerCausedStun:            5,
			NOOFTicksInvulnerableAfterStun: 3,
			NOOFTicksStunned:               10,
			StartObstacles:                 5,
			StartPowerUps:                  0,
			GameDurationInSeconds:          2,
			ExplosionRange:                 4,
			PointsPerTick:                  false,
		},
		ReceivingPlayerID: nil,
		Timestamp:         time.Now(),
	}

	if err := conn.WriteJSON(registerMSG); err != nil {
		panic(err)
	}
}

func sendClientInfo(conn *websocket.Conn, msg models.GameMessage) {

	if err := conn.WriteJSON(models.ClientInfoMSG{
		Type:                   "se.cygni.paintbot.api.event.GameStartingEvent",
		Language:               "Go",
		LanguageVersion:        "1.14",
		OperatingSystem:        "",
		OperatingSystemVersion: "",
		ClientVersion:          "0.1",
		ReceivingPlayerID:      msg.ReceivingPlayerID,
		Timestamp:              time.Now(),
	}); err != nil {
		panic(err)
	}
}

func StartGame(conn *websocket.Conn) {
	startGame := models.StartGameEvent{
		Type:              "se.cygni.paintbot.api.request.StartGame",
		ReceivingPlayerID: nil,
		Timestamp:         time.Now(),
	}

	if err := conn.WriteJSON(startGame); err != nil {
		panic(err)
	}
}
