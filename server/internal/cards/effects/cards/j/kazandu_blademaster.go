package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kazandu Blademaster — Creature — Human Soldier Ally {W}{W}, 1/1 (Zendikar, #5583):
//
//	"First strike, vigilance
//	 Whenever this creature or another Ally you control enters, you may
//	 put a +1/+1 counter on this creature."
//
// First strike and vigilance are printed keywords. The Zendikar Ally
// trigger (zen_b2_helpers.go): its own entry or another Ally's under
// its controller's control, then an optional +1/+1 counter on itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "133f5d30-d883-493e-93a1-cf9583db460b",
		Name:            "Kazandu Blademaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "vigilance"},
		Triggered:       []game.TriggeredAbility{effects.ZenB2AllyCounter("Kazandu Blademaster")},
	})
}
