package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Weather Maker — Artifact {3}:
//
//	"Landfall — Whenever a land you control enters, put a charge
//	 counter on this artifact.
//	 {T}: Add one mana of any color.
//	 {T}, Remove two charge counters from this artifact: Add {C}{C}.
//	 {T}, Remove three charge counters from this artifact: It deals 3
//	 damage to any target."
//
// Landfall banks a charge counter (Door of Destinies' body). The two
// counter costs are paid at announce (#625), so a response can't spend
// the same counters twice; the {C}{C} row is Pentad Prism's counter
// cost on a mana ability, with the tap added. "It deals 3 damage" is
// the artifact as the source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4097fea2-ae34-4b99-ab7e-a276d608b489",
		Name:         "Weather Maker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Weather Maker — landfall: put a charge counter on this artifact", effects.PutChargeCounterOnThis),
		},
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color",
			},
			{
				Cost: effects.ManaAbilityCost{
					Tap:            true,
					RemoveCounters: effects.RemoveCountersFromThis(game.CounterCharge, 2).RemoveCounters,
				},
				Produced: "{C}{C}",
				Label:    "Remove two charge counters from this artifact: Add {C}{C}",
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Remove three charge counters from this artifact: It deals 3 damage to any target.",
			Cost:    effects.Plus(effects.TapCost(), effects.RemoveCountersFromThis(game.CounterCharge, 3)),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
			Effect:  effects.DealDamageToTheTarget(3),
		}},
	})
}
