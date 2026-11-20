package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mystic of the Hidden Way — Creature — Human Monk {4}{U}, 3/2 (Khans of
// Tarkir, #5554):
//
//	"This creature can't be blocked.
//	 Morph {2}{U} (You may cast this card face down as a 2/2 creature
//	 for {3}. Turn it face up any time for its morph cost.)"
//
// Dimir Infiltrator's RestrictSelf(CantBeBlocked) static, which the
// declare-blockers check reads (CR 509.1b); face down the 2/2 has no
// such text (CR 708.2a). Morph is the keyword's declaration (morph.go,
// CR 702.37): cast face down for {3} it is a nameless 2/2 with no text
// (CR 708.2a), and {2}{U} turns it face up as the same object (CR
// 708.8).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "530853b4-68a0-4814-b0aa-ffeb5585cd43",
		Name:             "Mystic of the Hidden Way",
		Completeness:     CompletenessFull,
		Static:           []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
		AlternativeCosts: []game.AlternativeCost{Morph("{2}{U}")},
	})
}
