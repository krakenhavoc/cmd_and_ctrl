package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reclaim — Instant {G}:
//
//	"Put target card from your graveyard on top of your library."
//
// The same instruction as Noxious Revival with a narrower clause:
// only the caster's own graveyard. The target is re-checked at
// resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "797155cd-faf4-4321-8629-c8c352392748",
		Name:         "Reclaim",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target card from your graveyard", YouOwn()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return PutChosenTargetOnTopOfLibrary(ctx.Game, item)
		},
	})
}
