package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Natural Balance — Sorcery {2}{G}{G} (Mirage, S82 OLD b798, #5633):
//
//	"Each player who controls six or more lands chooses five lands they
//	 control and sacrifices the rest. Each player who controls four or fewer
//	 lands may search their library for up to X basic land cards and put them
//	 onto the battlefield, where X is five minus the number of lands they
//	 control. Then each player who searched their library this way shuffles."
//
// Cataclysm's keep-and-sacrifice (each player asked once in turn order, the
// rest sacrificed at once) with a fixed keep of five, then each qualifying
// player's own optional search for the basic lands they are short of, put onto
// the battlefield untapped; a player who searches shuffles their own library.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69acaa6f-3ba2-420c-9069-02958f844121",
		Name:         "Natural Balance",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 5},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.OldB798NaturalBalance(ctx.Game, item)
		},
	})
}
