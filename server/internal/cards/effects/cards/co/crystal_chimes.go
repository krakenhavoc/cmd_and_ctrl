package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crystal Chimes — Artifact {3} (Urza's Saga):
//
//	"{3}, {T}, Sacrifice this artifact: Return all enchantment cards from
//	 your graveyard to your hand."
//
// Praetor's Counsel's return, narrowed to enchantment cards. The cards
// are read as the ability resolves, so an enchantment put into your
// graveyard in response comes back too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fbff7452-1da4-4cd0-8247-a625e4b4da2b",
		Name:         "Crystal Chimes",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice this artifact: Return all enchantment cards from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ReturnWholeGraveyardToHand(effects.NewContext(g, item), game.Card.IsEnchantment)
			},
		}},
	})
}
