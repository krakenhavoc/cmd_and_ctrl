package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Makindi Shieldmate — Creature — Kor Soldier Ally {2}{W}, 0/3 (Zendikar, #5583):
//
//	"Defender
//	 Whenever this creature or another Ally you control enters, you may
//	 put a +1/+1 counter on this creature."
//
// Defender is the printed keyword. The Zendikar Ally trigger
// (zen_b2_helpers.go): its own entry or another Ally's under its
// controller's control, then an optional +1/+1 counter on itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b8a65e49-2379-42bf-823e-c208c2550aeb",
		Name:            "Makindi Shieldmate",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered:       []game.TriggeredAbility{zenB2AllyCounter("Makindi Shieldmate")},
	})
}
