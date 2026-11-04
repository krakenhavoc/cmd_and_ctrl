package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rapturous Moment — Sorcery {4}{U}{R} (Secrets of Strixhaven, #4456):
//
//	"Draw three cards, then discard two cards. Add {U}{U}{R}{R}{R}."
//
// The draws finish before the discard prompt opens, so a drawn card is a
// legal discard (drawThenDiscard, as Thirst for Knowledge); the caster picks
// which two to pitch, and a smaller hand gives up what it has (CR 701.9a).
// The mana is added as the spell resolves, fixed colours, and empties at the
// end of the phase as any mana does (CR 500.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c997afb7-82cb-4579-885e-3628cb1b51e4",
		Name:         "Rapturous Moment",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := drawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), 3, 2, ""); err != nil {
				return err
			}
			return AddMana{Produced: "{U}{U}{R}{R}{R}"}.Apply(ctx)
		},
	})
}
