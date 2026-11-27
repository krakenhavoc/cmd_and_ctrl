package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mercadian Lift — Artifact {2} (Mercadian Masques, S82 OLD batch 698, #5619):
//
//	"{1}, {T}: Put a winch counter on this artifact.
//	 {T}, Remove X winch counters from this artifact: You may put a creature
//	 card with mana value X from your hand onto the battlefield."
//
// Brain in a Jar's X: the count is announced and removed as a cost, and the
// effect reads it back as ctx.CountersRemoved(). The put is a real "may": only
// creature cards with exactly that mana value are offered, and declining is
// an answer (CR 110.4: it is put, not cast).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88c3bb73-c765-4e1c-b4b9-97eaab02182f",
		Name:         "Mercadian Lift",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}, {T}: Put a winch counter on this artifact.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
				Effect:  effects.Do(effects.CounterOnThis{Kind: "winch", N: 1}),
			},
			{
				Label:   "{T}, Remove X winch counters from this artifact: You may put a creature card with mana value X from your hand onto the battlefield.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.TapCost(), effects.RemoveCountersXFromThis("winch", 0)),
				Effect:  effects.S82OldB698PutCreatureOfManaValueXFromHand("Mercadian Lift"),
			},
		},
	})
}
