package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vivien's Grizzly — Creature — Bear Spirit {2}{G}, 2/3 (War of the Spark, #5520):
//
//	"{3}{G}: Look at the top card of your library. If it's a creature or
//	 planeswalker card, you may reveal it and put it into your hand. If you
//	 don't put the card into your hand, put it on the bottom of your library."
//
// Domri Rade's +1 take (look, reveal, hand), with a creature-or-planeswalker
// filter, and the card that stays put on the bottom of the library. A top card
// of another type goes to the bottom too, as the second sentence says.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6dbd0764-4bfe-4a58-905d-aeee70ceb798",
		Name:         "Vivien's Grizzly",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}: Look at the top card of your library. If it's a creature or planeswalker card, you may reveal it and put it into your hand. If you don't put the card into your hand, put it on the bottom of your library.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{3}{G}"),
			Effect:  effects.S82OldB172VivienLook,
		}},
	})
}
