package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const RatatwotwoOracle = "0edd7d9a-af37-4d3b-8b4b-913c36dadf82"

// pushRatatwotwo seeds the real 2/2 on the battlefield.
func PushRatatwotwo(g *game.Game, owner uuid.UUID) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ratatwotwo", OracleID: RatatwotwoOracle,
		TypeLine: "Legendary Creature — Rat", Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}
