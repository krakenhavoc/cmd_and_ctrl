package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scent of Jasmine — Instant {W} (Urza's Destiny, S82 #5621):
//
//	"Reveal any number of white cards in your hand. You gain 2 life for each card
//	 revealed this way."
//
// OldB706RevealAnyNumber: the reveal is chosen as the spell resolves; a hand with
// no white card reveals nothing and gains nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a532c2b-a08c-46ab-86f7-c6dee0d6985f",
		Name:         "Scent of Jasmine",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.OldB706RevealAnyNumber("Scent of Jasmine", "W", "choose-cards/s82-old-b706-scent-of-jasmine", effects.OldB706GainTwoPerRevealed),
	})
}
