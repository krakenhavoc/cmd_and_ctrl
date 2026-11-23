package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pillar Launch — Instant {G}:
//
//	"Target creature gets +2/+2 and gains reach until end of turn.
//	 Untap it."
//
// The pump and the grant are pumpAndGrantFirstTarget's body; "untap it"
// follows on the same still-legal target (CR 608.2b), so a creature
// that left in response gets nothing.
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	const label = "Pillar Launch — +2/+2 and reach"
	pumpAndGrant := effects.PumpAndGrantFirstTarget(2, 2, label, "reach")
	effects.Register(effects.Spec{
		OracleID:     "a39c758d-82be-4d8c-bb3e-8ee8a6c6788b",
		Name:         "Pillar Launch",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "reach")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := pumpAndGrant(ctx.Game, item); err != nil {
				return err
			}
			target := effects.FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			return effects.UntapTarget{Target: target}.Apply(ctx)
		},
	})
}
