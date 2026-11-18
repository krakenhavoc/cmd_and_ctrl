package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ainok Bond-Kin — Creature — Dog Soldier {1}{W}, 2/1:
//
//	"Outlast {1}{W} ({1}{W}, {T}: Put a +1/+1 counter on this creature.
//	 Outlast only as a sorcery.)
//	 Each creature you control with a +1/+1 counter on it has first
//	 strike."
//
// Outlast is the activated ability Adaptive Gemguard's shape names
// (CR 702.107a), at sorcery speed. The static is Duskshell Crawler's
// layer-6 grant over every creature its controller controls with a +1/+1
// counter, itself included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8a4028cd-6004-4fb0-a044-0bee662a78e1",
		Name:         "Ainok Bond-Kin",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "Outlast {1}{W}",
			Cost:         Plus(ManaCost("{1}{W}"), TapCost()),
			SorcerySpeed: true,
			Effect:       Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
		Static: []game.StaticAbility{{
			Label: "Each creature you control with a +1/+1 counter on it has first strike.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && target.Counters["+1/+1"] > 0
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, k := range c.Abilities {
					if k == "first strike" {
						return
					}
				}
				c.Abilities = append(c.Abilities, "first strike")
			},
		}},
	})
}
