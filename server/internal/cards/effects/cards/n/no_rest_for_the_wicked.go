package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// No Rest for the Wicked — Enchantment {1}{B} (Urza's Saga, #5624):
//
//	"Sacrifice this enchantment: Return to your hand all creature cards in
//	 your graveyard that were put there from the battlefield this turn."
//
// The sacrifice is the cost. The set is read as the ability resolves, from
// the controller's own graveyard, and is returned together.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d03fd358-ff3f-45ab-9601-07dbb535cb51",
		Name:         "No Rest for the Wicked",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this enchantment: Return to your hand all creature cards in your graveyard that were put there from the battlefield this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.SacrificeThis(),
			Effect:  effects.S82OldB726ReturnCreaturesThatDiedThisTurn,
		}},
	})
}
