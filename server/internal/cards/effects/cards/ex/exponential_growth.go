package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Exponential Growth — Sorcery {X}{X}{G}{G} (Strixhaven: School of Mages,
// S82 batch 047, #5502):
//
//	"Until end of turn, double target creature's power X times."
//
// One doubling is +P/+0 with P the creature's power as that doubling is
// applied (CR 701.10b), so doubling X times is X separate readings, each
// seeing the one before it: DoublePowerUntilEOT read X times. X is the
// announced one (the {X}{X} cost is two mana for each), and X of 0 doubles
// nothing. A target that stopped being a legal creature does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "18616f70-e30c-4ff7-b62a-55939a96ca1b",
		Name:         "Exponential Growth",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			for i := 0; i < ctx.X(); i++ {
				if err := effects.DoublePowerUntilEOT(ctx, id, "Exponential Growth — double power until end of turn"); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
