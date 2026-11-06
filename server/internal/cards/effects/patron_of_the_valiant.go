package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Patron of the Valiant — Creature — Angel {3}{W}{W}, 4/4:
//
//	"Flying
//	 When this creature enters, put a +1/+1 counter on each creature
//	 you control with a +1/+1 counter on it."
//
// The set is read as the trigger resolves, so a creature that got its
// first counter in response counts and one that lost its last does not;
// Patron itself counts if it entered with a counter. One counter each,
// as one placement per creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4d36ea21-e8cb-442d-ba3f-c399289ac208",
		Name:            "Patron of the Valiant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Patron of the Valiant — put a +1/+1 counter on each creature you control with a +1/+1 counter on it",
				putPlusOneCountersOnEachYouControlThat(1, tomikHasPlusCounter)),
		},
	})
}
