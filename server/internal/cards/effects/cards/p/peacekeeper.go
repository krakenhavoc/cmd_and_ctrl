package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Peacekeeper — Creature — Human {2}{W}, 1/1 (Weatherlight, S82 older sets batch 779, #5629):
//
//	"At the beginning of your upkeep, sacrifice this creature unless you pay {1}{W}.
//	 Creatures can't attack."
//
// The Elder Dragons' upkeep tax (UpkeepSacrificeUnlessYouPay), and a layer-6 static that gives every creature on the
// battlefield, whoever controls it, the can't-attack restriction (CR 508.1c) while Peacekeeper is there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e82ce5c6-1e42-4e87-b696-67aa3822ec9c",
		Name:         "Peacekeeper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.UpkeepSacrificeUnlessYouPay("Peacekeeper", "{1}{W}"),
		},
		Static: []game.StaticAbility{effects.S82OldB779CreaturesCantAttack("Creatures can't attack.")},
	})
}
