package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sea Legs — Enchantment — Aura {U} (Rivals of Ixalan, S82 OLD batch 217,
// #5531):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets +0/+2 as long as it's a Pirate. Otherwise, it gets
//	 -2/-0."
//
// One layer 7c modification whose sign depends on the creature as the earlier
// layers have left it (CR 613.4c, after the type-changing layer 4): a changeling
// or a creature turned into a Pirate gets +0/+2, one that stops being a Pirate
// gets -2/-0, on every recompute. The Pirate word is read through the Aura's
// own text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2e66dc2f-44f3-44f1-8ed7-7f991b3718be",
		Name:            "Sea Legs",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{{
			Label:     "Enchanted creature gets +0/+2 as long as it's a Pirate. Otherwise, it gets -2/-0.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.AttachedToSource,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				if c.AllCreatureTypes || effects.HasFold(c.Subtypes, source.TextSubtype("Pirate")) {
					c.Toughness += 2
					return
				}
				c.Power -= 2
			},
		}},
	})
}
