package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yavimaya Sapherd — Creature — Fungus {2}{G}, 2/2 (Dominaria, #5530):
//
//	"When this creature enters, create a 1/1 green Saproling creature token."
//
// One 1/1 green Saproling when it enters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "42a1954f-774e-4d80-98b3-a736f23e2060",
		Name:         "Yavimaya Sapherd",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(WhenThisEnters("Yavimaya Sapherd — create a 1/1 green Saproling creature token",
				Do(CreateToken{Template: TokenCard("1/1 green Saproling"), N: 1})), game.Purpose{Tokens: 1}),
		},
	})
}
