package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enraging Licid — Creature — Licid {1}{R}, 1/1 (Tempest, S82 OLD b753,
// #5628):
//
//	"{R}, {T}: This creature loses this ability and becomes an Aura
//	 enchantment with enchant creature. Attach it to target creature. You
//	 may pay {R} to end this effect.
//	 Enchanted creature has haste."
//
// The licid shape (LicidAbility, LicidForm, ReturnedAura) with Prophetic
// Ravings' haste grant to the host.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ec9f317-887a-4e08-a165-e7cf15eb071d",
		Name:         "Enraging Licid",
		Completeness: effects.CompletenessFull,
		ReturnedAura: effects.EnchantCreature(),
		Static:       append(effects.LicidForm(0), effects.GrantToAttached("haste")),
		Activated:    []effects.ActivatedAbility{effects.LicidAbility("{R}", "{R}")},
	})
}
