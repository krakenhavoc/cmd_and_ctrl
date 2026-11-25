package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Throne Warden — Creature — Human Soldier {1}{W}, 2/2 (Conspiracy: Take
// the Crown, S82 older sets batch 272, #5540):
//
//	"At the beginning of your end step, if you're the monarch, put a +1/+1
//	 counter on this creature."
//
// Faramir's end-step shape: an intervening "if" (CR 603.4) on the monarch
// (CondMonarch), checked as the step begins and again as the ability
// resolves, so a Warden whose controller loses the crown in response gets
// no counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b50c1dd3-0495-427f-b5a9-9ac44a4554d1",
		Name:         "Throne Warden",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondMonarch{},
				effects.AtYourEndStep("Throne Warden — put a +1/+1 counter on this creature",
					effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}
