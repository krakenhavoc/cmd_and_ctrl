package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chocobo Racetrack — Artifact {3}{G}{G} (Final Fantasy, #4462):
//
//	"Landfall — Whenever a land you control enters, create a 2/2 green
//	 Bird creature token with "Whenever a land you control enters, this
//	 token gets +1/+0 until end of turn.""
//
// A landfall that makes the set's Chocobo Bird (chocoboBirdToken), once
// per land. A Bird made by a land's trigger was not on the battlefield
// when that land entered, so that land does not pump it; every later
// land does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7522f7fd-6fda-4ed3-9131-21bf91616a73",
		Name:         "Chocobo Racetrack",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			LandfallPaying("Chocobo Racetrack — create a 2/2 green Bird (landfall)",
				game.LandfallPayoff{Tokens: 1},
				Do(CreateToken{Template: chocoboBirdToken(), N: 1})),
		},
	})
}
