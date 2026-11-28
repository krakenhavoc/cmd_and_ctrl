package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sacred Guide — Creature — Human Cleric {W}, 1/1 (Tempest, S82 OLD batch 766,
// #5628):
//
//	"{1}{W}, Sacrifice this creature: Reveal cards from the top of your library
//	 until you reveal a white card. Put that card into your hand and exile all
//	 other cards revealed this way."
//
// The sacrifice is the cost. The run is revealed as one reveal (Spoils of the
// Vault's shape), the white card goes to hand and every other revealed card is
// exiled once that move has settled; with no white card in the library every
// card is exiled and nothing goes to hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "de9e8b71-897d-4a9d-942b-af10f31e0425",
		Name:         "Sacred Guide",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, Sacrifice this creature: Reveal cards from the top of your library until you reveal a white card. Put that card into your hand and exile all other cards revealed this way.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.SacrificeThis()),
			Effect:  effects.SacredGuideReveal,
		}},
	})
}
