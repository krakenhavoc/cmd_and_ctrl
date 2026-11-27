package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Foratog — Creature — Atog {2}{G} 1/2 (Eighth Edition, S82 older sets batch 619, #5607):
//
//	"{G}, Sacrifice a Forest: This creature gets +2/+2 until end of turn."
//
// Rootwater Alligator's Forest sacrifice with a mana cost and a pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77f257b9-2047-40a3-85c2-00ad17deeb08",
		Name:         "Foratog",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, Sacrifice a Forest: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.SacrificeN(1, "a Forest", effects.Subtype("Forest"))),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 2, "Foratog — +2/+2 until end of turn"),
		}},
	})
}
