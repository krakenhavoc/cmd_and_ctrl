package effects

import (
	"github.com/google/uuid"

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
	pumpAndGrant := pumpAndGrantFirstTarget(1, 3, "Leaping Ambush — +1/+3 and reach", "reach")
	Register(Spec{
		OracleID:     "e920c955-66f5-4c6e-a2a8-8b282f9e456e",
		Name:         "Leaping Ambush",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 3, "reach")),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			target := FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			if err := pumpAndGrant(ctx.Game, item); err != nil {
				return err
			}
			return UntapTarget{Target: target}.Apply(ctx)
		},
	})
}
