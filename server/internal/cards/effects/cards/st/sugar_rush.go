package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sugar Rush — Instant {1}{B} (Wilds of Eldraine, #4471):
//
//	"Target creature gets +3/+0 until end of turn.
//	 Draw a card."
//
// The pump is the one until-end-of-turn modify on the target; the draw
// follows in printed order. A spell whose only target is illegal as it
// resolves does nothing at all (CR 608.2b), so the card is drawn only
// when the pump happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57085b8e-a85e-4d50-a7d2-9bddfdaf411c",
		Name:         "Sugar Rush",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.BenefitUntilEOT(0, 3, 0)), Draws: 1},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.BoostTheTargetUntilEOT(3, 0, "Sugar Rush — +3/+0")(item, ctx); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
