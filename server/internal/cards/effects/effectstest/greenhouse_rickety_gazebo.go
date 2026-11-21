package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const greenhouseOracle = "a341eb75-e6a0-468b-8c83-61102249c648"

func GreenhouseCard(owner uuid.UUID) game.Card {
	return RoomsDCard(owner, greenhouseOracle, "Greenhouse", "{2}{G}", "Rickety Gazebo", "{3}{G}", "G")
}
