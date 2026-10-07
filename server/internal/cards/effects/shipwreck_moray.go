package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shipwreck Moray — Creature — Fish {3}{U}, 0/5:
//
//	"When this creature enters, you get {E}{E}{E}{E} (four energy
//	 counters).
//	 Pay {E}: This creature gets +2/-2 until end of turn."
//
// ADR 0129 PR 1 (#1995).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6c0b07c4-bb7c-4100-81e5-b0ee95d18625",
		Name:         "Shipwreck Moray",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 4},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Shipwreck Moray", 4),
		},
		Activated: []ActivatedAbility{{
			Label:  "Pay {E}: This creature gets +2/-2 until end of turn.",
			Cost:   PayEnergy(1),
			Effect: thisGetsUntilEndOfTurn(2, -2, "Shipwreck Moray — +2/-2 until end of turn"),
		}},
	})
}
