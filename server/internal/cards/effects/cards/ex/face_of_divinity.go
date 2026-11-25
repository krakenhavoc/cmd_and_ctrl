package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Face of Divinity — Enchantment — Aura {2}{W} (Jumpstart, #5510):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2.
//	 As long as another Aura is attached to enchanted creature, it has first
//	 strike and lifelink."
//
// The pump is a layer 7c +2/+2. The keywords are a layer 6 grant (Champion's
// Helm's while-condition) that is re-read on every recompute: another Aura,
// anyone's, a Role included, on the creature switches first strike and
// lifelink on, and it goes off again when that Aura leaves. This Aura never
// counts as "another".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "90b5ed05-7a67-4bd3-a8c4-2a4565d084fc",
		Name:         "Face of Divinity",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttachedWhile(effects.S82OldB100AnotherAuraAttachedTo, "first strike", "lifelink"),
		},
	})
}
