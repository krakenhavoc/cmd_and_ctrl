package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deepglow Skate — Creature — Fish {4}{U}, 3/3 (EDHREC rank 1838):
//
//	"When this creature enters, double the number of each kind of
//	 counter on any number of target permanents."
//
// The counters deck's ultimate-enabler: every planeswalker at
// loyalty 6 goes to 12. "Any number of target permanents" is the
// Appa shape — a multi-target trigger, Min 0 and Max unbounded,
// answered through the ordinary multi-pick — and each target still
// legal at resolution (CR 608.2b per slot) gets as many counters of
// each kind as it already has, the counts snapshotted before the
// first lands so a Doubling Season on the first kind cannot change
// the second. "Each kind" is every kind — loyalty, charge, -1/-1 and
// all — as printed.
//
// Same engine-wide note as Appa: a trigger with no legal target is
// dropped before the prompt (CR 603.3d), where the printed Min-0
// clause would put it on the stack targeting nothing. Unobservable
// while the Skate itself is a permanent on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "debce64a-18bd-42f5-9e85-158c2242e9e9",
		Name:         "Deepglow Skate",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets:   TargetPermanent("any number of target permanents").WithCount(0, 0),
			Key:       "Deepglow Skate — double the counters on the targets",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range item.Targets {
					if t.Kind != game.TargetCard || !g.TargetStillLegalForEffect(item, t) {
						continue
					}
					if err := b17DoubleCountersOn(ctx, t.ID); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
