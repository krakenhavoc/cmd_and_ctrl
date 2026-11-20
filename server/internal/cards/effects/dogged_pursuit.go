package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dogged Pursuit — Enchantment {3}{B} (Kaldheim, #5505):
//
//	"At the beginning of your end step, each opponent loses 1 life and
//	 you gain 1 life."
//
// The shared end-step trigger (AtYourEndStep) with Grave Venerations'
// drain: each opponent loses 1, then you gain 1 (a flat 1, not the sum
// lost).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "987f8c13-0164-4536-b505-d9fa799222fb",
		Name:         "Dogged Pursuit",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Dogged Pursuit — each opponent loses 1 life and you gain 1 life",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1})),
		},
	})
}
