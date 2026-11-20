package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ravenous Intruder — Creature — Gremlin {1}{R}, 1/2 (Aether Revolt,
// #5536):
//
//	"Sacrifice an artifact: This creature gets +2/+2 until end of turn."
//
// Atog: each activation eats one artifact and is its own +2/+2 until
// end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "06a94ea3-6cda-42fe-ba3b-40a0e9044f8e",
		Name:         "Ravenous Intruder",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice an artifact: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    SacrificeN(1, "an artifact", Artifact()),
			Effect:  thisGetsUntilEndOfTurn(2, 2, "Ravenous Intruder — +2/+2"),
		}},
	})
}
