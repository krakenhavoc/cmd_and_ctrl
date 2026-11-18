package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rust Monster — Creature — Beast {2}{R}, 2/1 (Adventures in the
// Forgotten Realms):
//
//	"First strike
//	 Sacrifice an artifact: This creature gets +2/+0 until end of turn."
//
// Atog's shape: a free activation whose only cost is one artifact the
// controller picks at announce, each its own +2/+0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3da90ae6-a8cc-4633-bf49-7413ce91b697",
		Name:            "Rust Monster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice an artifact: This creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerPump},
			Cost:    SacrificeN(1, "an artifact", Artifact()),
			Effect:  thisGetsUntilEndOfTurn(2, 0, "Rust Monster — +2/+0"),
		}},
	})
}
