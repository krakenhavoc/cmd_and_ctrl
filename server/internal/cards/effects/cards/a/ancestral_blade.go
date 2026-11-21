package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ancestral Blade — Artifact — Equipment {1}{W} (Foundations Jumpstart):
//
//	"When this Equipment enters, create a 1/1 white Soldier creature token,
//	 then attach this Equipment to it.
//	 Equipped creature gets +1/+1.
//	 Equip {1}"
//
// Auxiliary Boosters' entry trigger over the catalog's 1/1 white Soldier.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ab28a801-6ab1-4db2-8ae2-c24bb4c10ebb",
		Name:         "Ancestral Blade",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ancestral Blade — create a 1/1 white Soldier creature token, then attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateTokenAndAttachSource(g, item, "1/1 white Soldier", true)
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
