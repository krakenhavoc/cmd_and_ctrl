package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oran-Rief Survivalist — Creature — Human Warrior Ally {1}{G}, 1/1 (Zendikar, #5583):
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
	effects.Register(effects.Spec{
		OracleID:     "36b4c44a-7682-4876-ae2a-35112cb55ec0",
		Name:         "Oran-Rief Survivalist",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.ZenB2AllyCounter("Oran-Rief Survivalist")},
	})
}
