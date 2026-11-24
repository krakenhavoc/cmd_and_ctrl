package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Merfolk Windrobber — Creature — Merfolk Rogue {U}, 1/1 (Zendikar Rising,
// S82 OLD #5508):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, that player
//	 mills a card.
//	 Sacrifice this creature: Draw a card. Activate only if an opponent has
//	 eight or more cards in their graveyard."
//
// The mill is the damaged player's. The activation is gated on any
// opponent's graveyard (CR 602.5b), checked as it is activated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "be8b9bef-a98d-40a8-bedd-d78f1f08f93b",
		Name:            "Merfolk Windrobber",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Merfolk Windrobber — that player mills a card",
				func(g *game.Game, item *game.StackItem) error {
					ev := item.Trigger.Event
					if ev.Amount <= 0 || g.PlayerByIDForEffect(ev.Target) == nil {
						return nil
					}
					return effects.MillCards{Player: ev.Target, N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     "Sacrifice this creature: Draw a card. Activate only if an opponent has eight or more cards in their graveyard.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.SacrificeThis(),
			Condition: effects.S82OldB089AnOpponentHasEightInGraveyard,
			Effect:    effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
