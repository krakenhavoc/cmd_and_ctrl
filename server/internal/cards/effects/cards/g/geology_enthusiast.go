package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Geology Enthusiast — Creature — Human Artificer {3}{U}{U}, 2/5 (The
// Brothers' War, #5482):
//
//	"At the beginning of your end step, create a tapped Powerstone token.
//	 {6}: Draw a card and put a +1/+1 counter on this creature."
//
// The end-step trigger is AtYourEndStep. The activated ability is
// She-Hulk's draw-and-counter shape with a mana cost only: the draw
// happens whether or not the creature is still there to take the
// counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c7c1352a-0332-4495-a384-8d6fed01f91b",
		Name:         "Geology Enthusiast",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Geology Enthusiast — create a tapped Powerstone token",
				effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.PowerstoneToken()).EntersTapped(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}: Draw a card and put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{6}"),
			Effect:  effects.Do(effects.DrawCards{N: 1}, effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
