package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Essence Pulse — Sorcery {3}{B} (Commander 2021):
//
//	"You gain 2 life. Each creature gets -X/-X until end of turn, where X is
//	 the amount of life you gained this turn."
//
// The life comes first and X is read after it (CR 608.2c), once, from the
// turn's life-gained tally, so life gained earlier this turn counts and so
// does these 2 (a prevented or replaced gain does not). The -X/-X lands on
// every creature on the battlefield as the spell resolves, yours included
// (CR 611.2c), and does not follow a creature that enters later.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0d4bd5e0-8346-4025-8694-d2e1897ee0c2",
		Name:         "Essence Pulse",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, AmountIsX: true, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.GainLife{Player: ctx.Controller(), Amount: 2}).Apply(ctx); err != nil {
				return err
			}
			x := effects.B15LifeGainedThisTurn(ctx.Game, ctx.Controller())
			if x <= 0 {
				return nil
			}
			return effects.BoostUntilEOT{
				Match:     effects.Creature(),
				Power:     -x,
				Toughness: -x,
				Label:     "Essence Pulse — each creature gets -X/-X",
			}.Apply(ctx)
		},
	})
}
