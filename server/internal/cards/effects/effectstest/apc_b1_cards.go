package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// apcB1Push seeds a permanent on the battlefield, catalog row or not,
// untapped and not summoning sick.
func ApcB1Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle, manaCost string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}
