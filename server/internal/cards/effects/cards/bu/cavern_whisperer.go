package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cavern Whisperer — Creature — Nightmare {4}{B}, 4/4 (Ikoria: Lair of
// Behemoths, #5512):
//
//	"Mutate {3}{B} (If you cast this spell for its mutate cost, put it over or
//	 under target non-Human creature you own. They mutate into the creature on
//	 top plus all abilities from under it.)
//	 Menace (This creature can't be blocked except by two or more creatures.)
//	 Whenever this creature mutates, each opponent discards a card."
//
// Mutate is the engine's (ADR 0223); each opponent chooses the card they
// discard (EachOpponentDiscardsOne), and one with an empty hand discards
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "b42147e1-cf90-4c12-8abe-7f2931d0607c",
		Name:             "Cavern Whisperer",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"menace"},
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{3}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventMutate, effects.ThisCreatureMutates, "Cavern Whisperer — each opponent discards a card", effects.EachOpponentDiscardsOne),
		},
	})
}
