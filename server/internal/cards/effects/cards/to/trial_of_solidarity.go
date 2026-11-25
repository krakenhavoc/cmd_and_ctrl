package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trial of Solidarity — {2}{W} Enchantment (Amonkhet, S82 old batch 248):
//
//	"When this enchantment enters, creatures you control get +2/+1 and gain
//	 vigilance until end of turn.
//	 When a Cartouche you control enters, return this enchantment to its owner's
//	 hand."
//
// The creatures are chosen as the trigger resolves (CR 611.2c): later arrivals
// do not get the bonus.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cbcc2577-7556-4174-83ff-9e00611412b3",
		Name:         "Trial of Solidarity",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Trial of Solidarity — creatures you control get +2/+1 and gain vigilance until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					match := effects.And(effects.Creature(), effects.YouControl())
					if err := (effects.BoostUntilEOT{Match: match, Power: 2, Toughness: 1, Label: "Trial of Solidarity — +2/+1"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Match: match, Keywords: []string{"vigilance"}, Label: "Trial of Solidarity — vigilance"}.Apply(ctx)
				}),
			effects.WhenACartoucheYouControlEntersReturnThis("Trial of Solidarity"),
		},
	})
}
