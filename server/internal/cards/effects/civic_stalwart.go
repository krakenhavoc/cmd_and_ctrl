package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Civic Stalwart — Creature — Elephant Soldier {3}{W}, 3/3 (Ravnica
// Allegiance, #5521):
//
//	"When this creature enters, creatures you control get +1/+1 until end
//	 of turn."
//
// Inspiring Captain's enters trigger: a +1/+1 until end of turn over the
// creatures you control as it resolves (CR 611.2c locks the set then).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "80d7e395-f774-44b5-9442-fa2fc03aaa29",
		Name:         "Civic Stalwart",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Civic Stalwart — creatures you control get +1/+1 until end of turn",
				Do(BoostUntilEOT{
					Match: And(Creature(), YouControl()),
					Power: 1, Toughness: 1,
					Label: "Civic Stalwart — +1/+1",
				})),
		},
	})
}
