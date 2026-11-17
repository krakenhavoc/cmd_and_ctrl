package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Syndicate Infiltrator — Creature — Vampire Wizard {2}{U}{B}, 3/3 (Streets of
// New Capenna):
//
//	"Flying
//	 As long as there are five or more mana values among cards in your
//	 graveyard, this creature gets +2/+2."
//
// Flying is printed. The pump is a layer 7c self-modifier read live on every
// recompute (sncSelfPT), so it comes and goes as the graveyard's mana values do.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0d6be8a4-b05b-40fb-a653-5aabf71809e0",
		Name:            "Syndicate Infiltrator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{sncSelfPT(2, 2)},
	})
}
