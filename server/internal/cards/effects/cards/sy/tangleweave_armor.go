package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b9cec31a-9891-414f-b528-954612168832",
		Name:         "Tangleweave Armor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LivingWeapon("Tangleweave Armor"),
		},
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, func(g *game.Game, source *game.Card) int {
				return g.GreatestCommanderManaValueForEffect(source.Controller)
			}),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{4}"),
		},
	})
}
