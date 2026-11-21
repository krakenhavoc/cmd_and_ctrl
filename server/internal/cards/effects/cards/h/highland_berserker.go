package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Highland Berserker — Creature — Human Berserker Ally {1}{R}, 2/1 (Zendikar, #5583):
//
//	"Whenever this creature or another Ally you control enters, you may
//	 have Ally creatures you control gain first strike until end of
//	 turn."
//
// The Zendikar Ally trigger (zen_b2_helpers.go) with Indelicate
// Healer's team grant: the Ally creatures its controller controls as
// the ability resolves gain the keyword until end of turn (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "633849f6-fad7-49d7-8bf5-2a3a4510965f",
		Name:         "Highland Berserker",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.ZenB2AllyGrant("Highland Berserker", "first strike")},
	})
}
