package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trickster's Elk — Enchantment Creature — Elk {2}{G}, 3/3 (Modern Horizons 3):
//
//	"Bestow {1}{G}
//	 Enchanted creature loses all abilities and is a green Elk creature with
//	 base power and toughness 3/3."
//
// Kenrith's Transformation's static set; it applies to the enchanted creature
// only, so an unattached Elk is an ordinary 3/3.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "5bca05c1-1c0a-45f4-aa20-69f68b99cd7e",
		Name:             "Trickster's Elk",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Bestow("{1}{G}")},
		Static: []game.StaticAbility{
			effects.SetAttachedTypes([]string{"Creature"}, []string{"Elk"}),
			effects.SetAttachedColors("G"),
			effects.LoseAllAbilities(),
			effects.SetAttachedBasePT(3, 3),
		},
	})
}
