package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rumbling Sentry — Creature — Giant {3}{W}{W}, 3/6 (Theros Beyond Death, #5514):
//
//	"When this creature enters, scry 1."
//
// Chrome Cat's enters scry.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7126772a-d586-4859-81ec-46bbe9d1a678",
		Name:         "Rumbling Sentry",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Rumbling Sentry — scry 1", Do(Scry{N: 1})),
		},
	})
}
