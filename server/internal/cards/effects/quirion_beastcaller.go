package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quirion Beastcaller — Creature — Dryad Warrior {1}{G}, 2/2 (Dominaria
// United):
//
//	"Whenever you cast a creature spell, put a +1/+1 counter on this
//	 creature.
//	 When this creature dies, distribute X +1/+1 counters among any
//	 number of target creatures you control, where X is the number of
//	 +1/+1 counters on this creature."
//
// The death trigger divides the counters it had as it last existed
// (DivideSourcePlusOneCounters, CR 603.10), the way Orca, Siege Demon
// divides its power: the division is announced as the trigger goes on the
// stack, each target getting at least one, and a target gone by
// resolution loses its share (PutDividedCounters, CR 608.2b). With no
// counters there is nothing to distribute and no target to choose.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d777e428-cc50-40b1-a241-41576d2575d1",
		Name:         "Quirion Beastcaller",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Creature(), "Quirion Beastcaller — put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return cardDied(ev, source)
				},
				Targets: TargetCreature("any number of target creatures you control", YouControl()).WithCount(0, 0).Dividing(DivideBy(DivideSourcePlusOneCounters)),
				Key:     "Quirion Beastcaller — distribute X +1/+1 counters among any number of target creatures you control, where X is the number of +1/+1 counters on it",
				Effect: func(g *game.Game, item *game.StackItem) error {
					return PutDividedCounters(NewContext(g, item), game.CounterPlusOne)
				},
			},
		},
	})
}
