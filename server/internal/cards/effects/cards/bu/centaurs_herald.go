package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Centaur's Herald — Creature — Elf Scout {G}, 0/1 (Return to Ravnica, #5566):
//
//	"{2}{G}, Sacrifice this creature: Create a 3/3 green Centaur creature token."
//
// The sacrifice is a cost (CR 118.3): the Centaur comes even if the Herald is
// removed in response.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b47fd32d-1a56-414f-9d91-ddd46d4b5ac7",
		Name:         "Centaur's Herald",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}, Sacrifice this creature: Create a 3/3 green Centaur creature token.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{G}"), effects.SacrificeThis()),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 green Centaur"), N: 1}),
		}},
	})
}
