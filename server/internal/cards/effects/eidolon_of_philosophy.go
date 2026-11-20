package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eidolon of Philosophy — Enchantment Creature — Spirit {U}, 1/2 (Theros Beyond Death, #5514):
//
//	"{6}{U}, Sacrifice this creature: Draw three cards."
//
// Shore Keeper's draw-three, with the sacrifice and no tap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "65acea17-c900-4574-b22f-910285552aa2",
		Name:         "Eidolon of Philosophy",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{6}{U}, Sacrifice this creature: Draw three cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{6}{U}"), SacrificeThis()),
			Effect:  Do(DrawCards{N: 3}),
		}},
	})
}
