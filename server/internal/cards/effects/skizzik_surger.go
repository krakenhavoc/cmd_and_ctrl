package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skizzik Surger — Creature — Elemental, {4}{R}{R}, 6/4:
//
//	"Haste
//	 Echo—Sacrifice two lands. (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)"
//
// The echo cost is sacrificing two lands (ADR 0108 §5): the controller
// picks both as they pay, and with fewer than two lands they cannot pay.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ded349de-0599-4744-ad3e-aec95caf9f99",
		Name:            "Skizzik Surger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			EchoPayment("Skizzik Surger", SacrificePayment(2, "land", "lands", game.PermanentQuery{Types: []string{"land"}})),
		},
	})
}
