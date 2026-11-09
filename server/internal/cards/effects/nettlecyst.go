package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nettlecyst — Artifact — Equipment {3}:
//
//	"Living weapon (When this Equipment enters, create a 0/0 black
//	 Phyrexian Germ creature token, then attach this to it.)
//	 Equipped creature gets +1/+1 for each artifact and/or enchantment you
//	 control.
//	 Equip {2}"
//
// Batterskull's living weapon and equip shape with Lion Sash's counted
// pump. A permanent that is both an artifact and an enchantment counts
// once (the predicate is one "is either" test per permanent), and
// Nettlecyst itself is an artifact it counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "04c7f4fe-2098-4311-866d-6733c08d5178",
		Name:         "Nettlecyst",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			LivingWeapon("Nettlecyst"),
		},
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 1, func(g *game.Game, source *game.Card) int {
				return countControlled(g, source.Controller, func(c game.Card) bool {
					return c.IsArtifact() || c.IsEnchantment()
				})
			}),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
