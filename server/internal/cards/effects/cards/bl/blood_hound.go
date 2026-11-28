package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blood Hound — Creature — Dog {2}{R}, 1/1 (Mercadian Masques, S82 OLD batch
// 698, #5619):
//
//	"Whenever you're dealt damage, you may put that many +1/+1 counters on this
//	 creature.
//	 At the beginning of your end step, remove all +1/+1 counters from this
//	 creature."
//
// Darien's trigger: "dealt damage", not life loss (CR 119.3), for the amount
// actually applied, one trigger and one question per damage event, and the
// counters go on only if the Hound is still on the battlefield. The end-step
// trigger takes every +1/+1 counter off it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ace532a3-fa6e-479d-b169-1fb24c3d0f8c",
		Name:         "Blood Hound",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.S82OldB698YoureDealtDamageMayPutCounters("Blood Hound"),
			effects.AtYourEndStep("Blood Hound — remove all +1/+1 counters from it", effects.S82OldB698RemoveAllPlusOneCountersFromThis),
		},
	})
}
