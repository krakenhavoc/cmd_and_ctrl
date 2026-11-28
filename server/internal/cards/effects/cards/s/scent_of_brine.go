package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scent of Brine — Instant {1}{U} (Urza's Destiny, S82 #5621):
//
//	"Reveal any number of blue cards in your hand. Counter target spell unless its
//	 controller pays {1} for each card revealed this way."
//
// OldB706RevealAnyNumber, then Miscalculation's counter-unless-pays for that many
// generic mana: with no card revealed the cost is {0}, which any spell's
// controller pays. A spell that has left the stack does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c22b5968-adea-4f74-877d-8ac2ee452769",
		Name:         "Scent of Brine",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve:    effects.OldB706RevealAnyNumber("Scent of Brine", "U", "choose-cards/s82-old-b706-scent-of-brine", effects.OldB706CounterUnlessPaysPerRevealed("Scent of Brine")),
	})
}
