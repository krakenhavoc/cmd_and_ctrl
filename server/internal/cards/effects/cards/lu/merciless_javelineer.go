package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Merciless Javelineer — Creature — Minotaur Warrior {2}{B}{R}, 4/2 (Amonkhet, S82 OLD batch 242, #5535):
//
//	"{2}, Discard a card: Put a -1/-1 counter on target creature. That creature
//	 can't block this turn."
//
// The discard is part of the cost (CR 118.3); the target is chosen on
// activation. The counter and the "can't block" both apply to that creature as
// the ability resolves, and neither happens if it is no longer a legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd2b59b4-281e-4b51-8998-9b20a846ddbf",
		Name:         "Merciless Javelineer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Discard a card: Put a -1/-1 counter on target creature. That creature can't block this turn.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.DiscardACard()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterMinusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
					if err := (effects.RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock, Label: "Merciless Javelineer — can't block"}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
