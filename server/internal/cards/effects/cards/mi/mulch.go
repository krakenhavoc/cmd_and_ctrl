package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mulch — Sorcery {1}{G} (Duskmourn Commander, #5368):
//
//	"Reveal the top four cards of your library. Put all land cards revealed
//	 this way into your hand and the rest into your graveyard."
//
// "All land cards" is not a choice, so there is no prompt. The rest are put
// into the graveyard, not milled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "12c03a03-36bf-487e-b8be-e55399525da5",
		Name:         "Mulch",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.Cmdr2B15RevealTopTakeAllMatching(ctx.Game, item.Controller, ctx.Source(), 4,
				"Mulch — reveal the top four cards, lands to your hand and the rest to your graveyard", effects.Land())
		},
	})
}
