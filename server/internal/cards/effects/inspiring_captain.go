package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inspiring Captain — Creature — Human Knight {3}{W}, 3/3 (Innistrad
// Remastered):
//
//	"When this creature enters, creatures you control get +1/+1 until end
//	 of turn."
//
// Includes this creature; the set is fixed as the ability resolves
// (CR 611.2c).
func init() {
	Register(Spec{
		OracleID:     "4a294001-645b-4691-bb32-c240ba5b5320",
		Name:         "Inspiring Captain",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Inspiring Captain — creatures you control get +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Match: And(Creature(), YouControl()),
						Power: 1, Toughness: 1,
						Label: "Inspiring Captain — +1/+1",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
