package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloud of Faeries — Creature — Faerie {1}{U}, 1/1 (Wilds of Eldraine
// Commander, #5368):
//
//	"Flying
//	 When this creature enters, untap up to two lands.
//	 Cycling {2}"
//
// The untap is Snap's UntapUpToLands: "up to two lands" prints no "you
// control", so the choice runs over every tapped land at the table, made
// as the trigger resolves. Cycling is the stock cycling ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "adc1f3b4-cce7-42c9-a208-dbab1d1924f5",
		Name:         "Cloud of Faeries",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Cycling("{2}")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Cloud of Faeries — untap up to two lands", func(g *game.Game, item *game.StackItem) error {
				return UntapUpToLands{N: 2, Question: "Cloud of Faeries — untap up to two lands"}.Apply(NewContext(g, item))
			}),
		},
	})
}
