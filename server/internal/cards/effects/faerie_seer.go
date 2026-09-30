package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Faerie Seer — Creature — Faerie Wizard {U}, 1/1 (EDHREC rank 1129):
//
//	"Flying
//	 When this creature enters, scry 2."
//
// Flying rides PrintedKeywords; the ETB is an ordinary triggered
// ability watching EventETB with the shared Self condition, so a
// counterspell on the creature spell stops the trigger from ever
// firing (the permanent never enters), and a Doom Blade in response
// to the trigger does not — the scry already happened, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b2e65e8b-5f08-4cc2-ab1d-00f8903dbea2",
		Name:            "Faerie Seer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Faerie Seer — scry 2", Do(Scry{N: 2})),
		},
	})
}
