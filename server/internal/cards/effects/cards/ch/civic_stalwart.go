package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "80d7e395-f774-44b5-9442-fa2fc03aaa29",
		Name:         "Civic Stalwart",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Civic Stalwart — creatures you control get +1/+1 until end of turn",
				effects.Do(effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl()),
					Power: 1, Toughness: 1,
					Label: "Civic Stalwart — +1/+1",
				})),
		},
	})
}
