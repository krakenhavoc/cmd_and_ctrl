package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dreamstealer — {2}{B} Creature — Human Wizard 1/2 (Hour of Devastation):
//
//	"Menace
//	 Whenever this creature deals combat damage to a player, that player
//	 discards that many cards.
//	 Eternalize {4}{B}{B}"
//
// "That many" is the combat damage dealt, as the trigger recorded it; a player
// with fewer cards discards what they have.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bf7201e6-1b3f-40f1-9b79-44921b00b0de",
		Name:            "Dreamstealer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Dreamstealer — that player discards that many cards",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					victim, n := item.Trigger.Event.Target, item.Trigger.Event.Amount
					if hand := effects.CardsInHand(g, victim); n > hand {
						n = hand
					}
					if n <= 0 {
						return nil
					}
					g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
						Player:   victim,
						Source:   item.SourceCardID,
						N:        n,
						Question: "Dreamstealer — discard that many cards",
					})
					return nil
				}),
		},
		Activated: []effects.ActivatedAbility{effects.Eternalize("{4}{B}{B}")},
	})
}
