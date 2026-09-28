package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Two-Headed Giant of Foriys — Creature — Giant {4}{R}, 4/4:
//
//	"Trample
//	 This creature can block an additional creature each combat."
//
// Trample is the engine's keyword; the second line is
// CanBlockAdditional(1) on itself (#1706). Trample does nothing for a
// blocker, so a Giant blocking two attackers simply divides its 4
// between them (CR 510.1d).
func init() {
	Register(Spec{
		OracleID:        "38aa31bd-7145-43b9-9409-463d9ad6cd69",
		Name:            "Two-Headed Giant of Foriys",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static:          []game.StaticAbility{CanBlockAdditional(selfOnly, 1)},
	})
}
