package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fight for the Throne — Instant {1}{G} (Marvel Super Heroes Commander,
// #4455):
//
//	"Put a +1/+1 counter on target creature you control. Then it fights
//	 target creature an opponent controls. When the creature an opponent
//	 controls dies this turn, if you control your commander, you become
//	 the monarch."
//
// The counter and the fight are Savage Stomp's two steps. The last
// sentence is a delayed triggered ability the spell creates (CR 603.7a),
// about the opponent's creature as the object it is now: a creature that
// left and came back is a new object and never fires it. Its "if" is an
// intervening one, so the condition (a creature that is its controller's
// commander, controlled by its owner) is checked as the creature dies and
// again as the trigger resolves. The delayed trigger is made whether or
// not the fight happened, as long as the spell resolved with the
// opponent's creature still on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "53129de9-4809-4c8c-9d11-37369899e70a",
		Name:         "Fight for the Throne",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature an opponent controls", OpponentControls())),
		OnResolve: ResolveSteps(
			Instruction(counterOnFirstCreatureTarget),
			Instruction(fightTargetClauses),
			Instruction(func(ctx *Context) error {
				theirs, ok := ctx.ClauseTarget(1)
				if !ok {
					return nil
				}
				return whenThatCreatureDiesThisTurnIf(ctx, theirs.ID,
					"Fight for the Throne — if you control your commander, you become the monarch", fightForTheThroneDied,
					fightForTheThroneMonarchBody)
			}),
		),
	})
}

var (
	// fightForTheThroneDied is "that creature died while its delayed
	// trigger's controller controls their commander".
	fightForTheThroneDied = game.DelayedEachCondition("dies/that-creature-you-control-your-commander",
		func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, p game.EffectParams) []uuid.UUID {
			if !b18ControlsYourCommander(g, dt.Controller) {
				return nil
			}
			return thatCreatureDied(ev, dt, g, p)
		})

	fightForTheThroneMonarchBody = game.SimpleDelayedBody("fight-for-the-throne/monarch",
		func(g *game.Game, item *game.StackItem) error {
			if !b18ControlsYourCommander(g, item.Controller) {
				return nil
			}
			return BecomeTheMonarch{}.Apply(NewContext(g, item))
		})
)
