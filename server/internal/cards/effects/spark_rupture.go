package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spark Rupture — Enchantment {2}{W} (March of the Machine: The
// Aftermath, #5476):
//
//	"When this enchantment enters, draw a card.
//	 Each planeswalker with one or more loyalty counters on it loses all
//	 abilities and is a creature with power and toughness each equal to
//	 the number of loyalty counters on it."
//
// One static sentence, three continuous effects over the same set (any
// player's planeswalkers with a loyalty counter, read as the layers are
// applied): the creature card type is added in layer 4 (it is still a
// planeswalker, CR 613.1d), every ability is removed in layer 6
// (CR 613.1f, declared with RemovesAbilities so the catalogued loyalty
// abilities and statics stop working, Sludge Monster's form), and the
// power and toughness are set to the number of loyalty counters in layer
// 7b (CR 613.4b). The counters are counted live on each layer pass, so
// the creature grows and shrinks with its loyalty and one with no counters
// is no longer in the set and stops being a creature. Damage dealt to it
// is damage to a creature that is also a planeswalker (CR 120.3): it
// removes loyalty counters and is marked as damage, so it shrinks as
// counters go and dies by lethal damage or by zero toughness, the way
// CR 704.5f and 704.5g say.
//
// No simplification.
func sparkRuptureScope(target *game.Card, _ *game.Game, _ *game.Card) bool {
	return target.IsPlaneswalker() && target.Counters[game.CounterLoyalty] > 0
}

func init() {
	Register(Spec{
		OracleID:     "82e5642a-0803-432c-85d3-091fca75c82f",
		Name:         "Spark Rupture",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "Each planeswalker with one or more loyalty counters on it is a creature.",
				Layer:     game.Layer4Type,
				AppliesTo: sparkRuptureScope,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !slices.Contains(c.Types, "Creature") {
						c.Types = append(c.Types, "Creature")
					}
				},
			},
			{
				Label:            "Each planeswalker with one or more loyalty counters on it loses all abilities.",
				Layer:            game.Layer6Ability,
				RemovesAbilities: true,
				AppliesTo:        sparkRuptureScope,
				Apply:            func(*game.Characteristic, *game.Card, *game.Game, *game.Card) {},
			},
			{
				Label:     "Each planeswalker with one or more loyalty counters on it has power and toughness each equal to the number of loyalty counters on it.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7B_Set,
				AppliesTo: sparkRuptureScope,
				Apply: func(c *game.Characteristic, target *game.Card, _ *game.Game, _ *game.Card) {
					n := target.Counters[game.CounterLoyalty]
					c.Power, c.Toughness = n, n
				},
			},
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Spark Rupture — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
