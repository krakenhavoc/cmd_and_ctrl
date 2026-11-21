package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seascape Aerialist — Creature — Merfolk Wizard Ally {4}{U}, 2/3 (Zendikar, #5583):
//
//	"Whenever this creature or another Ally you control enters, you may
//	 have Ally creatures you control gain flying until end of turn."
//
// The Zendikar Ally trigger (zen_b2_helpers.go) with Indelicate
// Healer's team grant: the Ally creatures its controller controls as
// the ability resolves gain the keyword until end of turn (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "877c6bb4-487c-4eea-b4af-f28fe1515fb6",
		Name:         "Seascape Aerialist",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.ZenB2AllyGrant("Seascape Aerialist", "flying")},
	})
}
