package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urborg Syphon-Mage — Creature — Human Spellshaper {2}{B}, 2/2
// (Dominaria Remastered):
//
//	"{2}{B}, {T}, Discard a card: Each other player loses 2 life. You gain
//	 life equal to the life lost this way."
//
// Blood Tithe's drain (B21DrainEachOpponentAndGainTheTotal) with the
// discard and tap in the cost. Each other player is each opponent, and
// the gain is the life actually lost: a player who can't lose life, or
// loses less, adds less (CR 119.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6608efe-b014-4b4d-a995-0be5047567c5",
		Name:         "Urborg Syphon-Mage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, {T}, Discard a card: Each other player loses 2 life. You gain life equal to the life lost this way.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.TapCost(), effects.DiscardACard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B21DrainEachOpponentAndGainTheTotal(g, item, 2)
			},
		}},
	})
}
