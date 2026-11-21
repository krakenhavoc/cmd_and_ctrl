package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Salvage — Sorcery {G} (Portal Second Age, #5625):
//
//	"Put target card from your graveyard on top of your library."
//
// Reclaim at sorcery speed: only the caster's own graveyard, and the
// target is checked again at resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c01a090a-11e4-469e-aded-74203c06fef4",
		Name:         "Salvage",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target card from your graveyard", effects.YouOwn()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PutChosenTargetOnTopOfLibrary(ctx.Game, item)
		},
	})
}
