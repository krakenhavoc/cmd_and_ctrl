package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Battlewing Mystic — Creature — Bird Wizard {1}{U}, 2/1 (Dominaria
// United, #5487):
//
//	"Kicker {R} (You may pay an additional {R} as you cast this spell.)
//	 Flying
//	 When this creature enters, if it was kicked, discard your hand, then
//	 draw two cards."
//
// The kicked check is Nullpriest of Oblivion's intervening "if" (CR 603.4),
// read off the record the resolution carried onto the permanent, so an
// unkicked or reanimated Mystic puts nothing on the stack. The body is
// Case of the Crimson Pulse's discard-the-hand-then-draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "49f0b317-6b06-482e-ab64-e02713e6c7e8",
		Name:            "Battlewing Mystic",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OptionalCosts:   []game.AdditionalCost{Kicker("{R}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, ThisKickedAtLeast(1)),
				"Battlewing Mystic — kicked: discard your hand, then draw two cards",
				Do(
					Instruction(func(ctx *Context) error {
						_, err := discardWholeHand(ctx.Game, ctx.Controller())
						return err
					}),
					DrawCards{N: 2},
				)),
		},
	})
}
