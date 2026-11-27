package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyshroud Vampire — Creature — Vampire {3}{B}{B}, 3/3 (Tempest, S82 OLD
// b755, #5628):
//
//	"Flying
//	 Discard a creature card: This creature gets +2/+2 until end of turn."
//
// Vampire Hounds' "Discard a creature card" cost and +2/+2 self pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "58214e60-3258-4159-bf5a-b4069d80186d",
		Name:            "Skyshroud Vampire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a creature card: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
			Cost:    effects.DiscardCardsMatching(1, "a creature card", effects.MatchCreature),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 2, "Skyshroud Vampire — +2/+2 until end of turn"),
		}},
	})
}
