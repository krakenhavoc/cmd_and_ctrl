package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcbound Javelineer — Artifact Creature — Soldier {W}, 0/0 (Modern
// Horizons 2, #5501):
//
//	"{T}, Remove X +1/+1 counters from this creature: It deals X damage to
//	 target attacking or blocking creature.
//	 Modular 1 (This creature enters with a +1/+1 counter on it. When it
//	 dies, you may put its +1/+1 counters on target artifact creature.)"
//
// X is the number of counters removed as the cost is paid (the Infused
// Arrows shape, CR 602.2b); the Javelineer deals the damage as it last
// existed if it has left by resolution (CR 608.2h). Modular is the
// engine's keyword (CR 702.43a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "48e3f6dd-68ab-420a-ad4d-fe7ee39486d3",
		Name:            "Arcbound Javelineer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"modular 1"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Remove X +1/+1 counters from this creature: It deals X damage to target attacking or blocking creature.",
			Purpose: effects.ForTargets(effects.DamageXToTarget(0)),
			Cost:    effects.Plus(effects.TapCost(), effects.RemoveCountersXFromThis(game.CounterPlusOne, 0)),
			Targets: effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := ctx.CountersRemoved()
				if x <= 0 || len(item.Targets) == 0 {
					return nil
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: x}.Apply(ctx)
			},
		}},
	})
}
