package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exoB1Push seeds a permanent on the battlefield, catalog row or not,
// untapped and not summoning sick.
func ExoB1Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, colors ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Colors: colors, Owner: owner, Controller: owner,
	})
}
