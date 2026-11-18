package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "83c69c05-173d-4c1b-9541-1dde474fef5f",
		Name:         "Bonehoard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			LivingWeapon("Bonehoard"),
		},
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 1, func(g *game.Game, _ *game.Card) int {
				return b41CreatureCardsInAllGraveyards(g)
			}),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
