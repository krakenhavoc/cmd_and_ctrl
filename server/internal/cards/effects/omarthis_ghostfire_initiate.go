package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Omarthis, Ghostfire Initiate — Legendary Creature — Spirit Snake {X}{X},
// 0/0 (Commander Masters):
//
//	"Omarthis enters with X +1/+1 counters on it.
//	 Whenever you put one or more +1/+1 counters on another colorless
//	 creature, you may put a +1/+1 counter on Omarthis.
//	 When Omarthis dies, manifest a number of cards from the top of your
//	 library equal to the number of counters on it."
//
// Endless One's entry counters (XMatters, the announced X). The counter
// trigger is the shared "whenever you put counters of a kind on a creature"
// watcher, once per batch for each creature (CR 603.2c), narrowed to another
// creature that is colorless as the counters land, with the "you may" asked
// as the trigger goes on the stack. The death trigger reads the counters
// Omarthis had as it died (CR 603.10a), of every kind, and manifests that
// many cards one at a time (CR 701.40e); none if it had none.
//
// No simplification.
func init() {
	counters := Optional(WheneverYouPutCountersOfKindOnACreature(game.CounterPlusOne,
		func(_ *game.Game, source *game.Card, c game.Card) bool {
			return c.InstanceID != source.InstanceID && c.IsCreature() && c.IsColorless()
		},
		"Omarthis — you may put a +1/+1 counter on Omarthis",
		gapB99CounterOnThis), "Omarthis — put a +1/+1 counter on Omarthis?")
	Register(Spec{
		OracleID:                   "50676eba-a55a-4749-bb07-74a30e015781",
		Name:                       "Omarthis, Ghostfire Initiate",
		Completeness:               CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		Triggered: []game.TriggeredAbility{
			counters,
			WhenThisDies("Omarthis — manifest a number of cards from the top of your library equal to the number of counters on it",
				omarthisManifest),
		},
	})
}

func omarthisManifest(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	n := 0
	for _, c := range info.Counters {
		n += c
	}
	return Manifest{Player: ctx.Controller(), N: n}.Apply(ctx)
}
