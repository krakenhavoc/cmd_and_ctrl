package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d40ea3d9-8c48-4846-b1f4-99018d4fd789",
		Name:         "Esper Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Destroy target enchantment.",
				TargetPermanent("target enchantment", Enchantment()),
				DestroyTheModesTarget),
			ModeWithPurpose(ModeDoing("Draw two cards.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
				}), game.Purpose{Draws: 2}),
			ModeDoing("Target player discards two cards.",
				TargetPlayer("target player"),
				TheModesTargetDiscards(2)),
		),
	})
}
