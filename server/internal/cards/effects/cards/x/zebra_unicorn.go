package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zebra Unicorn — Creature — Unicorn {2}{G}{W}, 2/2 (Mirage, #5633):
//
//	"Whenever this creature deals damage, you gain that much life."
//
// Any damage, combat or not, to anything; each damage event is one trigger and
// gains the amount dealt after replacement (CR 120.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f91818f-a6fe-4472-b52b-ab574f790905",
		Name:         "Zebra Unicorn",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB793ThisDealtDamageGainThatMuch("Zebra Unicorn — you gain that much life"),
		},
	})
}
