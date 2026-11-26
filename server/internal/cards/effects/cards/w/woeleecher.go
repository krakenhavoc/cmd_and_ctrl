package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Woeleecher — Creature — Elemental {5}{W}, 3/5 (Shadowmoor, S82 OLD b529,
// #5591):
//
//	"{W}, {T}: Remove a -1/-1 counter from target creature. If you do, you
//	 gain 2 life."
//
// The life is the "if you do": it comes only when a -1/-1 counter was really
// on the creature as the ability resolves (CR 608.2b: a target that has gone
// does nothing). The counter comes off through the effect-side adder at -1 so
// the layer pass sees it go.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52e73390-75e4-45a1-8dbe-96e7cceb0a14",
		Name:         "Woeleecher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Remove a -1/-1 counter from target creature. If you do, you gain 2 life.",
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					c, ok := g.LookupCardForEffect(t.ID)
					if !ok || c.Counters[game.CounterMinusOne] <= 0 {
						return nil
					}
					if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterMinusOne, N: -1}).Apply(ctx); err != nil {
						return err
					}
					return effects.GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
