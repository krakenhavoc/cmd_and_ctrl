package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knight of Old Benalia — Creature — Human Knight {3}{W}{W}, 3/3 (Modern
// Horizons):
//
//	"Suspend 5—{W}
//	 When this creature enters, other creatures you control get +1/+1 until
//	 end of turn."
//
// Suspend is the shared special action (CR 702.62a): the free cast it ends
// in gives the creature haste. The enters trigger pumps every other creature
// its controller controls as it resolves, the set locked at that moment (CR
// 611.2c): a creature that enters later this turn is not pumped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "673ddb72-d800-49c3-9ea9-ff74a0e20a97",
		Name:           "Knight of Old Benalia",
		Completeness:   effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{effects.Suspend(5, "{W}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Knight of Old Benalia — other creatures you control get +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.YouControl(), effects.OtherThan(item.SourceCardID)), Power: 1, Toughness: 1,
						Label: "Knight of Old Benalia — other creatures you control get +1/+1",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
