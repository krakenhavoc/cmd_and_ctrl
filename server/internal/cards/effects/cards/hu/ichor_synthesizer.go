package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ichor Synthesizer — Creature — Phyrexian Wizard {1}{U}, 1/3 (Phyrexia: All Will Be One):
//
//	"Whenever you cast a noncreature spell, put an oil counter on this creature.
//	 As long as this creature has four or more oil counters on it, it gets +2/+0
//	 and can't be blocked."
//
// One printed ability, two static effects (a layer 7c bonus and a layer 6 block
// restriction), both reading the live oil count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e82f46f3-143b-46ae-ab5d-4aa177437a64",
		Name:         "Ichor Synthesizer",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82OneB09OilOnNoncreatureCast("Ichor Synthesizer")},
		Static: []game.StaticAbility{
			{
				Label:    "As long as this creature has four or more oil counters on it, it gets +2/+0.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.InstanceID == source.InstanceID && effects.S82OneB09OilCount(source) >= 4
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) { c.Power += 2 },
			},
			{
				Label: "As long as this creature has four or more oil counters on it, it can't be blocked.",
				Layer: game.Layer6Ability,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.InstanceID == source.InstanceID && effects.S82OneB09OilCount(source) >= 4
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Restrictions |= game.CantBeBlocked
				},
			},
		},
	})
}
