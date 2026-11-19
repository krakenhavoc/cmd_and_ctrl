package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trickster's Elk — Enchantment Creature — Elk {2}{G}, 3/3 (Modern Horizons 3):
//
//	"Bestow {1}{G}
//	 Enchanted creature loses all abilities and is a green Elk creature with
//	 base power and toughness 3/3."
//
// Kenrith's Transformation's static set; it applies to the enchanted creature
// only, so an unattached Elk is an ordinary 3/3.
func init() {
	Register(Spec{
		OracleID:         "5bca05c1-1c0a-45f4-aa20-69f68b99cd7e",
		Name:             "Trickster's Elk",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Bestow("{1}{G}")},
		Static: []game.StaticAbility{
			SetAttachedTypes([]string{"Creature"}, []string{"Elk"}),
			SetAttachedColors("G"),
			LoseAllAbilities(),
			SetAttachedBasePT(3, 3),
		},
	})
}
