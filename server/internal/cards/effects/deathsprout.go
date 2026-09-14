package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deathsprout — Instant {1}{B}{B}{G} (EDHREC rank 3637):
//
//	"Destroy target creature. Search your library for a basic land
//	 card, put it onto the battlefield tapped, then shuffle."
//
// Removal that ramps. The destroy is the single-target verb, so an
// indestructible creature survives it; the search is not conditional
// on the destroy and runs whether the creature died or survived. The
// spell has one target, so a creature that left in response fizzles
// the whole spell — no land either (CR 608.2b), as printed. Basic is
// read off the supertype, so a Snow-Covered basic is offered; the
// searcher picks which.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "793b0c73-601c-41dc-b49a-45fee4970d52",
		Name:         "Deathsprout",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b34DestroyChosenThenSearchBasicTapped(ctx)
		},
	})
}
