package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Noggle Bridgebreaker — Creature — Noggle Rogue {2}{U/R}{U/R}, 4/3 (Eventide, #5590):
//
//	"When this creature enters, return a land you control to its owner's
//	 hand."
//
// The bounce lands' return on a creature: no target, so the land is picked
// as the trigger resolves; with no land there is nothing to return.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d7e127e4-56fd-4116-9c1a-421beab58e5c",
		Name:         "Noggle Bridgebreaker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Noggle Bridgebreaker — return a land you control", Do(ReturnOneYouControl{
				Match:    MatchLand,
				Question: "Noggle Bridgebreaker — return a land you control to its owner's hand",
			})),
		},
	})
}
