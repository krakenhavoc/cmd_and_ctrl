package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fuel the Flames — Instant {2}{R} (Aetherdrift, #4464):
//
//	"Fuel the Flames deals 2 damage to each creature.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Pyroclasm at instant speed, with cycling (CR 702.29a). The creatures are
// read once as it resolves, yours included, so an uncrewed Vehicle is not
// hit and a crewed one is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e7d4e952-d7b9-489d-a9a7-c61d08bf17cc",
		Name:         "Fuel the Flames",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, Creature(), 2)
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
