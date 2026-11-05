package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// McCoy, Not a Bricklayer — Legendary Creature — Human Doctor {3}{G}{W},
// 0/4 (Star Trek Commander, #4974):
//
//	"Whenever McCoy or another Doctor you control enters, put a +1/+1
//	 counter on each creature you control.
//	 Creatures you control with counters on them have lifelink."
//
// The trigger fires for McCoy himself and for every other Doctor, token
// or not, that enters under his controller's control (read through the
// layered view, so a changeling is a Doctor). The counters go on each
// creature you control as the trigger resolves, one placement each. The
// static is the shared layer-6 grant over creatures you control with any
// kind of counter, McCoy included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c71c7e0c-64d8-486a-9e66-1b343868eb0a",
		Name:         "McCoy, Not a Bricklayer",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CreaturesYouControlWithCountersHave("", false, "lifelink"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, trekB40SelfOrAnotherOfSubtypeYouControlEntered("Doctor"),
				"McCoy, Not a Bricklayer — put a +1/+1 counter on each creature you control",
				b11PutCounterOnEachCreatureYouControl),
		},
	})
}
