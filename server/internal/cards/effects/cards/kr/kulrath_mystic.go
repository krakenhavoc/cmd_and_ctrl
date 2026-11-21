package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kulrath Mystic — Creature — Elemental Wizard {2}{U}, 2/4 (Lorwyn
// Eclipsed, #4458):
//
//	"Whenever you cast a spell with mana value 4 or greater, this
//	 creature gets +2/+0 and gains vigilance until end of turn."
//
// The trigger is a cast trigger over the spell's mana value as it sits on
// the stack, so a spell with {X} counts the X chosen for it (CR 202.3e,
// the ruling). The bonus goes to the permanent that triggered: a Mystic
// that left in response gets nothing, and one that returned is a new
// object. The two parts are one effect, so a second trigger stacks +2/+0
// again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed428638-181b-47cb-92f4-379db020e634",
		Name:         "Kulrath Mystic",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.ManaValueGE(4),
				"Kulrath Mystic — this creature gets +2/+0 and gains vigilance until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					if err := effects.ThisGetsUntilEndOfTurn(2, 0, "Kulrath Mystic — +2/+0 until end of turn")(g, item); err != nil {
						return err
					}
					return effects.ThisGainsKeywordUntilEndOfTurn("vigilance", "Kulrath Mystic — vigilance until end of turn")(g, item)
				}),
		},
	})
}
