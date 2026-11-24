package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fortifying Draught — {G} Instant (Strixhaven: School of Mages, #5502):
//
//	"You gain 2 life. Target creature gets +X/+X until end of turn, where X is
//	 the amount of life you gained this turn."
//
// Blossoming Bogbeast's order: the 2 life goes through GainLife first, so a
// life-gain replacement has had its say before X is read. X is the turn's
// tally of life its controller gained (B15LifeGainedThisTurn), read once as
// the spell resolves (CR 608.2h), so it counts the spell's own 2 and any
// life gained earlier in the turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a34c78c6-ffbc-4461-a66d-1253a7aa963b",
		Name:         "Fortifying Draught",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.GainLife{Amount: 2}).Apply(ctx); err != nil {
				return err
			}
			x := effects.B15LifeGainedThisTurn(ctx.Game, ctx.Controller())
			return effects.BoostUntilEOT{Target: target, Power: x, Toughness: x,
				Label: "Fortifying Draught — +X/+X until end of turn"}.Apply(ctx)
		},
	})
}
