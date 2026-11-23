package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Bushwhacker — Creature — Goblin Warrior {R}, 1/1:
//
//	"Kicker {R}
//	 When this creature enters, if it was kicked, creatures you control get
//	 +1/+0 and gain haste until end of turn."
//
// Kicker is the optional additional cost; the enters trigger carries the
// kicked test as its intervening "if" (CR 603.4, Balduvian Atrocity's shape),
// and its body is Imodane's Recruiter's pump over creatures you control as it
// resolves, the Bushwhacker included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "878b6808-48af-4047-ac35-bfc58bf92e7d",
		Name:          "Goblin Bushwhacker",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{R}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
				"Goblin Bushwhacker — kicked, creatures you control get +1/+0 and gain haste until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					yours := effects.And(effects.Creature(), effects.YouControl())
					if err := (effects.BoostUntilEOT{Match: yours, Power: 1, Label: "Goblin Bushwhacker — +1/+0"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Match: yours, Keywords: []string{"haste"}, Label: "Goblin Bushwhacker — haste"}.Apply(ctx)
				}),
		},
	})
}
