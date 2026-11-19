package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ab28a801-6ab1-4db2-8ae2-c24bb4c10ebb",
		Name:         "Ancestral Blade",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ancestral Blade — create a 1/1 white Soldier creature token, then attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					return createTokenAndAttachSource(g, item, "1/1 white Soldier", true)
				}),
		},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}
