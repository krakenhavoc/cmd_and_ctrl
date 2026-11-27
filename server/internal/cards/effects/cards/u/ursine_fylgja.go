package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ursine Fylgja — Creature — Spirit Bear {4}{W}, 3/3 (Coldsnap, #5597):
//
//	"This creature enters with four healing counters on it.
//	 Remove a healing counter from this creature: Prevent the next 1 damage
//	 that would be dealt to this creature this turn.
//	 {2}{W}: Put a healing counter on this creature."
//
// Barkhide Troll's entry counters and removal cost (paid whole, so a response
// cannot spend one twice); the prevention is Wojek Apothecary's charged shield
// (CR 615.7) on this object.
//
// No simplification.
func init() {
	const healing = "healing"
	effects.Register(effects.Spec{
		OracleID:     "4d244787-66d2-4a32-914d-a3e97d393370",
		Name:         "Ursine Fylgja",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.B10EntersWithCounters(healing, 4, "Enters with four healing counters on it.")},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Remove a healing counter from this creature: Prevent the next 1 damage that would be dealt to this creature this turn.",
				Purpose: game.Purpose{Answers: game.AnswerPrevent},
				Cost:    effects.RemoveCountersFromThis(healing, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.PreventNextDamage{Target: item.SourceCardID, Amount: 1, Label: "Ursine Fylgja: prevent the next 1 damage"}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{2}{W}: Put a healing counter on this creature.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{2}{W}"),
				Effect:  effects.Do(effects.CounterOnThis{Kind: healing, N: 1}),
			},
		},
	})
}
