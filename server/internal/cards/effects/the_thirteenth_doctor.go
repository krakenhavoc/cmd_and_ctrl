package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Thirteenth Doctor — Legendary Creature — Time Lord Doctor {1}{G}{U}, 2/2
// (Doctor Who):
//
//	"Paradox — Whenever you cast a spell from anywhere other than your hand, put a
//	 +1/+1 counter on target creature.
//	 Team TARDIS — At the beginning of your end step, untap each creature you
//	 control with a counter on it."
//
// The target is any creature, chosen as the trigger goes on the stack and
// re-checked as it resolves. The end-step untap reads each creature's counters (of
// any kind) as the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3778e06c-62c7-4b8f-90ff-b2fc2d2e213c",
		Name:         "The Thirteenth Doctor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventCast, gapB87YouCastFromOutsideYourHand,
					"Paradox — put a +1/+1 counter on target creature",
					b36CounterOnChosenAnimal),
				TargetCreature("target creature")),
			AtYourEndStep("Team TARDIS — untap each creature you control with a counter on it", gapB87UntapCreaturesWithACounter),
		},
	})
}
