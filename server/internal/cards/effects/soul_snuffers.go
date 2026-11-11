package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soul Snuffers — Creature — Elemental Shaman {2}{B}{B}, 3/3 (Lorwyn
// Eclipsed Commander, #5368):
//
//	"When this creature enters, put a -1/-1 counter on each creature."
//
// Every creature on the battlefield, its own controller's and the
// Snuffers itself included. The set is read before the first counter goes
// on (cmdr2B25CountersOnEach).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cdd227fc-0809-45d6-bed0-e85c0c0b6c94",
		Name:         "Soul Snuffers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Soul Snuffers — put a -1/-1 counter on each creature", cmdr2B25CountersOnEach(false, nil)),
		},
	})
}
