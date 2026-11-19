package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Code of Constraint — Instant {2}{U} (Ravnica: Clue Edition, #5929):
//
//	"Target creature gets -4/-0 until end of turn.
//	 Draw a card.
//	 Addendum — If you cast this spell during your main phase, tap that
//	 creature and it doesn't untap during its controller's next untap step."
//
// The -4/-0, then the draw, then the addendum, read as the spell resolves
// with b14InYourMainPhase (the stack has to empty for a step to end, so that
// is where it was cast). The tap and the freeze are TapAndFreeze (CR 611.2c).
// A target that became illegal counters the whole spell, draw included
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b2f83483-24a4-45db-a04e-42ff732ca98a",
		Name:         "Code of Constraint",
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(HarmUntilEOT(0, -4, 0))},
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return BoostUntilEOT{Target: id, Power: -4, Label: "Code of Constraint — -4/-0"}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error { return DrawCards{N: 1}.Apply(ctx) }),
			Instruction(func(ctx *Context) error {
				if !b14InYourMainPhase(ctx.Game, ctx.Controller()) {
					return nil
				}
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Code of Constraint — it doesn't untap during its controller's next untap step"}.Apply(ctx)
			}),
		),
	})
}
