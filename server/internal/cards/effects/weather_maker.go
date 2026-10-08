package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "4097fea2-ae34-4b99-ab7e-a276d608b489",
		Name:         "Weather Maker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Weather Maker — landfall: put a charge counter on this artifact", putChargeCounterOnThis),
		},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color",
			},
			{
				Cost: ManaAbilityCost{
					Tap:            true,
					RemoveCounters: RemoveCountersFromThis(game.CounterCharge, 2).RemoveCounters,
				},
				Produced: "{C}{C}",
				Label:    "Remove two charge counters from this artifact: Add {C}{C}",
			},
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Remove three charge counters from this artifact: It deals 3 damage to any target.",
			Cost:    Plus(TapCost(), RemoveCountersFromThis(game.CounterCharge, 3)),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 3)),
			Effect:  DealDamageToTheTarget(3),
		}},
	})
}
