package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seven-Tail Mentor — Creature — Fox Samurai {3}{W}, 2/3 (Kamigawa: Neon
// Dynasty, S82):
//
//	"When this creature enters or dies, put a +1/+1 counter on target
//	 creature or Vehicle you control."
//
// One ability with two trigger conditions (Goblin Marshal). The target is a
// permanent that is a creature or a Vehicle, so a Vehicle that is not
// crewed can take the counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "301b8265-9e17-48c3-942d-b7243112b394",
		Name:         "Seven-Tail Mentor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEntersOrDies("Seven-Tail Mentor — put a +1/+1 counter on target creature or Vehicle you control", PlusOneCounterOnChosenTarget),
				TargetPermanent("target creature or Vehicle you control", Or(Creature(), HasSubtype("Vehicle")), YouControl())),
		},
	})
}
