package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Throne of Geth — Artifact {2} (Scars of Mirrodin, #5577):
//
//	"{T}, Sacrifice an artifact: Proliferate."
//
// Atog's "sacrifice an artifact" cost (the Throne itself may be the one)
// beside the tap, then Thrummingbird's proliferate, which asks its
// controller which permanents and players get a counter (CR 701.34a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5a5f39ab-21b4-415f-aa1f-d8fad420d372",
		Name:         "Throne of Geth",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice an artifact: Proliferate.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(TapCost(), SacrificeN(1, "an artifact", Artifact())),
			Effect:  Do(Proliferate{}),
		}},
	})
}
