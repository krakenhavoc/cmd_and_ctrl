package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mire Triton — Creature — Zombie Merfolk {1}{B}, 2/1 (Foundations Commander):
//
//	"Deathtouch
//	 When this creature enters, mill two cards and you gain 2 life."
//
// Mill first, then the life, in the printed order (CR 701.17a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4fe6f784-0290-4d4a-85ae-0b66522d31e9",
		Name:            "Mire Triton",
		PrintedKeywords: []string{"deathtouch"},
		Completeness:    CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Mire Triton — mill two cards and gain 2 life",
				Do(MillCards{N: 2}, GainLife{Amount: 2})),
		},
	})
}
