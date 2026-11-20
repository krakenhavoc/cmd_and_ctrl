package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Witness of Tomorrows — Enchantment Creature — Sphinx {4}{U}, 3/4 (Theros Beyond Death, #5514):
//
//	"Flying
//	 {3}{U}: Scry 1."
//
// Flying is the printed keyword; the scry is an ordinary activated
// ability (Lorehold Campus' body without the tap).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b46adfcd-fd7a-40c5-85f4-1cfee43a4c48",
		Name:            "Witness of Tomorrows",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Cost:    ManaCost("{3}{U}"),
			Effect:  Do(Scry{N: 1}),
		}},
	})
}
