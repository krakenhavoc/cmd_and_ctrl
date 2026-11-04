package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Splinter's Technique — Sorcery {3}{B} (Teenage Mutant Ninja Turtles,
// #4457):
//
//	"Sneak {1}{B} (You may cast this spell for {1}{B} if you also return an
//	 unblocked attacker you control to hand during the declare blockers
//	 step.)
//	 Search your library for a card, put that card into your hand, then
//	 shuffle."
//
// Sneak is the shared alternative cost (sneak.go, CR 702.190). The search
// is Demonic Tutor's: any card, not revealed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "532ee9c3-391c-44db-8ed5-a1d28fd7f876",
		Name:             "Splinter's Technique",
		Purpose:          game.Purpose{Tutors: 1},
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Sneak("{1}{B}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:  item.Controller,
				Dest:    game.ZoneHand,
				Limit:   1,
				Shuffle: true,
				Reason:  "Splinter's Technique — search your library for a card",
			}.Apply(ctx)
		},
	})
}
