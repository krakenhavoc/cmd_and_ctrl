package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Walking Ballista — Artifact Creature — Construct {X}{X}, 0/0
// (slice 296-m):
//
//	"This creature enters with X +1/+1 counters on it.
//	 {4}: Put a +1/+1 counter on this creature.
//	 Remove a +1/+1 counter from this creature: It deals 1 damage to
//	 any target."
//
// The X counters are the printed CR 614.1c entry clause and ride the
// CR 614 pipeline (XCounters, #1002), exactly as Benevolent Hydra's
// do. The two activated abilities are a plain mana-cost counter add
// (no target) and a cost-only-counter-removal damage ability
// (RemoveCountersFromThis, the same cost Benevolent Hydra's tap
// ability pays), targeted with TargetAny() for "any target" and
// dealing its damage from the Ballista (CR 609.7a — last-known
// information if it has left in response, since the cost already
// removed the counter and the ability is on the stack independent of
// the source).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "4b515bb0-f275-4400-8032-3173b799ab40",
		Name:                       "Walking Ballista",
		Completeness:               CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		Activated: []ActivatedAbility{
			{
				Label:  "{4}: Put a +1/+1 counter on this creature.",
				Cost:   ManaCost("{4}"),
				Effect: plusOneCountersOnThis(1),
			},
			{
				Label:   "Remove a +1/+1 counter from this creature: It deals 1 damage to any target.",
				Cost:    RemoveCountersFromThis(game.CounterPlusOne, 1),
				Targets: TargetAny(),
				Purpose: ForTargets(DamageToTarget(0, 1)),
				Effect:  b33DamageChosenTargetFromSource(1),
			},
		},
	})
}
