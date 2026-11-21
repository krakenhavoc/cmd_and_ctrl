package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tap_any_number_cost_cards_test.go — #3961: "As an additional cost to
// cast this spell, [you may] tap any number of untapped creatures you
// control" (CR 601.2b / 601.2f / 601.2h). #3868's TapOthers cost with the
// count the caster's: the creatures ride cast_spell's teamwork_ids, none
// included, are validated as #3868's are and tapped with the spell on the
// stack. Burn at the Stake reads the count at resolution; Explosive
// Singularity's discount reads it at CR 601.2f, through the one pricer.

const (
	BurnAtTheStakeOracle       = "eb23bbb0-dd12-44f1-8645-d92f46cbe8ff"
	ExplosiveSingularityOracle = "30d79f61-f9d9-441a-b600-83721af8553b"
)

// taoCreatures puts n untapped creatures onto `owner`'s battlefield.
func TaoCreatures(g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	out := make([]uuid.UUID, n)
	for i := range out {
		out[i] = Pr7Creature(g, owner, "Soldier", 1, "R")
	}
	return out
}
