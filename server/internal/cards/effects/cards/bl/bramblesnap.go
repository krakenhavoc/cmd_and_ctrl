package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bramblesnap — Creature — Elemental {1}{G}, 1/1 (Rise of the Eldrazi, S82
// OLD batch 483, #5581):
//
//	"Trample
//	 Tap an untapped creature you control: This creature gets +1/+1 until end
//	 of turn."
//
// Mothdust Changeling's cost: "tap an untapped creature you control" is not
// the {T} symbol, so summoning sickness does not stop the creature tapped
// (the Bramblesnap itself included).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "312ce96b-6abe-4f72-a205-9b27c885a5fe",
		Name:            "Bramblesnap",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap an untapped creature you control: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.TapAnUntapped("an untapped creature you control", effects.Creature()),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Bramblesnap — +1/+1 until end of turn"),
		}},
	})
}
