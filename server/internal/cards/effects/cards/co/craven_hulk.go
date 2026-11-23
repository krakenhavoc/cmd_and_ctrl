package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Craven Hulk — Creature — Giant Coward {3}{R}, 4/4 (Kaldheim, #5505):
//
//	"This creature can't block alone."
//
// Mogg Flunkies' block rule (CantBlockAlone, CR 509.1b): a declaration
// in which it is its controller's only blocker is illegal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3cd13364-3ee8-4a00-986a-489b0ccf90c7",
		Name:         "Craven Hulk",
		Completeness: effects.CompletenessFull,
		BlockRules:   []game.BlockRule{effects.CantBlockAlone()},
	})
}
