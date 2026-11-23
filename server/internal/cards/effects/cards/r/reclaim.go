package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "797155cd-faf4-4321-8629-c8c352392748",
		Name:         "Reclaim",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target card from your graveyard", effects.YouOwn()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PutChosenTargetOnTopOfLibrary(ctx.Game, item)
		},
	})
}
