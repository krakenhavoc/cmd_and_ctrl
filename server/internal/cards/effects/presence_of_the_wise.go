package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Presence of the Wise — Sorcery {2}{W}{W} (Saviors of Kamigawa, #5601):
//
//	"You gain 2 life for each card in your hand."
//
// Gerrard's Wisdom's text: the hand is counted as the spell resolves
// (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6728064e-1d4e-4d4f-a0ab-127c23478f3b",
		Name:         "Presence of the Wise",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return GainLife{Player: ctx.Controller(), Amount: 2 * sokB1CardsInYourHand(ctx)}.Apply(ctx)
		},
	})
}
