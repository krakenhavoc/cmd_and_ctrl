package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gallant Fowlknight — Creature — Kithkin Knight {3}{W}, 3/4 (Lorwyn
// Eclipsed, #4458):
//
//	"When this creature enters, creatures you control get +1/+0 until end
//	 of turn. Kithkin creatures you control also gain first strike until
//	 end of turn."
//
// Eagles of the North's enters trigger with the first strike narrowed to
// Kithkin. Both sets are fixed as it resolves (CR 611.2c): the Fowlknight
// itself is a Kithkin and gets both, and a creature that arrives later in
// the turn gets neither. "Kithkin" reads through Card.HasSubtypeWord, so a
// changeling counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4e0ba9d5-4bd6-49a0-ae00-70818738d0ba",
		Name:         "Gallant Fowlknight",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Gallant Fowlknight — creatures you control get +1/+0; Kithkin creatures you control gain first strike until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1,
						Label: "Gallant Fowlknight — +1/+0"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Match: And(OfCreatureType("Kithkin"), YouControl()),
						Keywords: []string{"first strike"},
						Label:    "Gallant Fowlknight — first strike until end of turn"}.Apply(ctx)
				}),
		},
	})
}
