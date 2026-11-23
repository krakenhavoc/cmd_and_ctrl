package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "898da62e-d29a-4ddc-be69-2ddec54614f5",
		Name:         "Cackling Prowler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Cackling Prowler — ward {2}"),
			effects.IfVentureCond(effects.CondACreatureDiedThisTurn,
				effects.AtYourEndStep("Cackling Prowler — morbid: put a +1/+1 counter on this creature",
					effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}
