package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gaze of the Gorgon — Instant {3}{B/G} (Ravnica: City of Guilds, S82 OLD batch
// 581, #5600):
//
//	"({B/G} can be paid with either {B} or {G}.)
//	 Regenerate target creature. At this turn's next end of combat, destroy all
//	 creatures that blocked or were blocked by it this turn."
//
// Two sentences as steps of one walk. The regeneration shield is the ordinary
// one (CR 701.19a). The delayed trigger (CR 603.7) is scheduled for the next
// end of combat step of this turn only, and with no end of combat left this
// turn it never fires; it reads this turn's block pairs as it resolves
// (GazeOfTheGorgonBody), so a creature that blocked or was blocked by the
// target is destroyed even when the target itself has since left. An illegal
// target as the spell resolves does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec62aed5-b59b-4ba0-b265-a9c2a798ffb3",
		Name:         "Gaze of the Gorgon",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.Regenerate{Target: t.ID}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.ScheduleDelayedTrigger{
					At:           game.StepEndCombat,
					Label:        "Gaze of the Gorgon — destroy all creatures that blocked or were blocked by it this turn",
					Cards:        []uuid.UUID{t.ID},
					Body:         effects.GazeOfTheGorgonBody,
					Params:       game.EffectParams{Amount: effects.EndCombatsBegunThisTurn(ctx.Game)},
					EachThisTurn: true,
				}.Apply(ctx)
			}),
		),
	})
}
