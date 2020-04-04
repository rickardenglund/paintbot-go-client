package maputility

import "paintbot-client/models"

type Action string

const (
	Left    Action = "LEFT"
	Right   Action = "RIGHT"
	Up      Action = "UP"
	Down    Action = "DOWN"
	Stay    Action = "STAY"
	Explode Action = "EXPLODE"
)

type Tile string

const (
	Obstacle Tile = "OBSTACLE"
	PowerUp  Tile = "POWERUP"
	Player   Tile = "PLAYER"
	Open     Tile = "OPEN"
)

type Coordinates struct {
	X int
	Y int
}

// Utility for getting information from the map object in a bit more developer friendly format
type MapUtility struct {
	mapMsg          models.Map
	currentPlayerID string
}

func (u *MapUtility) CanIMoveInDirection(action Action) bool {
	pos := u.GetMyPosition()
	pos = u.TranslateCoordinateByAction(action, pos)
	return u.IsTileAvailableForMovementTo(pos)
}

// Returns the coordinates given after an action has been performed successfully
func (u *MapUtility) TranslateCoordinateByAction(action Action, pos Coordinates) Coordinates {
	switch action {
	case Left:
		return Coordinates{X: pos.X - 1, Y: pos.Y}
	case Right:
		return Coordinates{X: pos.X + 1, Y: pos.Y}
	case Up:
		return Coordinates{X: pos.X, Y: pos.Y - 1}
	case Down:
		return Coordinates{X: pos.X, Y: pos.Y + 1}
	case Stay, Explode:
		return Coordinates{X: pos.X, Y: pos.Y}
	default:
		panic("Unknown Action: " + action)
	}
}

func (u *MapUtility) GetPlayerColouredPositions(playerId string) []Coordinates {
	return u.ConvertPositionsToCoordinates(u.GetCharacterInfo(playerId).ColouredPosition)
}

func (u *MapUtility) ListCoordinatesContainingPowerUps() []Coordinates {
	return u.ConvertPositionsToCoordinates(u.mapMsg.PowerUpPositions)
}

func (u *MapUtility) ListCoordinatesContainingObstacles() []Coordinates {
	return u.ConvertPositionsToCoordinates(u.mapMsg.ObstacleUpPositions)
}

func (u *MapUtility) IsTileAvailableForMovementTo(coord Coordinates) bool {
	tile := u.GetTileAt(coord)

	return tile == Open || tile == PowerUp
}

func (u *MapUtility) GetMyPosition() Coordinates {
	return u.ConvertPositionToCoordinates(u.GetMyCharacterInfo().Position)
}

func (u *MapUtility) GetMyCharacterInfo() models.CharacterInfo {
	return u.GetCharacterInfo(u.currentPlayerID)
}

func (u *MapUtility) GetCharacterInfo(playerID string) models.CharacterInfo {
	for i := range u.mapMsg.CharacterInfos {
		if u.mapMsg.CharacterInfos[i].ID == playerID {
			return u.mapMsg.CharacterInfos[i]
		}
	}
	panic("Trying to find invalid playerID: " + playerID)
}

func (u *MapUtility) IsCoordinatesOutOfBounds(coord Coordinates) bool {
	w := u.mapMsg.Width
	h := u.mapMsg.Height
	return coord.X < 0 || coord.Y < 0 || coord.X > w || coord.Y > h
}

func (u *MapUtility) GetTileAt(coordinates Coordinates) Tile {
	return u.getTileAtPosition(u.ConvertCoordinatesToPosition(coordinates))
}

func (u *MapUtility) getTileAtPosition(position int) Tile {
	if contains(u.mapMsg.ObstacleUpPositions, position) {
		return Obstacle
	}

	if contains(u.mapMsg.PowerUpPositions, position) {
		return PowerUp
	}

	if contains(u.getPlayerPositions(), position) {
		return Player
	}

	return Open
}

func contains(ns []int, n int) bool {
	for i := range ns {
		if ns[i] == n {
			return true
		}
	}
	return false
}

// Converts a position in the flattened single array representation
// of the Map to a Coordinates.
func (u *MapUtility) ConvertPositionToCoordinates(position int) Coordinates {
	w := u.mapMsg.Width
	return Coordinates{
		X: position % w,
		Y: position / w,
	}
}

// Converts a MapCoordinate to the same position in the flattened
// single array representation of the Map.
func (u *MapUtility) ConvertCoordinatesToPosition(coordinates Coordinates) int {
	w := u.mapMsg.Width
	return coordinates.Y*w + coordinates.X
}

func (u *MapUtility) ConvertPositionsToCoordinates(positions []int) []Coordinates {
	coords := make([]Coordinates, len(positions))
	for i := range positions {
		coords[i] = u.ConvertPositionToCoordinates(positions[i])
	}
	return coords
}

func (u *MapUtility) ConvertCoordinatesToPositions(coordinates []Coordinates) []int {
	positions := make([]int, len(coordinates))
	for i := range coordinates {
		positions[i] = u.ConvertCoordinatesToPosition(coordinates[i])
	}
	return positions
}

func (u *MapUtility) getPlayerPositions() []int {
	positions := make([]int, len(u.mapMsg.CharacterInfos))
	for i := range u.mapMsg.CharacterInfos {
		positions[i] = u.mapMsg.CharacterInfos[i].Position
	}
	return positions

}
