package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Argothian Opportunist — Creature — Human Scout {2}{G}, 3/2 (The
// Brothers' War, #5482):
//
//	"When this creature enters, create a tapped Powerstone token."
//
// A tapped Powerstone made on enter, as Stone Retrieval Unit's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cc9d3868-9814-4506-80b3-660bc462ff98",
		Name:         "Argothian Opportunist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Argothian Opportunist — create a tapped Powerstone token",
				Do(CreateTokenAdvanced{Spec: Token(PowerstoneToken()).EntersTapped(), N: 1})),
		},
	})
}
