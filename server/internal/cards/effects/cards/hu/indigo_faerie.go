package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Indigo Faerie — {1}{U} Creature — Faerie Wizard, 1/1:
//
//	"Flying
//	 {U}: Target permanent becomes blue in addition to its other colors until
//	 end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "10ae611f-e56d-47c5-8b6e-d38d95208bba",
		Name:            "Indigo Faerie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: Target permanent becomes blue in addition to its other colors until end of turn.",
			Cost:    effects.ManaCost("{U}"),
			Targets: effects.TargetPermanent("target permanent"),
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Effect:  effects.S82OldB523BlueInAdditionToItsOtherColors,
		}},
	})
}
