package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mgbLand puts an untapped land under owner's control.
func MgbLand(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: owner, Controller: owner,
	})
}
