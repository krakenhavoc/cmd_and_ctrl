package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Umara Raptor — Creature — Bird Ally {2}{U}, 1/1 (Zendikar, #5583):
//
//	"Flying
//	 Whenever this creature or another Ally you control enters, you may
//	 put a +1/+1 counter on this creature."
//
// Flying is the printed keyword. The Zendikar Ally trigger
// (zen_b2_helpers.go): its own entry or another Ally's under its
// controller's control, then an optional +1/+1 counter on itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a58ee84f-1d9c-4924-b7b1-14a9b2ba3b98",
		Name:            "Umara Raptor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{zenB2AllyCounter("Umara Raptor")},
	})
}
