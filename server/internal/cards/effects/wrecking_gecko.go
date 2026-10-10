package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wrecking Gecko — Artifact Creature — Lizard Construct {4}{G}, 5/5:
//
//	"Ward {2}
//	 {6}{G}{G}: This creature gets +4/+4 and gains trample until end of
//	 turn."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4a9e0294-c967-4e90-b776-6d3195034a23",
		Name:         "Wrecking Gecko",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Wrecking Gecko — ward {2}"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{6}{G}{G}: This creature gets +4/+4 and gains trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
			Cost:    ManaCost("{6}{G}{G}"),
			Effect:  thisCreatureUntilEOT("Wrecking Gecko — +4/+4 and trample", 4, 4, "trample"),
		}},
	})
}
