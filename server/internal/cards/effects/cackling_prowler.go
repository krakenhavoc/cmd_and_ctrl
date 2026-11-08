package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cackling Prowler — Creature — Hyena Rogue {3}{G}, 4/3 (Foundations,
// #4465):
//
//	"Ward {2}
//	 Morbid — At the beginning of your end step, if a creature died this
//	 turn, put a +1/+1 counter on this creature."
//
// Ward is the shared Ward trigger (Aboleth Spawn's ward {2}). Morbid is
// Needletooth Pack's end-step trigger with its intervening "if" (Zombie
// Ogre's condACreatureDiedThisTurn): any creature, any controller,
// checked as your end step begins and again as the trigger resolves. The
// counter goes on this creature only while it is still on the
// battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "898da62e-d29a-4ddc-be69-2ddec54614f5",
		Name:         "Cackling Prowler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Cackling Prowler — ward {2}"),
			IfVentureCond(condACreatureDiedThisTurn,
				AtYourEndStep("Cackling Prowler — morbid: put a +1/+1 counter on this creature",
					Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}
