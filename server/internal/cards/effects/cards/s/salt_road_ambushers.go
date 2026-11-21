package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Salt Road Ambushers — Creature — Dog Warrior {3}{G}, 3/3 (Murders at
// Karlov Manor Commander):
//
//	"Whenever another permanent you control is turned face up, if it's a
//	 creature, put two +1/+1 counters on it.
//	 Megamorph {3}{G}{G}"
//
// An intervening "if" (CR 603.4): it triggers only for a creature, and
// the counters go on at resolution only if it still is one and is still
// the same object on the battlefield (a permanent that left gets
// nothing; last-known information is only read). The permanent is face
// up by the time the event is emitted (CR 708.8), so its card types are
// the real ones. The Ambushers face down has no abilities (CR 708.2a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "2373e035-25f7-473a-9386-2093f77d3f94",
		Name:             "Salt Road Ambushers",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Megamorph("{3}{G}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventTurnedFaceUp, effects.Cmdr2S36AnotherCreatureYouControlTurnedFaceUp,
				"Salt Road Ambushers — put two +1/+1 counters on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					info, ok := ctx.TriggeringPermanent()
					id := ctx.Trigger().Event.CardID
					c, found := g.LookupCardForEffect(id)
					if !ok || info.Left || !found || !c.IsCreature() {
						return nil
					}
					return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
				}),
		},
	})
}
