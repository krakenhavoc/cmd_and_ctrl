package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Impelled Giant — {4}{R}{R} Creature — Giant Warrior, 3/3:
//
//	"Trample
//	 Tap an untapped red creature you control other than this creature: This
//	 creature gets +X/+0 until end of turn, where X is the power of the creature
//	 tapped this way."
//
// The tap is not the {T} symbol (CR 302.6): a creature that arrived this turn
// may pay it.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "eb7c9442-7bfa-4bcf-a394-b5de187f224a",
		Name:            "Impelled Giant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap an untapped red creature you control other than this creature: This creature gets +X/+0 until end of turn, where X is the power of the creature tapped this way.",
			Cost:    effects.TapAnotherUntapped("untapped red creature you control other than this creature", effects.Creature(), effects.OfColor("R")),
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect:  effects.S82OldB523PumpByTappedPower,
		}},
	})
}
