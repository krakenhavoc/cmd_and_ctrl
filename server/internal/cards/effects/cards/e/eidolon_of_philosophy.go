package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eidolon of Philosophy — Enchantment Creature — Spirit {U}, 1/2 (Theros Beyond Death, #5514):
//
//	"{6}{U}, Sacrifice this creature: Draw three cards."
//
// Shore Keeper's draw-three, with the sacrifice and no tap.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "65acea17-c900-4574-b22f-910285552aa2",
		Name:         "Eidolon of Philosophy",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}{U}, Sacrifice this creature: Draw three cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{6}{U}"), effects.SacrificeThis()),
			Effect:  effects.Do(effects.DrawCards{N: 3}),
		}},
	})
}
