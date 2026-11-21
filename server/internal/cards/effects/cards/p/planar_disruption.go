package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Planar Disruption — Enchantment — Aura {1}{W} (Phyrexia: All Will Be One):
//
//	"Enchant artifact, creature, or planeswalker
//	 Enchanted permanent can't attack or block, and its activated abilities
//	 can't be activated."
//
// Arrest's lock on a wider host. Unlike Realmbreaker's Grasp the printed text
// has no "unless they're mana abilities" rider, so CantActivateMana is set as
// well: a mana ability is an activated ability (CR 605.1a), and the enchanted
// Sol Ring or mana creature taps for nothing. The enchant clause is the
// three-type union, so the Aura stays on a Vehicle that stops being a
// creature (CR 704.5m).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6128679e-7e54-41e4-ab6e-25aeeffed49f",
		Name:         "Planar Disruption",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("enchant artifact, creature, or planeswalker", effects.Or(effects.Artifact(), effects.Creature(), effects.Planeswalker())),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana),
		},
	})
}
