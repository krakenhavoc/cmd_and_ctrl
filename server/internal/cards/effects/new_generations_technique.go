package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// New Generation's Technique — Sorcery {3}{G} (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"Sneak {2}{G} (You may cast this spell for {2}{G} if you also return an
//	 unblocked attacker you control to hand during the declare blockers
//	 step.)
//	 Search your library for up to two basic land cards, put them onto the
//	 battlefield tapped, then shuffle."
//
// Sneak is the shared alternative cost (sneak.go, CR 702.190): a sorcery
// cast for it is cast in the declare blockers step. The search is Blighted
// Woodland's, without the reveal the text does not print.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "e7ae3ee7-824f-4dcb-8f28-7de99e4d8c3e",
		Name:             "New Generation's Technique",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Sneak("{2}{G}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:        item.Controller,
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         2,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "New Generation's Technique — up to two basic lands, onto the battlefield tapped",
			}.Apply(ctx)
		},
	})
}
