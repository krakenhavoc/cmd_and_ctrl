package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Needle Storm — Sorcery {2}{G} (Portal, #5630):
//
//	"Needle Storm deals 4 damage to each creature with flying."
//
// Tremor's sweep, turned round: only creatures with flying are hit.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f513d1e-03ca-4e0f-af48-6490d71d5c41",
		Name:         "Needle Storm",
		Completeness: effects.CompletenessFull,
		// ADR 0126 §6: a creature without flying is spared.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 4, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.HasKeyword("flying")), 4)
		},
	})
}
