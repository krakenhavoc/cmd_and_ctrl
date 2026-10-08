package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leyline of Anticipation — Enchantment {2}{U}{U}:
//
//	"If this card is in your opening hand, you may begin the game with
//	 it on the battlefield.
//	 You may cast spells as though they had flash."
//
// Vedalken Orrery in blue, and the same one `Spec.CastTimings` entry
// (#1195). Derived from the battlefield per query, so a table with a
// Leyline and an Orrery has two sources saying the same thing and
// losing either changes nothing.
//
// The opening-hand clause (CR 103.6a) is Spec.OpeningHand (ADR 0133):
// the seat holding it is asked as the mulligan window closes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9dc65ffe-17fc-4280-b4bd-78073ac7e12b",
		Name:         "Leyline of Anticipation",
		Completeness: CompletenessFull,
		OpeningHand:  BeginTheGameOnTheBattlefield(),
		CastTimings:  []game.CastTimingRule{CastAsThoughFlash()},
	})
}
