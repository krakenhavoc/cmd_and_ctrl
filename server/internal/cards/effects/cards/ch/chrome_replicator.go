package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chrome Replicator — Artifact Creature — Construct {5}, 4/4 (Core Set 2021,
// #5511):
//
//	"When this creature enters, if you control two or more nonland, nontoken
//	 permanents with the same name as one another, create a 4/4 colorless
//	 Construct artifact creature token."
//
// An intervening "if" (CR 603.4): checked as it enters and again as the
// trigger resolves. The Replicator itself counts. Names are read through the
// layers (a copy has its copied name).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f379fa0b-e9fc-4307-a574-a0c95e5e7f72",
		Name:         "Chrome Replicator",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you control two or more nonland, nontoken permanents with the same name as one another", effects.S82OldB109TwoSameNamePermanents,
				effects.On(game.EventETB, effects.Self, "Chrome Replicator — create a 4/4 colorless Construct artifact creature token",
					effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 colorless Construct artifact"), N: 1}))),
		},
	})
}
