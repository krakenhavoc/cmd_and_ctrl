package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phyrexian Boon — Enchantment — Aura {2}{B} (Alliances, #5634):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+1 as long as it's black. Otherwise, it
//	 gets -1/-2."
//
// A layer 7c change (CR 613.4c) to the enchanted creature whose sign
// depends on its colour. Colour is settled in layer 5, before this runs,
// so a creature painted black gets +2/+1 and one painted another colour
// gets -1/-2. "Black" is a word of the Aura's text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a18e6468-b8a2-4681-b8fa-223ab5f3061a",
		Name:         "Phyrexian Boon",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{{
			Label:     "Enchanted creature gets +2/+1 as long as it's black. Otherwise, it gets -1/-2.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: AttachedToSource,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				if !slices.Contains(c.Types, "Creature") {
					return
				}
				if slices.Contains(c.Colors, source.TextColor("B")) {
					c.Power += 2
					c.Toughness++
					return
				}
				c.Power--
				c.Toughness -= 2
			},
		}},
	})
}
