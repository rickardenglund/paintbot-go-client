package main

import (
	"encoding/json"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	"paintbot-client/models"
	"paintbot-client/utilities/maputility"
	"paintbot-client/utilities/timeHelper"
)

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

func main() {
	u := url.URL{
		Scheme: "ws",
		//Host:   "server.paintbot.cygni.se:80",
		Host: "localhost:8080",
		Path: "/training",
	}
	log.Debugf("connecting to: %s\n", u.String())
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
			calculateMove(conn, updateEvent)
		case "se.cygni.paintbot.api.event.GameEndedEvent":
			log.Infof("Game ended: %s\n", msg)
			return true
		}
	}
	return false
}

var moves = []models.Action{models.Explode, models.Left, models.Down, models.Right, models.Up} //, models.Stay}
var lastDir = 0

func calculateMove(conn *websocket.Conn, updateEvent models.MapUpdateEvent) {
	utility := maputility.MapUtility{Map: updateEvent.Map, CurrentPlayerID: *updateEvent.ReceivingPlayerID}
	me := utility.GetMyCharacterInfo()
	move := models.Stay
	if me.StunnedForGameTicks > 0 {
		log.Warn("stunned")
		sendMove(conn, updateEvent, models.Stay)
		return
	}

	if me.CarryingPowerUp {
		log.Warn("bombing")
		sendMove(conn, updateEvent, models.Explode)
		return
	}

	for i := range moves {
		p := (i + lastDir) % len(moves)
		if utility.CanIMoveInDirection(moves[p]) {
			move = moves[p]
			lastDir = p
			break
		}
	}

	sendMove(conn, updateEvent, move)
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

func registerPlayer(conn *websocket.Conn) {
	registerMSG := models.RegisterPlayerEvent{
		Type:       "se.cygni.paintbot.api.request.RegisterPlayer",
		PlayerName: "Simple Go Bot",
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
		LanguageVersion:        "1.14",
		OperatingSystem:        "",
		OperatingSystemVersion: "",
		ClientVersion:          "0.1",
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
