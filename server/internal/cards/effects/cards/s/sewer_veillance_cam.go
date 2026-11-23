package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sewer-veillance Cam — {U} Artifact (Teenage Mutant Ninja Turtles, #4457):
//
//	"Flash
//	 When this artifact enters or leaves the battlefield, you may tap or
//	 untap target creature.
//	 {3}{U}, Sacrifice this artifact: Draw two cards."
//
// One trigger with two conditions (WhenThisEntersOrLeaves, CR 603.6c for the
// leaving half), made optional and given Merrow Reejerey's tap-or-untap pick
// at resolution. Sacrificing it for the cards triggers the leaving half.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "25c83d45-581f-48ba-b6bb-feb66e8a65c7",
		Name:            "Sewer-veillance Cam",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.Targeting(
					effects.WhenThisEntersOrLeaves("Sewer-veillance Cam — tap or untap target creature",
						effects.B38TapOrUntapTarget("Sewer-veillance Cam")),
					effects.TargetCreature("target creature")),
				"Sewer-veillance Cam — tap or untap target creature?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}, Sacrifice this artifact: Draw two cards.",
			Cost:    effects.Plus(effects.ManaCost("{3}{U}"), effects.SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 2},
			Effect:  effects.Do(effects.DrawCards{N: 2}),
		}},
	})
}
