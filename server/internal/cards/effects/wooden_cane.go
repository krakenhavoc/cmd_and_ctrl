package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wooden Cane — Artifact — Equipment {3}{W} (TMNT Commander, #5011):
//
//	"When this Equipment enters, create a 2/2 red Mutant creature token,
//	 then attach this Equipment to it.
//	 Equipped creature gets +2/+1.
//	 Equip {3}"
//
// Dwarven Shortsword's shape (createTokenAndAttachSource) with the
// Mutant token: if the Cane has left in response, the Mutant is made and
// nothing is attached (CR 701.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d41eec6c-de14-4809-abae-b94355674a33",
		Name:         "Wooden Cane",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 1)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Wooden Cane — create a 2/2 red Mutant creature token, then attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					return createTokenAndAttachSource(g, item, "2/2 red Mutant", true)
				}),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
