package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stone Kavu — Creature — Kavu {4}{G}, 3/3 (Planeshift, S82 OLD #5615):
//
//	"{R}: This creature gets +1/+0 until end of turn.
//	 {W}: This creature gets +0/+1 until end of turn."
//
// Two firebreathing rows through the shared ThisGetsUntilEndOfTurn body; no
// {T}, so each can be activated any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01565f61-a3c8-42ca-99f3-07cfc64034a2",
		Name:         "Stone Kavu",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}: This creature gets +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
				Cost:    effects.ManaCost("{R}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Stone Kavu — +1/+0 until end of turn"),
			},
			{
				Label:   "{W}: This creature gets +0/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Toughness: 1}},
				Cost:    effects.ManaCost("{W}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(0, 1, "Stone Kavu — +0/+1 until end of turn"),
			},
		},
	})
}
