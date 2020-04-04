package maputility

import (
	"testing"

	"paintbot-client/models"
)

func TestMapUtility_convertPositionToCoordinates(t *testing.T) {
	mu := MapUtility{
		mapMsg: models.Map{Width: 5},
	}

	c := mu.ConvertPositionToCoordinates(0)
	if c.X != 0 || c.Y != 0 {
		t.Fail()
	}

	c = mu.ConvertPositionToCoordinates(5)
	if c.X != 0 || c.Y != 1 {
		t.Fail()
	}

	c = mu.ConvertPositionToCoordinates(6)
	if c.X != 1 || c.Y != 1 {
		t.Fail()
	}
}

func TestMapUtility_convertCoordinatesToPosition(t *testing.T) {
	mu := MapUtility{
		mapMsg: models.Map{Width: 5},
	}
	pos := mu.ConvertCoordinatesToPosition(Coordinates{0, 0})
	if pos != 0 {
		t.Fail()
	}

	pos = mu.ConvertCoordinatesToPosition(Coordinates{1, 1})
	if pos != 6 {
		t.Fail()
	}
}

func TestMapUtility_coordOutOfBounds(t *testing.T) {
	mu := MapUtility{
		mapMsg: models.Map{Width: 5, Height: 5},
	}

	if mu.IsCoordinatesOutOfBounds(Coordinates{X: 1, Y: 1}) {
		t.Fail()
	}

	if !mu.IsCoordinatesOutOfBounds(Coordinates{X: -1, Y: 1}) {
		t.Fail()
	}

	if !mu.IsCoordinatesOutOfBounds(Coordinates{X: 6, Y: 1}) {
		t.Fail()
	}
}
