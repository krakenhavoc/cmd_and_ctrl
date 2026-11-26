package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Idle Thoughts — Enchantment {3}{U}:
//
//	"{2}: Draw a card if you have no cards in hand."
//
// The hand is checked as the ability resolves, not when it is activated, so
// it can be activated with cards in hand and draws only if the hand has been
// emptied by then.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a2a40a19-8704-42fe-b9a6-5ae37c110303",
		Name:         "Idle Thoughts",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Draw a card if you have no cards in hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if p := g.PlayerByIDForEffect(item.Controller); p == nil || (p.Hand != nil && len(p.Hand.Cards) > 0) {
					return nil
				}
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
