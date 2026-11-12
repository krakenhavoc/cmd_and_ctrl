package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Charging Troll — Creature — Troll {2}{G}{W}, 3/3 (Invasion, #5616):
//
//	"Vigilance
//	 {G}: Regenerate this creature."
//
// Vigilance is a printed keyword. The regeneration is Albino Troll's: a
// shield on the creature itself (CR 701.19a), as many as you pay for, at
// instant speed and with no tap, so it can be used while attacking.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bbc1b96d-5a63-4091-a2c7-08bd289e0bf2",
		Name:            "Charging Troll",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "{G}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
