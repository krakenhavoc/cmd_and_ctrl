package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rathi Fiend — Creature — Phyrexian Horror Mercenary {3}{B}, 2/2
// (Nemesis, #5618):
//
//	"When this creature enters, each player loses 3 life.
//	 {3}, {T}: Search your library for a Mercenary permanent card with mana value 3 or less, put it onto the battlefield, then shuffle."
//
// The enters trigger uses the stack (nemEachPlayerLoses at 3: every
// player, the controller included, loses 3 life). The second row is the
// Mercenary recruiter (nemRecruiterRow) at mana value 3 or less; the
// searcher may find nothing (CR 701.23b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b336d819-f299-4de3-9b23-51f83da59efb",
		Name:         "Rathi Fiend",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Rathi Fiend — each player loses 3 life", nemEachPlayerLoses(3)),
		},
		Activated: []ActivatedAbility{
			nemRecruiterRow("Rathi Fiend", "{3}", "a", "Mercenary", 3),
		},
	})
}
