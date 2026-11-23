package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Doomskar Titan — Creature — Giant Berserker {4}{R}{R}, 4/4 (Kaldheim,
// #5505):
//
//	"When this creature enters, creatures you control get +1/+0 and gain
//	 haste until end of turn.
//	 Foretell {4}{R}"
//
// Imodane's Recruiter's enters trigger, plus the shared foretell
// special action (CR 702.143). The Titan is one of the creatures you
// control as it resolves, so it can attack the turn it arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "86692047-813f-47b1-9f4e-c8b9d5b17b1c",
		Name:           "Doomskar Titan",
		Completeness:   effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{effects.Foretell("{4}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Doomskar Titan — creatures you control get +1/+0 and gain haste until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					yours := effects.And(effects.Creature(), effects.YouControl())
					if err := (effects.BoostUntilEOT{Match: yours, Power: 1, Label: "Doomskar Titan — +1/+0"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Match: yours, Keywords: []string{"haste"}, Label: "Doomskar Titan — haste"}.Apply(ctx)
				}),
		},
	})
}
