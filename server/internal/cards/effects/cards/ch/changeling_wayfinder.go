package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Changeling Wayfinder — Creature — Shapeshifter {3}, 1/2 (Lorwyn
// Eclipsed, #4458):
//
//	"Changeling (This card is every creature type.)
//	 When this creature enters, you may search your library for a basic
//	 land card, reveal it, put it into your hand, then shuffle."
//
// Changeling is printed and honoured in every zone (Irregular Cohort).
// The search is Mycosynth Wellspring's: "you may" is the search prompt's
// decline, and the card found is revealed and goes to the hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "93409d98-93c8-4742-a635-2a4408e66c32",
		Name:            "Changeling Wayfinder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Changeling Wayfinder — search for a basic land card", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Changeling Wayfinder — a basic land card",
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
