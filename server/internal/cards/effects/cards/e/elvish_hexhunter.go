package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvish Hexhunter — Creature — Elf Shaman {G/W}, 1/1 (Shadowmoor, #5591):
//
//	"{G/W}, {T}, Sacrifice this creature: Destroy target enchantment."
//
// Ronom Unicorn's destruction with a hybrid mana cost and a tap, all paid
// at announce (CR 602.2b), so a summoning-sick Hexhunter cannot use it
// (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f8f450c1-f405-4ccc-aef3-f0186d1f5056",
		Name:         "Elvish Hexhunter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G/W}, {T}, Sacrifice this creature: Destroy target enchantment.",
			Cost:    effects.Plus(effects.ManaCost("{G/W}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target enchantment", effects.Enchantment()),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
