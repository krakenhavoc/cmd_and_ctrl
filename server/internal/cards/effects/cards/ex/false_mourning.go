package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// False Mourning — Sorcery {G} (Portal Three Kingdoms, S82 OLD b714, #5622):
//
//	"Put target card from your graveyard on top of your library."
//
// Salvage and Reclaim's body: only the caster's own graveyard, and the
// target is checked again at resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "31a331a0-3065-4ecb-9e6a-ab11e241833b",
		Name:         "False Mourning",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target card from your graveyard", effects.YouOwn()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PutChosenTargetOnTopOfLibrary(ctx.Game, item)
		},
	})
}
