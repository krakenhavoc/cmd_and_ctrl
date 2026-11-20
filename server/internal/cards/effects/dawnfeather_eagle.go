package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dawnfeather Eagle — Creature — Bird {4}{W}, 3/3 (Aether Revolt,
// #5536):
//
//	"Flying
//	 When this creature enters, creatures you control get +1/+1 and gain
//	 vigilance until end of turn."
//
// Stagecoach Security's enters trigger: the affected set is the
// creatures you control as it resolves, the Eagle included (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e17db9d8-14b0-4f72-aacc-f51fae4597e2",
		Name:            "Dawnfeather Eagle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dawnfeather Eagle — creatures you control get +1/+1 and gain vigilance until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					yours := And(Creature(), YouControl())
					if err := (BoostUntilEOT{Match: yours, Power: 1, Toughness: 1, Label: "Dawnfeather Eagle — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Match: yours, Keywords: []string{"vigilance"}, Label: "Dawnfeather Eagle — vigilance"}.Apply(ctx)
				}),
		},
	})
}
