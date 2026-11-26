package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Encrust — Enchantment — Aura {1}{U}{U} (Magic 2015, #5556):
//
//	"Enchant artifact or creature
//	 Enchanted permanent doesn't untap during its controller's untap step and
//	 its activated abilities can't be activated."
//
// Bubble Snare's untap lock with Bound in Gold's activation lock, whole: no
// activated ability of the enchanted permanent can be activated, mana
// abilities included (CR 602.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "2e056743-78ff-45b6-80df-7486d52d5df6",
		Name:                  "Encrust",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantPermanent(effects.Or(effects.Artifact(), effects.Creature())),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantActivate | game.CantActivateMana),
		},
	})
}
