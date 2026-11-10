package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Towering Titan — Creature — Giant {4}{G}{G}, 0/0:
//
//	"This creature enters with X +1/+1 counters on it, where X is the
//	 total toughness of other creatures you control.
//	 Sacrifice a creature with defender: All creatures gain trample until
//	 end of turn."
//
// The counters are counted as the Titan enters (b19EntersWithCountersCounted),
// from effective toughness. The sacrifice is a cost, so the creature's
// toughness is no longer counted by anything later; the trample grant
// covers every creature on the battlefield as the ability resolves
// (CR 611.2c). A Titan with defender may sacrifice itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a95309ac-f920-474d-81b5-0f68c8a54f34",
		Name:         "Towering Titan",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted(game.CounterPlusOne, cmdrB21OtherCreatureToughness,
				"Towering Titan: enters with X +1/+1 counters, X the total toughness of other creatures you control"),
		},
		Activated: []ActivatedAbility{{
			Label: "Sacrifice a creature with defender: All creatures gain trample until end of turn.",
			Cost:  SacrificeN(1, "a creature with defender", Creature(), HasKeyword("defender")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{Match: Creature(), Keywords: []string{"trample"}, Label: "Towering Titan — trample"}.Apply(NewContext(g, item))
			},
		}},
	})
}
