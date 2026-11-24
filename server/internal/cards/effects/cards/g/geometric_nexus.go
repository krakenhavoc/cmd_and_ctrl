package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Geometric Nexus — Artifact {2} (Commander 2021, #5504):
//
//	"Whenever a player casts an instant or sorcery spell, put a number of
//	 charge counters on this artifact equal to that spell's mana value.
//	 {6}, {T}, Remove all charge counters from this artifact: Create a 0/0
//	 green and blue Fractal creature token. Put X +1/+1 counters on it,
//	 where X is the number of charge counters removed this way."
//
// Any player's cast triggers it. The count is the spell's mana value as the
// trigger resolves, its announced X included while it is on the stack (CR
// 202.3e). The ability removes every charge counter as a cost (a cost of
// removing none is still payable, CR 118.3), and X reaches the effect as the
// number removed (Catti-brie's reading); the Fractal is the create-then-put
// shape.
//
// No simplification.
func init() {
	const key = "geometric-nexus/fractal"
	effects.Register(effects.Spec{
		OracleID:     "935bcf49-ec8b-4ad6-b12d-77dc7cccf1d4",
		Name:         "Geometric Nexus",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnyPlayerCastAnInstantOrSorcery, "Geometric Nexus — put charge counters on it equal to the spell's mana value",
				func(g *game.Game, item *game.StackItem) error {
					n := effects.TriggeringSpellManaValue(g, item)
					if n <= 0 {
						return nil
					}
					return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterCharge, N: n}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}, {T}, Remove all charge counters from this artifact: Create a 0/0 green and blue Fractal creature token. Put X +1/+1 counters on it, where X is the number of charge counters removed this way.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1, CreatureTokens: 1},
			Cost:    effects.Plus(effects.ManaCost("{6}"), effects.TapCost(), effects.RemoveAllCountersFromThis(game.CounterCharge)),
			Effect:  effects.SocB8FractalWithCounters(key, func(ctx *effects.Context) int { return ctx.CountersRemoved() }),
		}},
	})
}
