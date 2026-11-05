package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pah-Wraith — Creature — Wraith {5}{R}{R}, 4/4 (Star Trek, #4973):
//
//	"First strike, haste
//	 Strange new worlds — Whenever this creature attacks, creatures you
//	 control get +X/+0 until end of turn, where X is the number of
//	 differently named lands you control."
//
// Strange new worlds is an ability word with no rules meaning (CR
// 207.2c). X is counted as the trigger resolves, each land name once
// however many lands share it, and the set of creatures that get the
// bonus is fixed then (CR 611.2c): a creature that arrives later this
// turn does not get it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e47c340d-b20b-4103-955b-3f76424c43f8",
		Name:            "Pah-Wraith",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Pah-Wraith — creatures you control get +X/+0, X the differently named lands you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return BoostUntilEOT{
						Match: And(Creature(), YouControl()),
						Power: socB23DifferentlyNamedLands(ctx),
						Label: "Pah-Wraith — strange new worlds",
					}.Apply(ctx)
				}),
		},
	})
}
