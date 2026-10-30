package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Extremis Elite — Creature — Human Mercenary Villain {1}{R}, 2/2 (Marvel
// Super Heroes Commander, #4455):
//
//	"Power-up — {4}{R}: Put two +1/+1 counters on this creature. It deals 1
//	 damage to any target. (Activate each power-up ability only once. Reduce
//	 the cost by its mana cost if it entered this turn.)"
//
// Power-up is the engine's (CR 702.193). The target is announced with the
// activation; the counters go on first, then the damage, dealt by the
// creature (a creature that left in response deals none: it is no longer on
// the battlefield to put counters on, and the damage is still dealt from
// last-known information).
//
// No simplification.
func init() {
	ab := PowerUp("{4}{R}: Put two +1/+1 counters on this creature. It deals 1 damage to any target.",
		ManaCost("{4}{R}"), func(g *game.Game, item *game.StackItem) error {
			if err := plusOneCountersOnThis(2)(g, item); err != nil {
				return err
			}
			return sourceDealsDamageToEachLegalTarget(1)(g, item)
		})
	ab.Targets = TargetAny()
	ab.Purpose = ForTargets(DamageToTarget(0, 1))
	Register(Spec{
		OracleID:     "b058ee2e-ddb9-4f49-bbe0-16b72d54d4da",
		Name:         "Extremis Elite",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{ab},
	})
}
