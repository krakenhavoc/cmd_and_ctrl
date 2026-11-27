package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gliding Licid — Creature — Licid {2}{U}, 2/2 (Stronghold, S82 OLD batch 746, #5627):
//
//	"{U}, {T}: This creature loses this ability and becomes an Aura
//	 enchantment with enchant creature. Attach it to target creature. You may
//	 pay {U} to end this effect.
//	 Enchanted creature has flying."
//
// The licid seam (#5781, CR 116.2c): LicidAbility is the printed ability,
// LicidForm what it turns the permanent into, and ReturnedAura the "enchant
// creature" clause. "Enchanted creature has flying" is Flight's grant. Paying to end the effect is the engine's end_effect
// special action.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a07b8619-224d-46d0-a17a-38d04aba574e",
		Name:         "Gliding Licid",
		Completeness: effects.CompletenessFull,
		ReturnedAura: effects.EnchantCreature(),
		Static:       append(effects.LicidForm(0), effects.GrantToAttached("flying")),
		Activated:    []effects.ActivatedAbility{effects.LicidAbility("{U}", "{U}")},
	})
}
