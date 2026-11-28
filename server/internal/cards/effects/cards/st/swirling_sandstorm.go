package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Swirling Sandstorm — Sorcery {3}{R} (Judgment, #5611):
//
//	"Threshold — Swirling Sandstorm deals 5 damage to each creature without
//	 flying if there are seven or more cards in your graveyard."
//
// Rolling Temblor's sweep at five damage, done only when threshold holds as
// the spell resolves (the spell itself is still on the stack, so it is not
// one of the seven cards).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dc5cb229-82c9-4a87-9a10-490350cbdfc0",
		Name:         "Swirling Sandstorm",
		Completeness: effects.CompletenessFull,
		// ADR 0126 §6: a creature with flying is spared.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 5, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if !effects.FraThresholdMet(ctx.Game, ctx.Controller()) {
				return nil
			}
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.WithoutKeyword("flying")), 5)
		},
	})
}
