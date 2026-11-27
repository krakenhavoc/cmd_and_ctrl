package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Corrupting Licid — Creature — Licid {2}{B}, 2/2 (Stronghold, S82 OLD batch 746, #5627):
//
//	"{B}, {T}: This creature loses this ability and becomes an Aura
//	 enchantment with enchant creature. Attach it to target creature. You may
//	 pay {B} to end this effect.
//	 Enchanted creature has fear. (It can't be blocked except by artifact creatures and/or black creatures.)"
//
// The licid seam (#5781, CR 116.2c): LicidAbility is the printed ability,
// LicidForm what it turns the permanent into, and ReturnedAura the "enchant
// creature" clause. "Enchanted creature has fear" is Fear's grant. Paying to end the effect is the engine's end_effect
// special action.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "503ddd47-4b0c-4a2a-b77a-fb321044b438",
		Name:         "Corrupting Licid",
		Completeness: effects.CompletenessFull,
		ReturnedAura: effects.EnchantCreature(),
		Static:       append(effects.LicidForm(0), effects.GrantToAttached("fear")),
		Activated:    []effects.ActivatedAbility{effects.LicidAbility("{B}", "{B}")},
	})
}
