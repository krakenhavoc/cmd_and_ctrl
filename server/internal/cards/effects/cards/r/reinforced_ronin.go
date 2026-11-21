package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reinforced Ronin — Artifact Creature — Human Samurai {R}, 2/2 (Neon
// Dynasty, #5493):
//
//	"Haste
//	 At the beginning of your end step, return this creature to its owner's
//	 hand.
//	 Channel — {1}{R}, Discard this card: Draw a card."
//
// The end-step trigger is Fleeting Effigy's: it returns the Ronin only while
// it is still the permanent the ability came from (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5e891927-b6cb-40ae-8270-0e72820fd7f5",
		Name:            "Reinforced Ronin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Reinforced Ronin — return this creature to its owner's hand",
				effects.Do(effects.Instruction(effects.StdB9ReturnThisToHand))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Channel — {1}{R}, Discard this card: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Effect:  effects.NeoB07DrawsCards(1),
		}},
	})
}
