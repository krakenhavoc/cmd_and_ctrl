package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hell-Bent Raider — Creature — Human Barbarian {1}{R}{R}, 2/2:
//
//	"First strike, haste
//	 Discard a card at random: This creature gains protection from white
//	 until end of turn."
//
// ADR 0109 §7, owner decision 3: "Discard a card at random" is a cost
// the activator chooses nothing for (CR 701.9b). The engine draws the
// card from the hand the rest of the cost leaves, paid after every other
// cost (CR 601.2h), and an empty hand can't pay it (CR 118.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ee8ab29b-0749-463a-abbd-eb0b9013aec8",
		Name:            "Hell-Bent Raider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a card at random: This creature gains protection from white until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.DiscardAtRandom(1, "a card at random"),
			Effect:  effects.ThisCreatureUntilEOT("Hell-Bent Raider — protection from white", 0, 0, "protection from white"),
		}},
	})
}
