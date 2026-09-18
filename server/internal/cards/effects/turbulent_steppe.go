package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Turbulent Steppe — Land — Mountain Plains (EDHREC rank 4247):
//
//	"({T}: Add {R} or {W}.)
//	 This land enters tapped unless your opponents control eight or
//	 more lands."
//
// The Boros member of the same "catch-up" cycle as Turbulent
// Wilderness beside it, and the third of five to reach the catalog
// after Turbulent Fen (batch 34). Identical machinery, a different
// colour pair; see turbulent_wilderness.go for why the pipe is
// declared and why the tapped entry is a replacement rather than a
// tap on arrival.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "db444f9d-4dde-4308-b0f2-7acfe6de871a",
		Name:          "Turbulent Steppe",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTappedUnless(b40CatchUpDualCondition())},
		ManaAbilities: []ManaAbility{dualManaAbility("R", "W")},
	})
}
