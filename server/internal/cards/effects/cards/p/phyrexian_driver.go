package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Driver — Creature — Phyrexian Zombie Mercenary {2}{B}, 1/1
// (Nemesis, S82 OLD batch 679, #5618):
//
//	"When this creature enters, other Mercenary creatures get +1/+1 until end
//	 of turn."
//
// Every other Mercenary creature on the battlefield, whoever controls it. The
// set is locked as the trigger resolves (CR 611.2c): a Mercenary that enters
// later this turn is not pumped. "Other" is the source object, so a Driver
// that left and came back is a new object (CR 400.7). A changeling counts as
// a Mercenary (CR 702.73a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e8d12d8e-afd2-438b-b121-5911fca479f9",
		Name:         "Phyrexian Driver",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Phyrexian Driver — other Mercenary creatures get +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.OfSubtype("Mercenary"), effects.TrekB11OtherThanSourceObject(item)),
						Power: 1, Toughness: 1,
						Label: "Phyrexian Driver — other Mercenaries get +1/+1",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
