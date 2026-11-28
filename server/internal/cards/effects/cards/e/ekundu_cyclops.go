package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ekundu Cyclops — Creature — Cyclops {3}{R}, 3/4 (Mirage, S82 OLD batch 810,
// #5633):
//
//	"If a creature you control attacks, this creature also attacks if able."
//
// The #6197 seam: a CR 508.1d requirement that exists only in a declaration in
// which another creature its controller controls is attacking
// (AlsoAttacksIfACreatureYouControlAttacks).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c89e4e6-c4d7-416c-b985-41739c6f5b50",
		Name:         "Ekundu Cyclops",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AlsoAttacksIfACreatureYouControlAttacks()},
	})
}
