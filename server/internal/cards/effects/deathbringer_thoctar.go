package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deathbringer Thoctar — Creature — Zombie Beast {4}{B}{R}, 3/3:
//
//	"Whenever another creature dies, you may put a +1/+1 counter on
//	 this creature.
//	 Remove a +1/+1 counter from this creature: It deals 1 damage to
//	 any target."
//
// The trigger is Orca's "another creature dies" with a "you may" gate
// (OptionalPrompt, CR 603.5) in front of the same counter body. The
// ability's cost is a counter removal paid at announcement
// (RemoveCountersFromThis, CR 602.2b), so it can be activated in
// response to a removal spell by spending counters the creature is
// about to lose, and each activation is one counter and one damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2500a811-2435-4915-ac83-9bfe2887621a",
		Name:         "Deathbringer Thoctar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b15AnotherCreatureDied(ev, source, g)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Deathbringer Thoctar — put a +1/+1 counter on it?"},
			Key:            "Deathbringer Thoctar — you may put a +1/+1 counter on it",
			Effect:         putCounterOnSelf,
		}},
		Activated: []ActivatedAbility{{
			Label:   "Remove a +1/+1 counter from this creature: It deals 1 damage to any target.",
			Cost:    RemoveCountersFromThis(game.CounterPlusOne, 1),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  DealDamageToTheTarget(1),
		}},
	})
}
