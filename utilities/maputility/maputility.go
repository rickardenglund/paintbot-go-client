package maputility

import "paintbot-client/models"

// Utility for getting information from the map object in a bit more developer friendly format
type MapUtility struct {
	Map             models.Map
	CurrentPlayerID string
}

// returns true if the current player can perform the given action given no action for all other players
func (u *MapUtility) CanIMoveInDirection(action models.Action) bool {
	info := u.GetMyCharacterInfo()

	if info.StunnedForGameTicks > 0 {
		return false
	}

	if action == models.Explode {
		return info.CarryingPowerUp
	}

	if action == models.Stay {
		return true
	}

	pos := u.GetMyCoordinates()
	pos = u.TranslateCoordinateByAction(action, pos)
	return u.IsTileAvailableForMovementTo(pos)
}

// Returns the coordinates given after an action has been performed successfully
func (u *MapUtility) TranslateCoordinateByAction(action models.Action, pos models.Coordinates) models.Coordinates {
	switch action {
	case models.Left:
		return models.Coordinates{X: pos.X - 1, Y: pos.Y}
	case models.Right:
		return models.Coordinates{X: pos.X + 1, Y: pos.Y}
	case models.Up:
		return models.Coordinates{X: pos.X, Y: pos.Y - 1}
	case models.Down:
		return models.Coordinates{X: pos.X, Y: pos.Y + 1}
	case models.Stay, models.Explode:
		return models.Coordinates{X: pos.X, Y: pos.Y}
	default:
		panic("Unknown Action: " + action)
	}
}

func (u *MapUtility) GetPlayerColouredPositions(playerId string) []models.Coordinates {
	return u.ConvertPositionsToCoordinates(u.GetCharacterInfo(playerId).ColouredPosition)
}

func (u *MapUtility) ListCoordinatesContainingPowerUps() []models.Coordinates {
	return u.ConvertPositionsToCoordinates(u.Map.PowerUpPositions)
}

func (u *MapUtility) ListCoordinatesContainingObstacles() []models.Coordinates {
	return u.ConvertPositionsToCoordinates(u.Map.ObstacleUpPositions)
}

// returns true if tile is walkable
func (u *MapUtility) IsTileAvailableForMovementTo(coord models.Coordinates) bool {
	tile := u.GetTileAt(coord)

	return tile == models.Open || tile == models.PowerUp || tile == models.Player
}

// returns the coordinates of the current player
func (u *MapUtility) GetMyCoordinates() models.Coordinates {
	return u.ConvertPositionToCoordinates(u.GetMyCharacterInfo().Position)
}

// returns information about the current player
func (u *MapUtility) GetMyCharacterInfo() models.CharacterInfo {
	return u.GetCharacterInfo(u.CurrentPlayerID)
}

// returns information about the given player
func (u *MapUtility) GetCharacterInfo(playerID string) models.CharacterInfo {
	for i := range u.Map.CharacterInfos {
		if u.Map.CharacterInfos[i].ID == playerID {
			return u.Map.CharacterInfos[i]
		}
	}
	panic("Trying to find invalid playerID: " + playerID)
}

// Returns true if the coordinate is withing the game field
func (u *MapUtility) IsCoordinatesOutOfBounds(coord models.Coordinates) bool {
	w := u.Map.Width
	h := u.Map.Height
	return coord.X < 0 || coord.Y < 0 || coord.X >= w || coord.Y >= h
}

// returns the type of object at the given coordinates
// returns OBSTACLE if the coordinate is out of bounds
func (u *MapUtility) GetTileAt(coordinates models.Coordinates) models.Tile {
	if u.IsCoordinatesOutOfBounds(coordinates) {
		return models.Obstacle
	}

	return u.getTileAtPosition(u.ConvertCoordinatesToPosition(coordinates))
}

func (u *MapUtility) getTileAtPosition(position int) models.Tile {
	if contains(u.Map.ObstacleUpPositions, position) {
		return models.Obstacle
	}

	if contains(u.Map.PowerUpPositions, position) {
		return models.PowerUp
	}

	if contains(u.getPlayerPositions(), position) {
		return models.Player
	}

	return models.Open
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
func (u *MapUtility) ConvertPositionToCoordinates(position int) models.Coordinates {
	w := u.Map.Width
	return models.Coordinates{
		X: position % w,
		Y: position / w,
	}
}

// Converts a MapCoordinate to the same position in the flattened
// single array representation of the Map.
func (u *MapUtility) ConvertCoordinatesToPosition(coordinates models.Coordinates) int {
	w := u.Map.Width
	return coordinates.Y*w + coordinates.X
}

// converts a list of positions to coordinates
func (u *MapUtility) ConvertPositionsToCoordinates(positions []int) []models.Coordinates {
	coords := make([]models.Coordinates, len(positions))
	for i := range positions {
		coords[i] = u.ConvertPositionToCoordinates(positions[i])
	}
	return coords
}

// converts a list of coordinates to positions
func (u *MapUtility) ConvertCoordinatesToPositions(coordinates []models.Coordinates) []int {
	positions := make([]int, len(coordinates))
	for i := range coordinates {
		positions[i] = u.ConvertCoordinatesToPosition(coordinates[i])
	}
	return positions
}

func (u *MapUtility) getPlayerPositions() []int {
	positions := make([]int, len(u.Map.CharacterInfos))
	for i := range u.Map.CharacterInfos {
		positions[i] = u.Map.CharacterInfos[i].Position
	}
	return positions
}
