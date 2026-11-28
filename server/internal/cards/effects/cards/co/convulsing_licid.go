package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Convulsing Licid — Creature — Licid {2}{R}, 2/2 (Stronghold, S82 OLD batch 746, #5627):
//
//	"{R}, {T}: This creature loses this ability and becomes an Aura
//	 enchantment with enchant creature. Attach it to target creature. You may
//	 pay {R} to end this effect.
//	 Enchanted creature can't block."
//
// The licid seam (#5781, CR 116.2c): LicidAbility is the printed ability,
// LicidForm what it turns the permanent into, and ReturnedAura the "enchant
// creature" clause. "Enchanted creature can't block" is Sluggishness's restriction, checked as blockers are declared. Paying to end the effect is the engine's end_effect
// special action.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "893c88cb-72ba-4f54-809c-bfac4d6efb74",
		Name:         "Convulsing Licid",
		Completeness: effects.CompletenessFull,
		ReturnedAura: effects.EnchantCreature(),
		Static:       append(effects.LicidForm(0), effects.RestrictAttached(game.CantBlock)),
		Activated:    []effects.ActivatedAbility{effects.LicidAbility("{R}", "{R}")},
	})
}
