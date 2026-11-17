package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Realmbreaker's Grasp — Enchantment — Aura {1}{W} (March of the Machine):
//
//	"Enchant artifact or creature
//	 Enchanted permanent can't attack or block, and its activated abilities
//	 can't be activated unless they're mana abilities."
//
// Arrest without the mana-ability clause: CantActivate stops activated
// abilities, and CantActivateMana (which Arrest adds) is left out because mana
// abilities stay usable. The enchant clause is an artifact or a creature, so
// the Aura stays on a Vehicle that stops being a creature (CR 704.5m).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f2baa0b9-ab0b-4ae9-b5f0-d5b3fcd9cf37",
		Name:         "Realmbreaker's Grasp",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("enchant artifact or creature", Or(Artifact(), Creature())),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			RestrictAttached(game.CantAttackOrBlock | game.CantActivate),
		},
	})
}
