package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "3684edb5-bbf1-4ed7-b5b7-f24b452dbd64",
		Name:         "Mishra's Onslaught",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Create two 1/1 colorless Soldier artifact creature tokens.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return CreateToken{Template: TokenCard(broB05SoldierToken), N: 2}.Apply(ctx)
				}),
			ModeDoing("Creatures you control get +2/+0 until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 2, Label: "Mishra's Onslaught — +2/+0"}.Apply(ctx)
				}),
		),
	})
}
