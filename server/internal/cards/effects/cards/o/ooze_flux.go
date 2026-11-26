package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ooze Flux — Enchantment {3}{G} (Gatecrash, #5565):
//
//	"{1}{G}, Remove one or more +1/+1 counters from among creatures you
//	 control: Create an X/X green Ooze creature token, where X is the number
//	 of +1/+1 counters removed this way."
//
// The cost is Jetfire's variable "from among" removal (#3276): the activator
// announces the total (at least one) and the split across creatures as the
// cost is paid, and the total is read off the paid-cost record as the
// ability resolves. The token is created 0/0 and sized to X as it is made
// (Seed Guardian's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "631d0de7-4810-4bc3-b86b-8e2505ed91f4",
		Name:         "Ooze Flux",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}, Remove one or more +1/+1 counters from among creatures you control: Create an X/X green Ooze creature token, where X is the number of +1/+1 counters removed this way.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
			Cost: effects.Plus(effects.ManaCost("{1}{G}"), game.AbilityCost{RemoveCounters: &game.CounterRemovalCost{
				Counter:  game.CounterPlusOne,
				N:        1,
				From:     effects.TargetPermanent("creatures you control", effects.Creature()),
				Among:    true,
				Variable: true,
			}}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := ctx.CountersRemoved()
				if x <= 0 {
					return nil
				}
				tok := ctx.TextToken(effects.TokenCard("0/0 green Ooze"))
				tok.Power, tok.Toughness = x, x
				return g.CreateTokenForEffect(ctx.Controller(), tok, 1)
			},
		}},
	})
}
