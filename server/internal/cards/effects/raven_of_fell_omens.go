package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raven of Fell Omens — Creature — Bird {1}{B}, 1/2 (Outlaws of Thunder
// Junction, #4468):
//
//	"Flying
//	 Whenever you commit a crime, each opponent loses 1 life and you gain
//	 1 life. This ability triggers only once each turn. (Targeting
//	 opponents, anything they control, and/or cards in their graveyards
//	 is a crime.)"
//
// Marauding Sphinx's once-each-turn crime trigger
// (WheneverYouCommitACrimeOnceEachTurn, CR 700.13) with Ayara's drain
// (drainEachOpponent): life loss, not damage, and the gain is 1 whatever
// the number of opponents.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eec7d4f4-4141-49ed-8978-aefacc0a6609",
		Name:            "Raven of Fell Omens",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCommitACrimeOnceEachTurn("Raven of Fell Omens — each opponent loses 1 life and you gain 1 life", drainEachOpponent),
		},
	})
}
