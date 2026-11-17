package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abbey Matron — Creature — Human Cleric {2}{W}, 1/3 (Homelands, #5635):
//
//	"{W}, {T}: This creature gets +0/+3 until end of turn."
//
// The {T} keeps it off the turn it arrives (CR 602.5a). The boost is a
// layer 7c change (CR 613.4c) that ends at cleanup (CR 514.2), and
// nothing happens if the Matron has left by the time it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "62e3f285-886c-414e-b4ff-403a7c01c23a",
		Name:         "Abbey Matron",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}, {T}: This creature gets +0/+3 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Toughness: 3}},
			Cost:    Plus(ManaCost("{W}"), TapCost()),
			Effect:  thisGetsUntilEndOfTurn(0, 3, "Abbey Matron — +0/+3 until end of turn"),
		}},
	})
}
