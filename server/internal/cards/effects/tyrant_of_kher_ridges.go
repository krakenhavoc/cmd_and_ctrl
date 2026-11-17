package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tyrant of Kher Ridges — Creature — Dragon {4}{R}{R}, 4/5 (The Brothers' War, #5482):
//
//	"Flying
//	 When this creature enters, it deals 4 damage to any target.
//	 {R}: This creature gets +1/+0 until end of turn."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e7c06af6-8d96-4bfa-8baf-9e80f741e8cb",
		Name:            "Tyrant of Kher Ridges",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Tyrant of Kher Ridges — 4 damage to any target",
				sourceDealsDamageToEachLegalTarget(4)), TargetAny()),
		},
		Activated: []ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{R}"),
			Effect:  thisGetsUntilEndOfTurn(1, 0, "Tyrant of Kher Ridges — +1/+0 until end of turn"),
		}},
	})
}
