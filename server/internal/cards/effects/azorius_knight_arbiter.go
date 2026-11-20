package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Azorius Knight-Arbiter — Creature — Human Knight {3}{W}{U}, 2/5 (Ravnica
// Allegiance, #5521):
//
//	"Vigilance
//	 This creature can't be blocked."
//
// Vigilance is printed data (CR 702.20b); the evasion is a RestrictSelf
// static the declare-blockers check reads (CR 509.1b), Talas Warrior's
// shape.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9adc1435-3094-427d-88f8-91d55c5a924d",
		Name:            "Azorius Knight-Arbiter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static:          []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
	})
}
