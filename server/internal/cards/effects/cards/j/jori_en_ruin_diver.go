package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jori En, Ruin Diver — Legendary Creature — Merfolk Wizard {1}{U}{R}, 2/3
// (Oath of the Gatewatch, #5545):
//
//	"Whenever you cast your second spell each turn, draw a card."
//
// The trigger is YouCastYourSecondSpellEachTurn, as Monk of the Open Hand
// reads it: the cast path bumps the per-turn tally before it emits the cast
// event, so a total of two is the second spell, and a spell cast before Jori En
// entered still counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4cd9b48e-5774-4a1c-857f-9f4a92db9b41",
		Name:         "Jori En, Ruin Diver",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn,
				"Jori En, Ruin Diver — draw a card",
				effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
