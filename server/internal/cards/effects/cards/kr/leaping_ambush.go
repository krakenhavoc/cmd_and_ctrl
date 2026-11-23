package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Leaping Ambush — Instant {G} (Wilds of Eldraine, #4471):
//
//	"Target creature gets +1/+3 and gains reach until end of turn. Untap
//	 it."
//
// Pillar Launch at +1/+3: the pump and the grant are
// pumpAndGrantFirstTarget's body, and "untap it" follows on the same
// still-legal target (CR 608.2b), so a creature that left in response
// gets nothing.
//
// No simplification.
func init() {
	pumpAndGrant := effects.PumpAndGrantFirstTarget(1, 3, "Leaping Ambush — +1/+3 and reach", "reach")
	effects.Register(effects.Spec{
		OracleID:     "e920c955-66f5-4c6e-a2a8-8b282f9e456e",
		Name:         "Leaping Ambush",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 3, "reach")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			target := effects.FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			if err := pumpAndGrant(ctx.Game, item); err != nil {
				return err
			}
			return effects.UntapTarget{Target: target}.Apply(ctx)
		},
	})
}
