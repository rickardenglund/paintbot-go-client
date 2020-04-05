package basebot

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sync"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	"paintbot-client/models"
	"paintbot-client/utilities/timeHelper"
)

var mux sync.Mutex

var gm models.GameMode

func Start(playerName string, gameMode models.GameMode, calculateMove func(event models.MapUpdateEvent) models.Action) {
	gm = gameMode
	conn := getWebsocketConnection(gameMode)
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
	var msg []byte
	var err error
	if _, msg, err = conn.ReadMessage(); err != nil {
		panic(err)
	}

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
		go heartbeat(conn, gameMSG.ReceivingPlayerID)
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
		if gm == models.Training {
			return true
		}
	case "se.cygni.paintbot.api.event.GameResultEvent":
		log.Infof("Game result: %s", msg)
	case "se.cygni.paintbot.api.event.TournamentEndedEvent":
		log.Infof("Tournament result: %s", msg)
		return true
	case "se.cygni.paintbot.api.response.HeartBeatResponse":
		log.Debug("Heatbeat response")
	default:
		panic(fmt.Sprintf("unknown message: %s\n", msg))
	}
	return false
}

func registerPlayer(conn *websocket.Conn, playerName string) {
	registerMSG := &models.RegisterPlayerEvent{
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
			GameDurationInSeconds:          15,
			ExplosionRange:                 4,
			PointsPerTick:                  false,
		},
		ReceivingPlayerID: nil,
		Timestamp:         timeHelper.Now(),
	}

	log.Debugf("Registering player: %v\n", registerMSG)
	send(conn, registerMSG)
}

func sendClientInfo(conn *websocket.Conn, msg models.GameMessage) {
	clientInfoMSG := &models.ClientInfoMSG{
		Type:                   "se.cygni.paintbot.api.event.GameStartingEvent",
		Language:               "Go",
		LanguageVersion:        runtime.Version(),
		OperatingSystem:        runtime.GOOS,
		OperatingSystemVersion: "",
		ClientVersion:          "0.3",
		ReceivingPlayerID:      msg.ReceivingPlayerID,
		Timestamp:              timeHelper.Now(),
	}
	send(conn, clientInfoMSG)
}

func StartGame(conn *websocket.Conn) {
	startGame := &models.StartGameEvent{
		Type:              "se.cygni.paintbot.api.request.StartGame",
		ReceivingPlayerID: nil,
		Timestamp:         timeHelper.Now(),
	}

	send(conn, startGame)
}

func sendMove(conn *websocket.Conn, updateEvent models.MapUpdateEvent, action models.Action) {
	moveEvent := &models.RegisterMoveEvent{
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

	send(conn, moveEvent)
}
