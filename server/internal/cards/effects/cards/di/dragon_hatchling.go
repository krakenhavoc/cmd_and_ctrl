package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragon Hatchling — Creature — Dragon {1}{R}, 0/1:
//
//	"Flying
//	 {R}: This creature gets +1/+0 until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3f48f6a7-906d-4287-8440-06bfabc93b89",
		Name:            "Dragon Hatchling",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Dragon Hatchling — +1/+0 until end of turn"),
		}},
	})
}
