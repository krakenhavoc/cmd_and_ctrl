package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angel of Suffering — Creature — Nightmare Angel {3}{B}{B}, 5/3:
//
//	"Flying
//	 If damage would be dealt to you, prevent that damage and mill twice that many cards."
//
// ADR 0108 §8 (#1906): a prevention static with an additional effect, one
// application per recipient (you) in a damage instance. "That many" is the
// damage it was applied to, so damage that can't be prevented still mills
// twice that many (CR 615.12; the ruling); the damage is prevented even
// when the library is shorter, and milling more than is there mills it all.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "92380c57-0a92-48eb-b562-2be149b5792a",
		Name:            "Angel of Suffering",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			PreventDamageDealtTo(PreventionStatic{
				To:    ToYou,
				Then:  angelOfSufferingMillBody,
				Label: "Angel of Suffering — prevent the damage and mill twice that many cards",
			}),
		},
	})
}

// The additional effect: mill twice "that many".
var angelOfSufferingMillBody = game.DelayedBody("angel-of-suffering/mill-twice", angelOfSufferingMill)

func angelOfSufferingMill(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	n := 2 * thatDamage(item)
	if n <= 0 {
		return nil
	}
	return g.MillNForEffect(item.Controller, n)
}
