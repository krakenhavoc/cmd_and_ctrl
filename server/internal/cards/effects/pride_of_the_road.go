package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pride of the Road — Creature — Zombie Cat Warrior {3}{W}, 2/5:
//
//	"Vigilance
//	 Start your engines!
//	 Max speed — At the beginning of combat on your turn, target creature
//	 or Vehicle you control gains double strike until end of turn."
//
// ADR 0138 (#2711). A real target clause (Perilous Snare's "creature or
// Vehicle you control"), so the trigger goes on the stack and can be
// answered; with no legal target it is removed (CR 603.3d). A Vehicle
// that is not a creature gains the keyword and it matters only if it is
// crewed before damage.
//
// No simplification.
func init() {
	const label = "Pride of the Road — target creature or Vehicle you control gains double strike until end of turn"
	Register(Spec{
		OracleID:        "2af236b0-6b30-4ddc-8735-864d2ffad68e",
		Name:            "Pride of the Road",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", StartYourEngines},
		Triggered: []game.TriggeredAbility{
			MaxSpeedTrigger(Targeting(
				AtBeginningOfYourCombat(label, ebTargetCreatureGainsUntilEOT("double strike", label)),
				TargetPermanent("target creature or Vehicle you control", Or(Creature(), OfSubtype("Vehicle")), YouControl()))),
		},
	})
}
