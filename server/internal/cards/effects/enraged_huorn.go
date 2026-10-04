package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Enraged Huorn — Creature — Treefolk {4}{G}, 4/5:
//
//	"Trample
//	 When this creature enters, the Ring tempts you."
//
// The tempt is the keyword action (ADR 0114, CR 701.54). The creature
// is on the battlefield as the trigger resolves, so it can be chosen as
// the Ring-bearer.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8a7dbf4e-f121-40c5-a688-3398e04c1311",
		Name:            "Enraged Huorn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Enraged Huorn — the Ring tempts you", Do(TheRingTemptsYou{})),
		},
	})
}
