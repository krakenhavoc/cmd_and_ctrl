package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mishra's Onslaught — Instant {3}{R} (The Brothers' War, #5482):
//
//	"Choose one —
//	 • Create two 1/1 colorless Soldier artifact creature tokens.
//	 • Creatures you control get +2/+0 until end of turn."
//
// The pump is a one-shot set locked at resolution (CR 611.2c): a creature
// that arrives later is not pumped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3684edb5-bbf1-4ed7-b5b7-f24b452dbd64",
		Name:         "Mishra's Onslaught",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Create two 1/1 colorless Soldier artifact creature tokens.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.CreateToken{Template: effects.TokenCard(effects.BroB05SoldierToken), N: 2}.Apply(ctx)
				}),
			effects.ModeDoing("Creatures you control get +2/+0 until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Label: "Mishra's Onslaught — +2/+0"}.Apply(ctx)
				}),
		),
	})
}
