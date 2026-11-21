package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Esper Charm — Instant {W}{U}{B} (Shards of Alara, #5588):
//
//	"Choose one —
//	 • Destroy target enchantment.
//	 • Draw two cards.
//	 • Target player discards two cards."
//
// A modal spell (CR 700.2): the mode is chosen as it is cast, and a
// targeted mode whose target has gone does nothing (CR 608.2b). The
// discard is the target player's own choice of two cards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d40ea3d9-8c48-4846-b1f4-99018d4fd789",
		Name:         "Esper Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Destroy target enchantment.",
				effects.TargetPermanent("target enchantment", effects.Enchantment()),
				effects.DestroyTheModesTarget),
			effects.ModeWithPurpose(effects.ModeDoing("Draw two cards.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
				}), game.Purpose{Draws: 2}),
			effects.ModeDoing("Target player discards two cards.",
				effects.TargetPlayer("target player"),
				effects.TheModesTargetDiscards(2)),
		),
	})
}
