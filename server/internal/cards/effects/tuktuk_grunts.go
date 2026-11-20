package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tuktuk Grunts — Creature — Goblin Warrior Ally {4}{R}, 2/2 (Zendikar, #5583):
//
//	"Haste
//	 Whenever this creature or another Ally you control enters, you may
//	 put a +1/+1 counter on this creature."
//
// Haste is the printed keyword. The Zendikar Ally trigger
// (zen_b2_helpers.go): its own entry or another Ally's under its
// controller's control, then an optional +1/+1 counter on itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "200ce414-8583-4bbd-88a6-ca30babb613d",
		Name:            "Tuktuk Grunts",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered:       []game.TriggeredAbility{zenB2AllyCounter("Tuktuk Grunts")},
	})
}
