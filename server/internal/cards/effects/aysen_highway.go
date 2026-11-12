package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aysen Highway — Enchantment {3}{W}{W}{W} (Homelands, #5635):
//
//	"White creatures have plainswalk. (They can't be blocked as long as
//	defending player controls a Plains.)"
//
// A layer 6 grant (CR 613.1f) to every white creature on the
// battlefield, whoever controls it. Plainswalk is landwalk, which the
// block check reads off the defending player's lands (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "415bc26d-60fb-41ac-8006-7dd861671c14",
		Name:         "Aysen Highway",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			Labelled(KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.HasColor(source.TextColor("W"))
			}, "plainswalk"), "White creatures have plainswalk."),
		},
	})
}
