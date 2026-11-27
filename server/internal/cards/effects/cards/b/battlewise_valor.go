package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battlewise Valor — Instant {1}{W} (Theros, S82 OLD b389, #5561):
//
//	"Target creature gets +2/+2 until end of turn. Scry 1."
//
// Furious Bellow's shape: the pump on the one target, then the scry
// (CR 701.22a).
//
// No simplification.
func init() {
	purpose := effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2))
	purpose.Scry = 1
	effects.Register(effects.Spec{
		OracleID:     "b4540f6a-bec8-4942-aebe-d3285b8a8134",
		Name:         "Battlewise Valor",
		Completeness: effects.CompletenessFull,
		Purpose:      purpose,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Battlewise Valor — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
