package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leonin Lightscribe — Creature — Cat Cleric {1}{W}, 2/2 (Strixhaven:
// School of Mages, #5502):
//
//	"Magecraft — Whenever you cast or copy an instant or sorcery spell,
//	 creatures you control get +1/+1 until end of turn."
//
// Inspiring Captain's team pump behind the shared Magecraft trigger. The
// set is fixed as the ability resolves (CR 611.2c), so a creature that
// arrives later in the turn is not pumped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "14079b33-cf68-4d16-962c-5c40518c4163",
		Name:         "Leonin Lightscribe",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Magecraft("Leonin Lightscribe — creatures you control get +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Match: And(Creature(), YouControl()),
						Power: 1, Toughness: 1,
						Label: "Leonin Lightscribe — +1/+1",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
