package models

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

type GameMode string

const (
	Tournament GameMode = "/tournament"
	Training   GameMode = "/training"
)
