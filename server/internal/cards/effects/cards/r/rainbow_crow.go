package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rainbow Crow (Invasion, S82 OLD batch 646, #5616):
//
//	Creature — Bird {3}{U}, 2/2:
//	"Flying
//	{1}: This creature becomes the color of your choice until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "79bf98c8-1169-477d-838e-2ebc0fa396dc",
		Name:            "Rainbow Crow",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: This creature becomes the color of your choice until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.ManaCost("{1}"),
			Effect:  effects.S82OldB646ThisBecomesChosenColor("Rainbow Crow"),
		}},
	})
}
