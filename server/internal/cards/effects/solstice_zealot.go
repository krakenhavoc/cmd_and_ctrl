package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Solstice Zealot — Creature — Rhino Cleric {2}{W}, 2/3:
//
//	"When this creature enters, you get {E}{E} (two energy counters).
//	 {T}, Pay {E}: Tap target creature."
//
// ADR 0129 PR 1 (#1995).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1ec63280-ae5b-4892-8854-de2d8127ea83",
		Name:         "Solstice Zealot",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Solstice Zealot", 2),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Pay {E}: Tap target creature.",
			Cost:    Plus(TapCost(), PayEnergy(1)),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
