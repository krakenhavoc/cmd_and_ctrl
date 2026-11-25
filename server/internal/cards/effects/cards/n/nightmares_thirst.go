package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nightmare's Thirst — Instant {B} (Core Set 2019, S82 OLD batch 198, #5527):
//
//	"You gain 1 life. Target creature gets -X/-X until end of turn, where X is
//	 the amount of life you gained this turn."
//
// The life first, then the shrink: X is read from the turn's life-gained
// tally as the shrink is applied (CR 608.2h), so the 1 just gained, lifelink
// and every earlier gain this turn count and a loss does not. With its one
// target illegal the spell does nothing at all, the life included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ecf56b9-1f36-4030-8333-1675ccc663d9",
		Name:         "Nightmare's Thirst",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.GainLife{Amount: 1}).Apply(ctx); err != nil {
				return err
			}
			x := effects.B15LifeGainedThisTurn(ctx.Game, ctx.Controller())
			return effects.BoostUntilEOT{Target: target, Power: -x, Toughness: -x, Label: "Nightmare's Thirst — -X/-X"}.Apply(ctx)
		},
	})
}
