package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ride the Avalanche — Instant {G}{U} (Adventures in the Forgotten
// Realms promo):
//
//	"The next spell you cast this turn can be cast as though it had
//	 flash. When you next cast a spell this turn, put X +1/+1 counters on
//	 up to one target creature, where X is the mana value of that spell."
//
// Two things happen as it resolves. The first sentence is a one-use
// promise about the next spell you cast this turn (Quicken's shape,
// GrantNextSpellPromise with Flash). The second creates a delayed
// triggered ability (CR 603.7) that waits for the next spell you cast
// this turn, whichever it is, fires once (CR 603.7b), and ends with the
// turn. When it fires it puts itself on the stack above that spell and
// chooses "up to one target creature" then (CR 603.3d, a
// TargetedDelayedBody, ADR 0187). X is that spell's mana value, from
// last-known information if it has left the stack first (CR 608.2h), with
// an {X} in its cost counted as the value chosen (CR 202.3e).
//
// The spell that cast Ride the Avalanche itself is not "next": its cast
// happened before this resolved.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d49c94fa-712a-47c0-b571-4db7e064a590",
		Name:         "Ride the Avalanche",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.GrantNextSpellPromise{From: "Ride the Avalanche", Promise: game.NextSpellPromise{
					Flash: true,
					Text:  "The next spell you cast this turn can be cast as though it had flash.",
				}}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.WhenYouNextCast(
					"Ride the Avalanche — put X +1/+1 counters on up to one target creature, where X is that spell's mana value",
					game.CastFilter{}, rideTheAvalancheCounters).Apply(ctx)
			}),
		),
	})
}

// rideTheAvalancheCounters is the delayed trigger's body: X +1/+1
// counters on the chosen creature, X being the cast spell's mana value
// (the item's payload). An on-disk key.
var rideTheAvalancheCounters = game.TargetedDelayedBody("ride-the-avalanche/counters",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		spells := ctx.PayloadCards()
		if len(spells) == 0 {
			return nil
		}
		x, ok := g.LastKnownSpellManaValueForEffect(spells[0])
		if !ok || x <= 0 {
			return nil
		}
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: x}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("up to one target creature").WithCount(0, 1)
	}))
