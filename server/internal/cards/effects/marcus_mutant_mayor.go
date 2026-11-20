package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marcus, Mutant Mayor — Legendary Creature — Mutant Advisor {3}{G}{U},
// 4/4 (Fallout, #5948):
//
//	"Vigilance, trample
//	 Whenever a creature you control deals combat damage to a player, draw a
//	 card if that creature has a +1/+1 counter on it. If it doesn't, put a
//	 +1/+1 counter on it."
//
// One trigger per creature and per player it hit. The "if" is in the effect,
// not an intervening one, so it is read as the ability resolves: a creature
// that already has a +1/+1 counter draws a card, one that doesn't gets one
// (CR 608.2c). A creature that has left the battlefield by then has no
// counter to put on and is read from its departure record for the draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c9db8f36-2c72-403e-b2b7-7cdfba5feaa4",
		Name:            "Marcus, Mutant Mayor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, gapB63CreatureYouControlDealtCombatDamageToAPlayer,
				"Marcus, Mutant Mayor — draw a card if that creature has a +1/+1 counter on it; otherwise put a +1/+1 counter on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if item.Trigger == nil {
						return nil
					}
					id := item.Trigger.Event.Source
					counters := 0
					if c, ok := g.LookupCardForEffect(id); ok && onBattlefield(g, id) {
						counters = c.Counters[game.CounterPlusOne]
					} else if info, ok := g.LastKnownPermanentForEffect(id); ok {
						counters = info.Counters[game.CounterPlusOne]
					}
					if counters > 0 {
						return DrawCards{N: 1}.Apply(ctx)
					}
					if !onBattlefield(g, id) {
						return nil
					}
					return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}),
		},
	})
}
