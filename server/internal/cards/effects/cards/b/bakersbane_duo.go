package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bakersbane Duo — Creature — Squirrel Raccoon {1}{G}, 2/2 (Bloomburrow,
// #4467):
//
//	"When this creature enters, create a Food token.
//	 Whenever you expend 4, this creature gets +1/+1 until end of turn.
//	 (You expend 4 as you spend your fourth total mana to cast spells
//	 during a turn.)"
//
// The Food is the catalog Food (ADR 0083). Expend is WheneverYouExpend
// (expend.go), Trailtracker Scout's tally of mana spent on spells this
// turn; the pump is thisCreatureUntilEOT, which does nothing if the Duo
// has left by the time the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "ecd9a0a4-1dfa-4d91-8339-062e2be4e5d9",
		Name:           "Bakersbane Duo",
		ReadsManaSpent: game.ReadsYourCasts,
		Purpose:        game.Purpose{Tokens: 1},
		Completeness:   effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Bakersbane Duo — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
			effects.WheneverYouExpend(4, "Bakersbane Duo — this creature gets +1/+1 until end of turn",
				effects.ThisCreatureUntilEOT("Bakersbane Duo — +1/+1", 1, 1)),
		},
	})
}
