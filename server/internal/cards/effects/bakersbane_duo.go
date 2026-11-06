package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:       "ecd9a0a4-1dfa-4d91-8339-062e2be4e5d9",
		Name:           "Bakersbane Duo",
		ReadsManaSpent: game.ReadsYourCasts,
		Purpose:        game.Purpose{Tokens: 1},
		Completeness:   CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Bakersbane Duo — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
			WheneverYouExpend(4, "Bakersbane Duo — this creature gets +1/+1 until end of turn",
				thisCreatureUntilEOT("Bakersbane Duo — +1/+1", 1, 1)),
		},
	})
}
