package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Petrify — Enchantment — Aura {1}{W} (The Lost Caverns of Ixalan,
// #4470):
//
//	"Enchant artifact or creature
//	 Enchanted permanent can't attack or block, and its activated
//	 abilities can't be activated."
//
// Arrest with Stuck in Summoner's Sanctum's enchant clause. The
// activation lock covers mana abilities too (CantActivateMana), since a
// mana ability is an activated ability (CR 605.1a) and the card spares
// none. On a noncreature artifact the attack and block half does nothing
// until it becomes a creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ebec110e-aff3-48aa-abf6-309dd2ffb478",
		Name:         "Petrify",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("enchant artifact or creature", Or(Artifact(), Creature())),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			RestrictAttached(
				game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana,
			),
		},
	})
}
