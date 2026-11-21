package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "b46adfcd-fd7a-40c5-85f4-1cfee43a4c48",
		Name:            "Witness of Tomorrows",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Cost:    effects.ManaCost("{3}{U}"),
			Effect:  effects.Do(effects.Scry{N: 1}),
		}},
	})
}
