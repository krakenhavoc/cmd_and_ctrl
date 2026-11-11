package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "2373e035-25f7-473a-9386-2093f77d3f94",
		Name:             "Salt Road Ambushers",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Megamorph("{3}{G}{G}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventTurnedFaceUp, cmdr2S36AnotherCreatureYouControlTurnedFaceUp,
				"Salt Road Ambushers — put two +1/+1 counters on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.TriggeringPermanent()
					id := ctx.Trigger().Event.CardID
					c, found := g.LookupCardForEffect(id)
					if !ok || info.Left || !found || !c.IsCreature() {
						return nil
					}
					return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
				}),
		},
	})
}
