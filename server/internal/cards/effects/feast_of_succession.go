package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feast of Succession — Sorcery {4}{B}{B} (Commander Masters, #5948):
//
//	"All creatures get -4/-4 until end of turn. You become the monarch."
//
// Drown in Sorrow's shrink (BoostUntilEOT over every creature), then the
// crown (BecomeTheMonarch). The crown moves whether or not any creature
// died: the two sentences are independent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c3601f22-7b22-4116-8633-ba67841524bd",
		Name:         "Feast of Succession",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{
			Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 4},
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (BoostUntilEOT{Match: Creature(), Power: -4, Toughness: -4, Label: "Feast of Succession — -4/-4"}).Apply(ctx); err != nil {
				return err
			}
			return BecomeTheMonarch{}.Apply(ctx)
		},
	})
}
