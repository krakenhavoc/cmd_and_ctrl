package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Noxious Revival — Instant {G/P}:
//
//	"({G/P} can be paid with either {G} or 2 life.)
//	 Put target card from a graveyard on top of its owner's library."
//
// Any graveyard, any card type: the target clause names no owner and
// no type. The card goes to the top of its OWNER's library, which is
// what TuckToLibraryForEffect does. The Phyrexian symbol is paid by
// the engine at cast (CR 107.4f), as for Gut Shot.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "97cabeda-9fe3-490d-99b4-4c8d87c17157",
		Name:         "Noxious Revival",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target card from a graveyard"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return PutChosenTargetOnTopOfLibrary(ctx.Game, item)
		},
	})
}
