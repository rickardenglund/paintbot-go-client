package main

import (
	"fmt"

	"golang.org/x/net/websocket"

	"paintbot-client/models"
)
const trainingURL = "ws://server.paintbot.cygni.se:80/training"
const origin = "http://localhost/"

func main() {
	conn, err := websocket.Dial(trainingURL, "", origin)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if err = websocket.JSON.Send(conn, models.RegisterPlayerEvent{
		Type:              "se.cygni.paintbot.api.request.RegisterPlayer",
		PlayerName:        "simple Go bot",
		GameSettings:      models.GameSettings{
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
		ReceivingPlayerID: "",
		Timestamp:         0,
	}); err != nil {
		panic(err)
	}


	playerRegistered := models.PlayerRegisteredEvent{}
	if err := websocket.JSON.Receive(conn, &playerRegistered); err != nil {
		panic(err)
	}
	fmt.Printf("registered: %v\n", playerRegistered)

}
