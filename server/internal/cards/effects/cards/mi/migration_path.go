package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Migration Path — Sorcery {3}{G} (Tarkir: Dragonstorm Commander):
//
//	"Search your library for up to two basic land cards, put them onto
//	 the battlefield tapped, then shuffle.
//	 Cycling {2}"
//
// Explosive Vegetation's body (cmdrB21UpToTwoBasicsTapped) plus Cycling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "db4907f3-4071-47c4-9329-96b4813435be",
		Name:         "Migration Path",
		Purpose:      game.Purpose{Lands: 2},
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.CmdrB21UpToTwoBasicsTapped(ctx, "Migration Path")
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
