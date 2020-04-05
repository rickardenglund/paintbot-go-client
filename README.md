# Paintbot client in Go

This is a Paintbot Client written in Go 1.14. 

For more information about what Paintbot is, see: https://paintbot.cygni.se/

For running your own server see [Paintbot Server Repository](https://github.com/cygni/paintbot)

## Requirements

* Go 1.14
* Paintbot Server (local or remote, there's one running by Cygni so no worries ;) )

## Usage
* Clone this repository
* Execute: run the client
```
> cd <repo>/cmd/examplebot
> go run main.go
```

## Implementation

You only need to implement when function in order to have your own bot up and running. see [ExampleBot](cmd/examplebot/main.go)

``` go
func calculateMove(updateEvent models.MapUpdateEvent) models.Action {
	utility := maputility.MapUtility{Map: updateEvent.Map, CurrentPlayerID: *updateEvent.ReceivingPlayerID}
	me := utility.GetMyCharacterInfo()
	move := models.Stay
    // ...
    // super smart logic
    // ...
	return move
}
```

calculateMove will be called every time a map update is received from the server.
You are expected to reply with a CharacterAction (UP, DOWN, LEFT, RIGHT, STAY or EXPLODE). 
And don't forget to respond within the time limit. default is 250 ms including networking.

### Help
There's a utility class with nifty methods to help you out. Take a look at [Map utility](utilities/maputility/README.md)
