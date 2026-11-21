package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghost-Lit Redeemer — Creature — Spirit {W}, 1/1 (Saviors of Kamigawa,
// #5601):
//
//	"{W}, {T}: You gain 2 life.
//	 Channel — {1}{W}, Discard this card: You gain 4 life."
//
// The channel row is Boseiju, Who Endures's shape: an activated ability
// that works only from the hand, with the discard as part of its cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6c22cbe-61e0-4d66-a78d-7016e84c303d",
		Name:         "Ghost-Lit Redeemer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{W}, {T}: You gain 2 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
				Effect:  effects.Do(effects.GainLife{Amount: 2}),
			},
			{
				Label:   "Channel — {1}{W}, Discard this card: You gain 4 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.DiscardThis()),
				Zones:   []game.ZoneKind{game.ZoneHand},
				Effect:  effects.Do(effects.GainLife{Amount: 4}),
			},
		},
	})
}
