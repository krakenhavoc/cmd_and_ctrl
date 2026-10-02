package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deepcavern Imp — Creature — Imp Rebel, {2}{B}, 2/2:
//
//	"Flying, haste
//	 Echo—Discard a card. (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)"
//
// The echo cost is a discard (ADR 0108 §5): the controller picks the card
// from their hand as they pay, and with an empty hand they cannot pay.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1f295f2f-969a-4b82-98d1-7fe307ec83a7",
		Name:            "Deepcavern Imp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			EchoPayment("Deepcavern Imp", DiscardPayment(1)),
		},
	})
}
