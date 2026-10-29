package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fateful Discovery — Enchantment {3}{U}{U}:
//
//	"Whenever an artifact you control enters, draw a card."
//
// Quicksmith Genius's trigger condition (ArtifactEnteredUnderYourControl)
// with a draw. It fires for every artifact permanent that enters under
// the controller's control, a token included, and an artifact creature
// too; one entering under an opponent's control does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "72c6174c-fa2c-4ff2-b76e-eb6e890e97c9",
		Name:         "Fateful Discovery",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, ArtifactEnteredUnderYourControl, "Fateful Discovery — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
