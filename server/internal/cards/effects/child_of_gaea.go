package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Child of Gaea — Creature — Elemental {3}{G}{G}{G}, 7/7 (Urza's Saga):
//
//	"Trample
//	 At the beginning of your upkeep, sacrifice this creature unless you
//	 pay {G}{G}.
//	 {1}{G}: Regenerate this creature."
//
// Trample is the printed keyword, the upkeep is the Elder Dragons'
// UpkeepSacrificeUnlessYouPay, and the regeneration is Albino Troll's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d4e8aa93-c0d1-49b7-bbf4-35a155048774",
		Name:            "Child of Gaea",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			UpkeepSacrificeUnlessYouPay("Child of Gaea", "{G}{G}"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{1}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
