package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"paintbot-client/basebot"
	"paintbot-client/models"
	"paintbot-client/utilities/maputility"
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
	basebot.Start("Simple Go Bot", calculateMove)
}

var moves = []models.Action{models.Explode, models.Left, models.Down, models.Right, models.Up} //, models.Stay}
var lastDir = 0

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
