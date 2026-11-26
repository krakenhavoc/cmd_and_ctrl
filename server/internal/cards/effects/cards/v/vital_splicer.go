package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vital Splicer — Creature — Phyrexian Human Artificer {3}{G}, 1/1:
//
//	"When this creature enters, create a 3/3 colorless Phyrexian Golem
//	 artifact creature token.
//	 {1}: Regenerate target Golem you control."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1899133-635d-46ff-b90d-6f8d46a1ffe1",
		Name:         "Vital Splicer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Vital Splicer — create a 3/3 colorless Phyrexian Golem artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 colorless Phyrexian Golem artifact"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: Regenerate target Golem you control.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{1}"),
			Targets: effects.TargetPermanent("target Golem you control", effects.HasSubtype("Golem"), effects.YouControl()),
			Effect:  effects.RegenerateTheTargetPermanent,
		}},
	})
}
