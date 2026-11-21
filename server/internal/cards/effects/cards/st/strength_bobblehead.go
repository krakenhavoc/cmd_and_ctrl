package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Strength Bobblehead — Artifact — Bobblehead {3} (Fallout, #5469):
//
//	"{T}: Add one mana of any color.
//	 {3}, {T}: Put X +1/+1 counters on target creature, where X is the
//	 number of Bobbleheads you control. Activate only as a sorcery."
//
// Intelligence Bobblehead's count (pipBobbleheadsYouControl), read as
// the ability resolves (CR 608.2h), put on the target as one placement
// of X counters. With its target gone the ability does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd0e22d1-4789-406d-9293-3ade30365b32",
		Name:         "Strength Bobblehead",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}, {T}: Put X +1/+1 counters on target creature, where X is the number of Bobbleheads you control. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets:      effects.TargetCreature("target creature"),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				target, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				n := effects.PipBobbleheadsYouControl(g, item.Controller)
				if n <= 0 {
					return nil
				}
				return effects.AddCounter{Target: target, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
			},
		}},
	})
}
