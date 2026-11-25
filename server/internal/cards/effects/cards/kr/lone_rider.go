package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// loneRiderOracle is shared by both faces: the front registers under it, the
// back under loneRiderOracle + "#1".
const loneRiderOracle = "849741a3-9e64-41a9-a4f4-d146a7d42389"

// Lone Rider // It That Rides as One — a transforming creature (Eldritch Moon,
// S82 OLD, #5541):
//
//	Lone Rider {1}{W}, Creature — Human Knight, 1/1
//	  "First strike, lifelink
//	   At the beginning of the end step, if you gained 3 or more life this
//	   turn, transform this creature."
//
//	It That Rides as One, Creature — Eldrazi Horror, 4/4
//	  "First strike, trample, lifelink"
//
// An intervening "if" (CR 603.4) on every end step, read from the
// controller's life gained this turn (Sorin of House Markov's tally): checked
// as the end step begins and again as the trigger resolves, and the creature
// turns over in place only if it is still the same face-up permanent
// (CR 701.27f, TransformThis).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        loneRiderOracle,
		Name:            "Lone Rider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventStepBegan, effects.AllOf(effects.StepBegan(game.StepEnd, false),
				func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B15LifeGainedThisTurn(g, source.Controller) >= 3
				}),
				"Lone Rider — transform this creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if effects.B15LifeGainedThisTurn(g, ctx.Controller()) < 3 {
						return nil // CR 603.4: checked again as it resolves
					}
					return effects.TransformThis{}.Apply(ctx)
				}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        loneRiderOracle + "#1",
		Name:            "It That Rides as One",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "trample", "lifelink"},
	})
}
