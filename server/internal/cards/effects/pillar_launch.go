package effects

import (
	"github.com/google/uuid"

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
	pumpAndGrant := pumpAndGrantFirstTarget(2, 2, label, "reach")
	Register(Spec{
		OracleID:     "a39c758d-82be-4d8c-bb3e-8ee8a6c6788b",
		Name:         "Pillar Launch",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2, "reach")),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := pumpAndGrant(ctx.Game, item); err != nil {
				return err
			}
			target := FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			return UntapTarget{Target: target}.Apply(ctx)
		},
	})
}
