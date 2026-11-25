package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shatterskull Charger — Creature — Giant Warrior {1}{R}{R}, 4/3 (Zendikar
// Rising, S82 OLD #5508):
//
//	"Kicker {2}
//	 Trample, haste
//	 If this creature was kicked, it enters with a +1/+1 counter on it.
//	 At the beginning of your end step, if this creature doesn't have a
//	 +1/+1 counter on it, return it to its owner's hand."
//
// The counter is the kicked entry counter (Gnarlid Colony's shape). The end
// step trigger is an intervening "if" (CR 603.4): it is checked as the step
// begins and again as it resolves, and does nothing for a Charger that is no
// longer the same object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "571f9d80-9c37-474e-9524-400e7a238e2c",
		Name:            "Shatterskull Charger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		OptionalCosts:   []game.AdditionalCost{effects.Kicker("{2}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			effects.CountersPerKick(game.CounterPlusOne, 1),
		},
		Triggered: []game.TriggeredAbility{shatterskullChargerTrigger()},
	})
}

func shatterskullChargerTrigger() game.TriggeredAbility {
	t := effects.InterveningIf(effects.CondBoard{
		Name: "this creature doesn't have a +1/+1 counter on it",
		Test: func(ctx *effects.Context) bool {
			c, ok := ctx.Game.LookupCardForEffect(ctx.Source())
			return ok && c.Counters[game.CounterPlusOne] == 0
		},
	}, effects.AtYourEndStep("Shatterskull Charger — return it to its owner's hand",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			if !effects.OnBattlefield(g, item.SourceCardID) || ctx.IsNewSourceObject(item.SourceCardID) {
				return nil
			}
			return effects.BounceToHand{Target: item.SourceCardID}.Apply(ctx)
		}))
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return source != nil && source.Counters[game.CounterPlusOne] == 0 && (when == nil || when(ev, source, lki, g))
	}
	return t
}
