package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "3778e06c-62c7-4b8f-90ff-b2fc2d2e213c",
		Name:         "The Thirteenth Doctor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventCast, effects.GapB87YouCastFromOutsideYourHand,
					"Paradox — put a +1/+1 counter on target creature",
					effects.B36CounterOnChosenAnimal),
				effects.TargetCreature("target creature")),
			effects.AtYourEndStep("Team TARDIS — untap each creature you control with a counter on it", effects.GapB87UntapCreaturesWithACounter),
		},
	})
}
