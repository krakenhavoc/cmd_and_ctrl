package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Necroplasm — Creature — Ooze {1}{B}{B}, 1/1 (Commander 2016, S82 older sets,
// #5537):
//
//	"At the beginning of your upkeep, put a +1/+1 counter on this creature.
//	 At the beginning of your end step, destroy each creature with mana value
//	 equal to the number of +1/+1 counters on this creature.
//	 Dredge 2"
//
// The counters are counted as the end step trigger resolves, from the
// Necroplasm's last-known information if it has gone (CR 608.2h); each
// creature's mana value is read through the layered view (a token is 0, CR
// 202.3). Necroplasm destroys itself when its mana value (3) equals its own
// counters. Dredge is the ordinary graveyard draw replacement (CR 702.52).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa1c40fb-c230-4e7f-99b1-4e2e8c39a168",
		Name:         "Necroplasm",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Replacements: []game.ReplacementEffect{effects.Dredge(2)},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Necroplasm — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			effects.AtYourEndStep("Necroplasm — destroy each creature with mana value equal to the number of +1/+1 counters on this creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					info, ok := ctx.SourcePermanent()
					if !ok {
						return nil
					}
					counters := info.Counters[game.CounterPlusOne]
					return effects.DestroyAllMatching{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.IsCreature() && c.ManaValue() == counters
					}}.Apply(ctx.AsGroupMember())
				}),
		},
	})
}
