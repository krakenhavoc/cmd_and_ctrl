package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vines of the Recluse — Instant {G} (Ravnica: Clue Edition):
//
//	"Target creature gets +1/+2 and gains reach until end of turn. Untap
//	 it."
//
// Massive Might's pump-and-grant, then the untap of the same still-legal
// target (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5372445b-5f67-435e-a736-61c6cb4f9c6e",
		Name:         "Vines of the Recluse",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 2, "reach")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.PumpAndGrantFirstTarget(1, 2, "Vines of the Recluse — +1/+2 and reach", "reach")(ctx.Game, item); err != nil {
				return err
			}
			if id := effects.FirstLegalBattlefieldTarget(ctx); id != [16]byte{} {
				return effects.UntapTarget{Target: id}.Apply(ctx)
			}
			return nil
		},
	})
}
