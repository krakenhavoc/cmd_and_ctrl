package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Willowdusk, Essence Seer — Legendary Creature — Dryad Druid {1}{B}{G},
// 3/3 (Commander 2021, #5504):
//
//	"{1}, {T}: Choose another target creature. Put a number of +1/+1
//	 counters on it equal to the amount of life you gained this turn or
//	 the amount of life you lost this turn, whichever is greater. Activate
//	 only as a sorcery."
//
// "Another" is object identity, so Willowdusk can't pick itself. The
// number is the larger of the two turn tallies, read as the ability
// resolves: life GAINED is the turn's total, not the net change, and life
// LOST counts every loss (damage included); a turn with neither puts no
// counters. A target gone by resolution makes the ability do nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "94723d61-3e29-4180-91bc-df37c4086a66",
		Name:         "Willowdusk, Essence Seer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}, {T}: Choose another target creature. Put a number of +1/+1 counters on it equal to the amount of life you gained this turn or the amount of life you lost this turn, whichever is greater. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets:      effects.Another(effects.TargetCreature("another target creature")),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				n := max(effects.B15LifeGainedThisTurn(g, ctx.Controller()), effects.B18LifeLostThisTurn(g, ctx.Controller()))
				if n <= 0 {
					return nil
				}
				return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
			},
		}},
	})
}
