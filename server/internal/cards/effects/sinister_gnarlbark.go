package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sinister Gnarlbark — Creature — Treefolk Warlock {2}{B}, 0/4 (Lorwyn
// Eclipsed Commander, #5368):
//
//	"At the beginning of your end step, draw a card and blight 1. (Put a
//	 -1/-1 counter on a creature you control.)"
//
// The draw comes first, then the blight, which is asked as the ability
// resolves (Blight, #4683): a creature you control must take the counter,
// the Gnarlbark itself included, and with none that can the blight is
// skipped (CR 701.68b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fcd2142e-ba84-4693-9b29-ecc43feba2b8",
		Name:         "Sinister Gnarlbark",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Sinister Gnarlbark — draw a card and blight 1", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
					return err
				}
				return Blight{N: 1}.Apply(ctx)
			}),
		},
	})
}
