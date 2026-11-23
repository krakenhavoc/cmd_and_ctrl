package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bonehoard — Artifact — Equipment {4}:
//
//	"Living weapon (When this Equipment enters, create a 0/0 black
//	 Phyrexian Germ creature token, then attach this to it.)
//	 Equipped creature gets +X/+X, where X is the number of creature
//	 cards in all graveyards.
//	 Equip {2}"
//
// The shared LivingWeapon trigger, PumpAttachedPer with Mortivore's
// all-graveyards creature-card count (re-read on every recompute, so the
// bonus tracks the graveyards), and one equip ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "83c69c05-173d-4c1b-9541-1dde474fef5f",
		Name:         "Bonehoard",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LivingWeapon("Bonehoard"),
		},
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, func(g *game.Game, _ *game.Card) int {
				return effects.B41CreatureCardsInAllGraveyards(g)
			}),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{2}"),
		},
	})
}
