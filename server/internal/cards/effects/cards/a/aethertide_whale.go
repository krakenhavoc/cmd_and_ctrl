package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aethertide Whale — Creature — Whale {4}{U}{U}, 6/4:
//
//	"Flying
//	 When this creature enters, you get six {E} (energy counters).
//	 Pay {E}{E}{E}{E}: Return this creature to its owner's hand."
//
// ADR 0129 PR 1 (#1995). The return is the ability's effect, not a
// cost, so the Whale can be answered with the ability on the stack; a
// Whale that left and came back in the meantime is a new object and
// stays.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "53b2c8f2-db9a-4ab0-a4e8-17385b2fa3bd",
		Name:            "Aethertide Whale",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Energy: 6},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Aethertide Whale", 6),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay {E}{E}{E}{E}: Return this creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.PayEnergy(4),
			Effect:  effects.ReturnThisPermanentToOwnersHand,
		}},
	})
}
