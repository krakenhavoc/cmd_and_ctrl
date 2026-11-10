package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ed428638-181b-47cb-92f4-379db020e634",
		Name:         "Kulrath Mystic",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(ManaValueGE(4),
				"Kulrath Mystic — this creature gets +2/+0 and gains vigilance until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					if err := thisGetsUntilEndOfTurn(2, 0, "Kulrath Mystic — +2/+0 until end of turn")(g, item); err != nil {
						return err
					}
					return thisGainsKeywordUntilEndOfTurn("vigilance", "Kulrath Mystic — vigilance until end of turn")(g, item)
				}),
		},
	})
}
