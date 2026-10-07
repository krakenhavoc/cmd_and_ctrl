package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spontaneous Artist — Creature — Human Rogue {3}{R}, 3/3:
//
//	"When this creature enters, you get {E} (an energy counter).
//	 Pay {E}: Target creature gains haste until end of turn."
//
// ADR 0129 PR 1 (#1995). Any creature may be the target, the Artist
// included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c0602911-7d44-4a7a-a04c-837036ed3b82",
		Name:         "Spontaneous Artist",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Spontaneous Artist", 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "Pay {E}: Target creature gains haste until end of turn.",
			Cost:    PayEnergy(1),
			Targets: TargetCreature("target creature"),
			Effect:  ebTargetCreatureGainsUntilEOT("haste", "Spontaneous Artist — haste until end of turn"),
		}},
	})
}
