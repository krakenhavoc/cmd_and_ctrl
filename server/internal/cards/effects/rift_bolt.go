package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rift Bolt — {2}{R} Sorcery: "Rift Bolt deals 3 damage to any
// target. Suspend 1—{R}"
//
// Lightning Bolt with a turn of delay, and the delay is the card: {R}
// now for 3 damage on your next upkeep. It is also the smallest real
// suspend card, which makes it the one worth reading — a SORCERY, so
// it may only be suspended at sorcery speed (CR 702.62c), and the
// free cast on the last counter happens during an UPKEEP, which is
// why the suspend grant carries flash timing (CR 608.2g).
//
// Everything about the keyword is the engine's: the special action,
// the time counter, the exile-zone countdown trigger, the free cast
// and the CR 118.6 reading of what "without paying its mana cost"
// means (#659). The card file is a damage clause and one line.
func init() {
	Register(Spec{
		OracleID:     "2b8afa9f-4236-4c02-a8d5-3c145caecfd6",
		Name:         "Rift Bolt",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		SpecialActions: []game.SpecialAction{
			Suspend(1, "{R}"),
		},
		OnResolve: damageToFirstTarget(3),
	})
}
