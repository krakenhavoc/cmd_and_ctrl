package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ransack — {3}{U} Sorcery (Starter 1999, S82 OLD batch 701, #5620):
//
//	"Look at the top five cards of target player's library. Put any number of
//	 them on the bottom of that library in any order and the rest on top of the
//	 library in any order."
//
// LookAtLibraryThenPlace with both lanes open and no top count: every card goes
// to exactly one lane, each in the order the controller chooses.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b4a400b-435d-4140-984b-65f034c6a9fc",
		Name:         "Ransack",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			owner, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			return effects.LookAtLibraryThenPlace{
				Owner:     owner,
				N:         5,
				Placement: game.LibraryPlaceTopOrBottom,
				Label:     "Ransack — put any number on the bottom of that library and the rest on top, in any order",
			}.Apply(ctx)
		},
	})
}
