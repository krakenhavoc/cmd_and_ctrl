package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Repulsive Mutation — Instant {X}{G}{U} (Murders at Karlov Manor, #4469):
//
//	"Put X +1/+1 counters on target creature you control. Then counter up
//	 to one target spell unless its controller pays mana equal to the
//	 greatest power among creatures you control."
//
// Two target clauses: the creature (slot 0) and up to one spell (slot 1),
// each judged on its own at resolution (CR 608.2b), so a creature that
// has gone does not stop the counter half. "Then" makes the amount
// follow the counters: the greatest power is read after they are placed,
// through the layered view, and counts the creature that got them. A
// greatest power of 0 asks for nothing and counters nothing, since the
// controller trivially pays {0}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "611bfaa6-b098-48d9-9e49-1e7dc8687f15",
		Name:         "Repulsive Mutation",
		Completeness: CompletenessFull,
		Targets: TargetCreature("target creature you control", YouControl()).Then(
			TargetSpell("up to one target spell").WithCount(0, 1)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok {
				if x := ctx.X(); x > 0 {
					if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: x}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			spell, ok := ctx.ClauseTarget(1)
			if !ok {
				return nil
			}
			power := b42GreatestPowerControlledBy(ctx.Game, ctx.Controller())
			if power <= 0 {
				return nil
			}
			return CounterUnlessPaid{
				StackID:  spell.ID,
				Cost:     fmt.Sprintf("{%d}", power),
				Question: fmt.Sprintf("Repulsive Mutation — pay {%d} or your spell is countered", power),
			}.Apply(ctx)
		},
	})
}
