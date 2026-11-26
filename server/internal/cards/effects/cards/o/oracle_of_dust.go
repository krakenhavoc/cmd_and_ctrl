package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oracle of Dust — Creature — Eldrazi Processor {4}{U}, 3/5 (Battle for
// Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 {2}, Put a card an opponent owns from exile into that player's graveyard:
//	 Draw a card, then discard a card."
//
// The cost is Void Attendant's processing (the activator names the card; with
// none the ability can't be activated, CR 118.3); the effect is Bag of
// Holding's loot.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6e5f50b2-0acc-42be-b9fd-32ea390c07fc",
		Name:            "Oracle of Dust",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Put a card an opponent owns from exile into that player's graveyard: Draw a card, then discard a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.ProcessCards(1, "a card an opponent owns from exile", nil)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.LootOne(g, item, 1)
			},
		}},
	})
}
