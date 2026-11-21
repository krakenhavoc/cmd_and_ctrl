package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Charisma Bobblehead — Artifact — Bobblehead {3} (Fallout, #5469):
//
//	"{T}: Add one mana of any color.
//	 {4}, {T}: Create X 1/1 white Soldier creature tokens, where X is
//	 the number of Bobbleheads you control. Activate only as a sorcery."
//
// Intelligence Bobblehead's count (pipBobbleheadsYouControl), read as
// the ability resolves (CR 608.2h), as the number of Soldiers made in
// one creation.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77897962-7d67-4fa8-858f-2f35a322d7c6",
		Name:         "Charisma Bobblehead",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{4}, {T}: Create X 1/1 white Soldier creature tokens, where X is the number of Bobbleheads you control. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			SorcerySpeed: true,
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				n := effects.PipBobbleheadsYouControl(ctx.Game, ctx.Controller())
				if n <= 0 {
					return nil
				}
				return effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: n}.Apply(ctx)
			})),
		}},
	})
}
