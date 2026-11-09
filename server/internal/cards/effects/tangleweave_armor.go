package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tangleweave Armor — Artifact — Equipment {2}{G}{G}:
//
//	"Living weapon (When this Equipment enters, create a 0/0 black
//	 Phyrexian Germ creature token, then attach this to it.)
//	 Equipped creature gets +X/+X, where X is the greatest mana value among
//	 your commanders.
//	 Equip {4}"
//
// Batterskull's living weapon and equip shape. X is read live from the
// controller's commanders wherever they are (a ruling: it does not matter
// where the commander is), by game.GreatestCommanderManaValueForEffect, so
// a Germ equipped with no commander in the deck stays 0/0 and dies.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b9cec31a-9891-414f-b528-954612168832",
		Name:         "Tangleweave Armor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			LivingWeapon("Tangleweave Armor"),
		},
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 1, func(g *game.Game, source *game.Card) int {
				return g.GreatestCommanderManaValueForEffect(source.Controller)
			}),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{4}"),
		},
	})
}
