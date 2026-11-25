package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heroes of the Revel — Creature — Satyr Soldier {4}{R}, 4/4 (Theros Beyond Death, #5514):
//
//	"When this creature enters, create a 1/1 red Satyr creature token with "This
//	 token can't block."
//	 Whenever you cast a spell that targets this creature, creatures you control get
//	 +1/+0 until end of turn."
//
// The Satyr is Anax's can't-block Satyr (GapB104SatyrCantBlockToken). The second
// ability is heroic's trigger (YouCastASpellThatTargetsThis, read off the spell's
// targets as it is cast, CR 601.2c); the creatures that get +1/+0 are the ones you
// control as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d4f72eac-2161-4fe8-8590-436e241c3b22",
		Name:         "Heroes of the Revel",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Heroes of the Revel — create a 1/1 red Satyr creature token with \"This token can't block.\"",
				effects.Do(effects.CreateToken{Template: effects.GapB104SatyrCantBlockToken(), N: 1})),
			effects.On(game.EventCast, effects.YouCastASpellThatTargetsThis,
				"Heroes of the Revel — creatures you control get +1/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Label: "Heroes of the Revel — +1/+0"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
