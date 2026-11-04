package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "25c83d45-581f-48ba-b6bb-feb66e8a65c7",
		Name:            "Sewer-veillance Cam",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Optional(
				Targeting(
					WhenThisEntersOrLeaves("Sewer-veillance Cam — tap or untap target creature",
						b38TapOrUntapTarget("Sewer-veillance Cam")),
					TargetCreature("target creature")),
				"Sewer-veillance Cam — tap or untap target creature?"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}, Sacrifice this artifact: Draw two cards.",
			Cost:    Plus(ManaCost("{3}{U}"), SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 2},
			Effect:  Do(DrawCards{N: 2}),
		}},
	})
}
