package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Morale — Instant {1}{W}{W} (The Dark, #5638):
//
//	"Attacking creatures get +1/+1 until end of turn."
//
// Iroh, Firebending Instructor's pump as a spell: every attacking
// creature, whoever controls it, as it resolves. Outside combat
// nothing is attacking, so it does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8af9d623-efa3-4cbc-a903-aaec32d78b0a",
		Name:         "Morale",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{Match: effects.AttackingCreature(), Power: 1, Toughness: 1, Label: "Morale — attacking creatures get +1/+1"}.Apply(ctx)
		},
	})
}
