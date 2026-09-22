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
// One declared deviation, the one every Leyline in this catalog
// shares: there is no "begin the game with it on the battlefield"
// step in this sandbox, so the enchantment is cast for {2}{U}{U} like
// any other. See Leyline of the Void, which carries the same note.
func init() {
	Register(Spec{
		OracleID:     "9dc65ffe-17fc-4280-b4bd-78073ac7e12b",
		Name:         "Leyline of Anticipation",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"You can't begin the game with it on the battlefield from your opening hand — it has to be cast.",
		},
		CastTimings: []game.CastTimingRule{CastAsThoughFlash()},
	})
}
