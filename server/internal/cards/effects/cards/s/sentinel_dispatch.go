package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sentinel Dispatch — Conspiracy (Conspiracy, S82 OLD batch 361, #5557):
//
//	"(Start the game with this conspiracy face up in the command zone.)
//	 At the beginning of the first upkeep, create a 1/1 colorless Construct
//	 artifact creature token with defender."
//
// A face-up conspiracy (CR 315.5, ADR 0227). "The first upkeep" is the game's
// first upkeep, the first player's (the step is read as that player's first,
// Player.UpkeepsBegun, and no other upkeep has begun before it); the token is
// the conspiracy's controller's whichever seat's upkeep it is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7df49476-7b6c-4b5d-bede-463a7d2e0a58",
		Name:         "Sentinel Dispatch",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, effects.S82OldB361FirstUpkeepOfTheGame,
				"Sentinel Dispatch — create a 1/1 colorless Construct artifact creature token with defender",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Construct artifact with defender"), N: 1})),
		},
	})
}
