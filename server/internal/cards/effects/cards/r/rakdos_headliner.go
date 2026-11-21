package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rakdos Headliner — Creature — Devil, {B}{R}, 3/3:
//
//	"Haste
//	 Echo—Discard a card. (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)"
//
// The echo cost is a discard (ADR 0108 §5): the controller picks the card
// from their hand as they pay, and with an empty hand they cannot pay.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a8828c17-190b-4e64-a0f2-8d14c013d692",
		Name:            "Rakdos Headliner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.EchoPayment("Rakdos Headliner", effects.DiscardPayment(1)),
		},
	})
}
