package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"paintbot-client/basebot"
	"paintbot-client/models"
	"paintbot-client/utilities/maputility"
)

func main() {
	basebot.Start("Simple Go Bot", models.Training, desiredGameSettings, calculateMove)
}

var moves = []models.Action{models.Explode, models.Left, models.Down, models.Right, models.Up} //, models.Stay}
var lastDir = 0

// Implement your paintbot here
func calculateMove(updateEvent models.MapUpdateEvent) models.Action {
	utility := maputility.MapUtility{Map: updateEvent.Map, CurrentPlayerID: *updateEvent.ReceivingPlayerID}
	me := utility.GetMyCharacterInfo()
	move := models.Stay
	if me.StunnedForGameTicks > 0 {
		return models.Stay
	}

	if me.CarryingPowerUp {
		return models.Explode
	}

	for i := range moves {
		p := (i + lastDir) % len(moves)
		if utility.CanIMoveInDirection(moves[p]) {
			move = moves[p]
			lastDir = p
			break
		}
	}
	return move
}

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

// desired game settings can be changed to nil to get default settings
var desiredGameSettings = &models.GameSettings{
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
}
