package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Merfolk Wayfinder — Creature — Merfolk Scout {2}{U}, 1/2 (Zendikar, S82 OLD
// b494):
//
//	"Flying
//	 When this creature enters, reveal the top three cards of your library. Put all
//	 Island cards revealed this way into your hand and the rest on the bottom of
//	 your library in any order."
//
// Sylvan Messenger's reveal-take-the-rest-to-the-bottom body over Island cards
// (a land type, a word of the card's text, CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "45b34d33-7d70-4308-8f07-bb044be891bb",
		Name:            "Merfolk Wayfinder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Merfolk Wayfinder — reveal the top three cards; Island cards to your hand, the rest on the bottom in any order",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					island := ctx.TextSubtype("Island")
					return effects.S82OldB074RevealTopTakeAllToHandRestBottom(g, item.Controller, ctx.Source(), 3,
						"Merfolk Wayfinder — reveal the top three cards, Island cards to your hand and the rest on the bottom of your library",
						effects.HasSubtype(island))
				}),
		},
	})
}
