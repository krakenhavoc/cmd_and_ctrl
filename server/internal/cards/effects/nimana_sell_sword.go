package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nimana Sell-Sword — Creature — Human Warrior Ally {3}{B}, 2/2 (Zendikar, #5583):
//
//	"Whenever this creature or another Ally you control enters, you may
//	 put a +1/+1 counter on this creature."
//
// The Zendikar Ally trigger (zen_b2_helpers.go): its own entry or
// another Ally's under its controller's control, then an optional +1/+1
// counter on itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5093df46-92cb-4f79-a5cf-4351c1fcf36d",
		Name:         "Nimana Sell-Sword",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{zenB2AllyCounter("Nimana Sell-Sword")},
	})
}
