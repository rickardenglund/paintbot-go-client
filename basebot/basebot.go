package basebot

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	"paintbot-client/models"
	"paintbot-client/utilities/timeHelper"
)

var mux sync.Mutex

var gm models.GameMode

func init() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:            true,
		ForceQuote:             true,
		FullTimestamp:          true,
		TimestampFormat:        "15:04:05.999",
		DisableLevelTruncation: true,
		PadLevelText:           true,
		QuoteEmptyFields:       true,
	})

	log.SetOutput(os.Stdout)

	log.SetLevel(log.InfoLevel)
}

func Start(playerName string, gameMode models.GameMode, desiredGameSettings *models.GameSettings, calculateMove func(event models.MapUpdateEvent) models.Action) {
	gm = gameMode
	conn := getWebsocketConnection(gameMode)
	defer conn.Close()

	registerPlayer(conn, playerName, desiredGameSettings)

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

	log.Debugf("Received: %s\n", msg)

	gameMSG := models.GameMessage{}
	if err := json.Unmarshal(msg, &gameMSG); err != nil {
		panic(err)
	}

	switch gameMSG.Type {
	case "se.cygni.paintbot.api.exception.InvalidMessage":
		panic("invalid message: " + string(msg))
	case "se.cygni.paintbot.api.response.PlayerRegistered":
		sendClientInfo(conn, gameMSG)
		log.Infof("Player registered")
		go heartbeat(conn, gameMSG.ReceivingPlayerID)
		StartGame(conn)
	case "se.cygni.paintbot.api.event.GameLinkEvent":
		gamelinkEvent := &models.GameLinkEvent{}
		if err := json.Unmarshal(msg, gamelinkEvent); err != nil {
			panic(err)
		}
		log.Infof("Game can be viewed at: %s\n", gamelinkEvent.URL)
	case "se.cygni.paintbot.api.event.GameStartingEvent":
		log.Infof("Game started\n")
	case "se.cygni.paintbot.api.event.MapUpdateEvent":
		updateEvent := models.MapUpdateEvent{}
		if err := json.Unmarshal(msg, &updateEvent); err != nil {
			panic(err)
		}
		if updateEvent.GameTick%10 == 0 {
			log.Infof("Game tick: %d\n", updateEvent.GameTick)
		}
		handleMapUpdate(conn, updateEvent)
	case "se.cygni.paintbot.api.event.GameResultEvent":
		event := models.GameResultEvent{}
		if err := json.Unmarshal(msg, &event); err != nil {
			panic(err)
		}

		log.Infof("### Game Results ###\n")
		for _, player := range event.PlayerRanks {
			log.Infof("%d: %s - %d\n", player.Rank, player.PlayerName, player.Points)
		}
	case "se.cygni.paintbot.api.event.GameEndedEvent":
		event := models.GameEndedEvent{}
		if err := json.Unmarshal(msg, &event); err != nil {
			panic(err)
		}

		if event.PlayerWinnerID == *event.ReceivingPlayerID {
			log.Info("You won the game")
		}
		if gm == models.Training {
			return true
		}
	case "se.cygni.paintbot.api.event.TournamentEndedEvent":
		event := models.TournamentEndedEvent{}
		if err := json.Unmarshal(msg, &event); err != nil {
			panic(err)
		}

		log.Infof("### Tournament Ended ###")
		for _, player := range event.GameResult {
			log.Infof("%s - %d\n", player.Name, player.Points)
		}
		return true
	case "se.cygni.paintbot.api.response.HeartBeatResponse":
	default:
		panic(fmt.Sprintf("unknown message: %s\n", msg))
	}
	return false
}

func registerPlayer(conn *websocket.Conn, playerName string, desiredGameSettings *models.GameSettings) {
	registerMSG := &models.RegisterPlayerEvent{
		Type:              "se.cygni.paintbot.api.request.RegisterPlayer",
		PlayerName:        playerName,
		GameSettings:      desiredGameSettings,
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
