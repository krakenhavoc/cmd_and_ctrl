package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sage of Shaila's Claim — Creature — Elf Druid {1}{G}, 2/1:
//
//	"When this creature enters, you get {E}{E}{E} (three energy
//	 counters)."
//
// ADR 0129 PR 1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a0b413cc-b16f-4112-8061-2ccded7f481d",
		Name:         "Sage of Shaila's Claim",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 3},
		Triggered:    []game.TriggeredAbility{WhenThisEntersYouGetEnergy("Sage of Shaila's Claim", 3)},
	})
}
