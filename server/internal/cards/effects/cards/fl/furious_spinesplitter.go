package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Furious Spinesplitter — Creature — Ogre Warrior {2}{R/G}{R/G}, 3/3
// (Ravnica: Clue Edition, #5929):
//
//	"Trample
//	 At the beginning of your end step, put a +1/+1 counter on this
//	 creature for each opponent who was dealt damage this turn."
//
// The count is read as the trigger resolves, from the turn's damage
// events: damage actually dealt (a prevented point never was), combat or
// not, from any source, counted once per opponent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f4964f74-07c0-4edc-984d-57722fb16961",
		Name:            "Furious Spinesplitter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Furious Spinesplitter — put a +1/+1 counter on it for each opponent who was dealt damage this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					n := effects.S82BnsB12OpponentsDealtDamage(g, item.Controller)
					if n == 0 {
						return nil
					}
					return effects.CounterOnThis{Kind: game.CounterPlusOne, N: n}.Apply(ctx)
				}),
		},
	})
}
