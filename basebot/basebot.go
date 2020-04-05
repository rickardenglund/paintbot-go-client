package basebot

import (
	"encoding/json"
	"net/url"
	"runtime"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	"paintbot-client/models"
	"paintbot-client/utilities/timeHelper"
)

var u = url.URL{
	Scheme: "ws",
	//Host:   "server.paintbot.cygni.se:80",
	Host: "localhost:8080",
	Path: "/training",
}

func Start(playerName string, calculateMove func(event models.MapUpdateEvent) models.Action) {
	log.Debugf("connecting to: %s\n", u.String())
	conn, _, connectionError := websocket.DefaultDialer.Dial(u.String(), nil)
	if connectionError != nil {
		panic(connectionError)
	}
	defer conn.Close()

	registerPlayer(conn, playerName)

	handleMapUpdate := func(conn *websocket.Conn, event models.MapUpdateEvent) {
		action := calculateMove(event)
		sendMove(conn, event, action)
	}

	more := true
	for more {
		more = !recv(conn, handleMapUpdate)
	}
}

func recv(conn *websocket.Conn, handleMapUpdate func(*websocket.Conn, models.MapUpdateEvent)) (done bool) {
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
			log.Debug("Player Registered\n")
			sendClientInfo(conn, gameMSG)
			StartGame(conn)
		case "se.cygni.paintbot.api.event.GameLinkEvent", "se.cygni.paintbot.api.event.GameStartingEvent":
			log.Infof("Received: %s\n", msg)
		case "se.cygni.paintbot.api.event.MapUpdateEvent":
			updateEvent := models.MapUpdateEvent{}
			if err := json.Unmarshal(msg, &updateEvent); err != nil {
				panic(err)
			}
			log.Debugf("Map update: %+v\n", updateEvent)
			handleMapUpdate(conn, updateEvent)
		case "se.cygni.paintbot.api.event.GameEndedEvent":
			log.Infof("Game ended: %s\n", msg)
			return true
		}
	}
	return false
}

func registerPlayer(conn *websocket.Conn, playerName string) {
	registerMSG := models.RegisterPlayerEvent{
		Type:       "se.cygni.paintbot.api.request.RegisterPlayer",
		PlayerName: playerName,
		GameSettings: models.GameSettings{
			MaxNOOFPlayers:                 5,
			TimeInMSPerTick:                250,
			ObstaclesEnabled:               true,
			PowerUpsEnabled:                true,
			AddPowerUpLikelihood:           38,
			RemovePowerUpLikelihood:        5,
			TrainingGame:                   true,
			PointsPerTileOwned:             1,
			PointsPerCausedStun:            5,
			NOOFTicksInvulnerableAfterStun: 3,
			NOOFTicksStunned:               10,
			StartObstacles:                 40,
			StartPowerUps:                  41,
			GameDurationInSeconds:          60,
			ExplosionRange:                 4,
			PointsPerTick:                  false,
		},
		ReceivingPlayerID: nil,
		Timestamp:         timeHelper.Now(),
	}

	if err := conn.WriteJSON(registerMSG); err != nil {
		panic(err)
	}
}

func sendClientInfo(conn *websocket.Conn, msg models.GameMessage) {

	if err := conn.WriteJSON(models.ClientInfoMSG{
		Type:                   "se.cygni.paintbot.api.event.GameStartingEvent",
		Language:               "Go",
		LanguageVersion:        runtime.Version(),
		OperatingSystem:        runtime.GOOS,
		OperatingSystemVersion: "",
		ClientVersion:          "0.2",
		ReceivingPlayerID:      msg.ReceivingPlayerID,
		Timestamp:              timeHelper.Now(),
	}); err != nil {
		panic(err)
	}
}

func StartGame(conn *websocket.Conn) {
	startGame := models.StartGameEvent{
		Type:              "se.cygni.paintbot.api.request.StartGame",
		ReceivingPlayerID: nil,
		Timestamp:         timeHelper.Now(),
	}

	if err := conn.WriteJSON(startGame); err != nil {
		panic(err)
	}
}

func sendMove(conn *websocket.Conn, updateEvent models.MapUpdateEvent, action models.Action) {
	moveEvent := models.RegisterMoveEvent{
		Type:              "se.cygni.paintbot.api.request.RegisterMove",
		GameID:            updateEvent.GameID,
		GameTick:          updateEvent.GameTick,
		Action:            string(action),
		ReceivingPlayerID: updateEvent.ReceivingPlayerID,
		Timestamp:         timeHelper.Now(),
	}
	if marshal, err := json.Marshal(moveEvent); err != nil {
		panic(err)
	} else {
		log.Debugf("send action: %s\n", marshal)
	}

	if err := conn.WriteJSON(moveEvent); err != nil {
		panic(err)
	}
}
