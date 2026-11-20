package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sorin's Guide — Creature — Vampire {3}{B}{B}, 4/2 (Core Set 2020, #5518):
//
//	"When this creature enters, you may search your library and/or graveyard
//	 for a card named Sorin, Vampire Lord, reveal it, and put it into your
//	 hand. If you search your library this way, shuffle."
//
// The partner-search trigger (mayFetchNamedFromLibraryOrGraveyard).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "36f5f161-b25e-48e4-b8f9-e6c36dfefe5b",
		Name:         "Sorin's Guide",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{mayFetchNamedFromLibraryOrGraveyard("Sorin's Guide", "Sorin, Vampire Lord")},
	})
}
