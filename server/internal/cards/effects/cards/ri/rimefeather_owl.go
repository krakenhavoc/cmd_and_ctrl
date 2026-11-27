package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rimefeather Owl — Snow Creature — Bird {5}{U}{U}, */* (Coldsnap, #5597):
//
//	"Flying
//	 Rimefeather Owl's power and toughness are each equal to the number of snow
//	 permanents on the battlefield.
//	 {1}{S}: Put an ice counter on target permanent.
//	 Permanents with ice counters on them are snow."
//
// The second static is layer 4 (CR 613.1d, 205.4g), before the characteristic-
// defining ability in layer 7a counts, so a permanent iced by this Owl or by
// another is counted by it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3c0f7c4d-ac93-4a58-a87f-bf694c709624",
		Name:            "Rimefeather Owl",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			{
				Label: "Permanents with ice counters on them are snow.",
				Layer: game.Layer4Type,
				AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
					return target.Counters[effects.S82OldB565IceCounter] > 0
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					for _, s := range c.Supertypes {
						if s == "Snow" || s == "snow" {
							return
						}
					}
					c.Supertypes = append(append([]string(nil), c.Supertypes...), "Snow")
				},
			},
			{
				Label:     "Rimefeather Owl's power and toughness are each equal to the number of snow permanents on the battlefield.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7A_CDA,
				AppliesTo: effects.SelfOnly,
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
					n := effects.S82OldB565SnowPermanentsOnTheBattlefield(g)
					c.Power, c.Toughness = n, n
				},
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{S}: Put an ice counter on target permanent.",
			Cost:    effects.ManaCost("{1}{S}"),
			Targets: effects.TargetPermanent("target permanent"),
			Effect:  effects.S82OldB565IceCounterOnTheTarget,
		}},
	})
}
