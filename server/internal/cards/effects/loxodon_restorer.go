package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loxodon Restorer — Creature — Elephant Cleric {4}{W}{W}, 3/4 (Guilds of Ravnica, #5525):
//
//	"Convoke
//	 When this creature enters, you gain 4 life."
//
// Convoke is the shared tap cost (CR 702.51a); the enter trigger is
// Obstinate Baloth's life gain.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77eb2374-44a0-4383-b229-edc116c7d835",
		Name:         "Loxodon Restorer",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Loxodon Restorer — you gain 4 life", Do(GainLife{Amount: 4})),
		},
	})
}
