package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Final Judgment — Sorcery {4}{W}{W}:
//
//	"Exile all creatures."
//
// A wrath that exiles: every creature on the battlefield, yours
// included, goes to exile, so indestructible and regeneration do
// nothing against it. The set is read once as the spell resolves
// (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "69ef9308-bc25-4f3c-a5fd-1ebaed6e4842",
		Name:         "Final Judgment",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ExileAllMatching{Match: Creature()}.Apply(ctx)
		},
	})
}
