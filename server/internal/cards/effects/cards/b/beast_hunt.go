package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Beast Hunt — Sorcery {3}{G} (Planechase, #5584):
//
//	"Reveal the top three cards of your library. Put all creature cards
//	 revealed this way into your hand and the rest into your graveyard."
//
// Mulch's body over creatures and three cards: "all creature cards" is not
// a choice, so there is no prompt, and the rest are put into the graveyard,
// not milled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "069c8fa1-fa04-4620-9356-a618158dad3e",
		Name:         "Beast Hunt",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.Cmdr2B15RevealTopTakeAllMatching(ctx.Game, item.Controller, ctx.Source(), 3,
				"Beast Hunt — reveal the top three cards, creatures to your hand and the rest to your graveyard", effects.Creature())
		},
	})
}
