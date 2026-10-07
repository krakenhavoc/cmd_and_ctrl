package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Electrostatic Pummeler — Artifact Creature — Construct {3}, 1/1:
//
//	"When this creature enters, you get {E}{E}{E} (three energy
//	 counters).
//	 Pay {E}{E}{E}: This creature gets +X/+X until end of turn, where X
//	 is its power."
//
// ADR 0129 PR 1 (#1995). X is read once, as the ability resolves, from
// the creature's power then (counters and other effects included); a
// later pump is not counted again. A Pummeler that left and came back
// is a new object and gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "79676f10-9fca-4bca-acc5-6994955142b4",
		Name:         "Electrostatic Pummeler",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 3},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Electrostatic Pummeler", 3),
		},
		Activated: []ActivatedAbility{{
			Label: "Pay {E}{E}{E}: This creature gets +X/+X until end of turn, where X is its power.",
			Cost:  PayEnergy(3),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				c, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok {
					return nil
				}
				x := c.PowerForComparison()
				return thisGetsUntilEndOfTurn(x, x, "Electrostatic Pummeler — +X/+X until end of turn")(g, item)
			},
		}},
	})
}
