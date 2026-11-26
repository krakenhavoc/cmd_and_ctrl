package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blood-Cursed Knight — Creature — Vampire Knight {1}{W}{B}, 3/2 (Magic
// Origins, S82 older sets):
//
//	"As long as you control an enchantment, this creature gets +1/+1 and has
//	 lifelink."
//
// Briarbridge Tracker's conditional static over an enchantment: a layer 7c
// +1/+1 and a layer 6 lifelink grant (CR 613.1f), both read live on the
// source alone, so they end the moment the last enchantment you control
// leaves. The Knight is not an enchantment, so it never turns itself on.
//
// No simplification.
func init() {
	whileEnchantment := effects.SelfWhileYouControlA(effects.Enchantment())
	grant := effects.KeywordGrant(whileEnchantment, "lifelink")
	grant.Label = "As long as you control an enchantment, this creature has lifelink."
	effects.Register(effects.Spec{
		OracleID:     "fdf51294-6270-4c94-a2a9-8df37dfcaa0d",
		Name:         "Blood-Cursed Knight",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "As long as you control an enchantment, this creature gets +1/+1.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: whileEnchantment,
				Apply:     effects.PumpBy(1, 1),
			},
			grant,
		},
	})
}
