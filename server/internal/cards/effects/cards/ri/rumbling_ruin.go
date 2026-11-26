package cardsri

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rumbling Ruin — Creature — Elemental {5}{R}, 6/6 (Ravnica Allegiance,
// #5521):
//
//	"When this creature enters, count the number of +1/+1 counters on
//	 creatures you control. Creatures your opponents control with power less
//	 than or equal to that number can't block this turn."
//
// The number is counted once, as the ability resolves, and is locked; the
// set stays live (CR 611.2c), so an opposing creature that arrives later in
// the turn with power at or under the number can't block either. Power is
// read live, after the layers (the bound is the number, the power is not
// frozen).
//
// No simplification.
func init() {
	const name = "Rumbling Ruin"
	effects.Register(effects.Spec{
		OracleID:     "8f36e0db-eea7-4c97-98ce-4bc2d8c7f734",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(name+" — creatures your opponents control with power less than or equal to the number of +1/+1 counters on creatures you control can't block this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					n := effects.PlusOneCountersOnCreaturesYouControl(g, ctx.Controller())
					return effects.CantBlockThisTurnMatching(true,
						name+" — creatures with power "+strconv.Itoa(n)+" or less can't block this turn",
						game.PermanentQuery{Types: []string{"creature"}}.WithPowerAtMost(n)).Apply(ctx)
				}),
		},
	})
}
