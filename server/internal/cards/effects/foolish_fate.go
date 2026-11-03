package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Foolish Fate — Instant {2}{B} (Secrets of Strixhaven, #4456):
//
//	"Destroy target creature.
//	 Infusion — If you gained life this turn, that creature's controller
//	 loses 3 life."
//
// Infusion is an ability word (CR 207.2c) and does nothing itself. The
// creature's controller is read before the destroy (CR 608.2h: the
// controller it has as the spell resolves), so a creature that dies, or that
// regenerates or is indestructible and stays, costs its controller 3 life
// either way. The life loss is a second step of the same resolution (CR
// 608.2c), so a destroy that waits on a question still gets it; "if you
// gained life this turn" is read as the spell resolves, off your per-turn
// tally of life gained.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9c51893f-9118-4d95-b88a-e264f4cbd7a8",
		Name:         "Foolish Fate",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				if c, ok := ctx.Game.LookupCardForEffect(t.ID); ok {
					ctx.RecordStep(foolishFateKey, []uuid.UUID{c.Controller})
				}
				return DestroyTarget{Target: t.ID}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				if b15LifeGainedThisTurn(ctx.Game, ctx.Controller()) == 0 {
					return nil
				}
				ids := ctx.StepRecord(foolishFateKey).IDs
				if len(ids) == 0 {
					return nil
				}
				victim := ids[0]
				if p := ctx.Game.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
					return nil
				}
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -3)
			}),
		),
	})
}

const foolishFateKey = "foolish-fate/controller"
