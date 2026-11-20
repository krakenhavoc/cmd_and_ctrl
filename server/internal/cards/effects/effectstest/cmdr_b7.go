package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b7Put seeds a permanent with printed power and toughness (the catalog
// supplies abilities, not P/T) and fires its enter event.
func B7Put(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, t int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: t, Owner: owner, Controller: owner,
	})
}
