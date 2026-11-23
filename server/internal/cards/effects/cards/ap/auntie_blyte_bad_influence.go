package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Auntie Blyte, Bad Influence — Legendary Creature — Devil Advisor {2}{R},
// 2/2 (Jumpstart 2022, #5948):
//
//	"Flying
//	 Whenever a source you control deals damage to you, put that many +1/+1
//	 counters on Auntie Blyte.
//	 {1}{R}, {T}, Remove X +1/+1 counters from Auntie Blyte: It deals X
//	 damage to any target."
//
// One trigger per damage event to you (CR 603.2), read as it was dealt and
// carried on the item: damage prevented is not dealt, and a Blyte that left
// before the trigger resolved has nothing to put counters on. Michiko
// Konda's watch, over the damage to its controller from its controller's own
// sources.
//
// X is announced as the cost is paid (Infused Arrows' shape) and reaches the
// effect as the counters removed; the damage is dealt by Blyte, as it last
// existed if it is gone. X may be 0 and then nothing is dealt.
//
// No simplification.
func init() {
	const label = "Auntie Blyte, Bad Influence — put that many +1/+1 counters on it"
	effects.Register(effects.Spec{
		OracleID:        "a5e9142e-76bf-40d4-aeaa-ffe197d38f25",
		Name:            "Auntie Blyte, Bad Influence",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: effects.GapB154SourceYouControlDealtDamageToYou,
			Key:       label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Amount = ev.Amount
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Params.Amount <= 0 || !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
					return nil
				}
				return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: item.Params.Amount}.Apply(effects.NewContext(g, item))
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, {T}, Remove X +1/+1 counters from Auntie Blyte: It deals X damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.TapCost(), effects.RemoveCountersXFromThis(game.CounterPlusOne, 0)),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := ctx.CountersRemoved()
				if x <= 0 {
					return nil
				}
				return effects.DamageToFirstTargetStep(x)(item, ctx)
			},
		}},
	})
}
