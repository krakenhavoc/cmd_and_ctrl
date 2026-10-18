package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// do_walk.go — a Do's later steps wait for an earlier step's prompt
// (#4012, CR 608.2c; the engine half is game/step_walk.go).
//
// A step that asks something only queues the prompt and returns. Before
// each step, the walk asks the engine whether the resolution is paused
// on one of its own prompts (game.ParkEffectStepsForEffect); if it is,
// the engine parks the position of that step and the walk returns. The
// answer's boundary resumes it through resumeDoSteps.
//
// The position is data: the item, and a path of step indexes into the
// item's own Effect. Restore re-derives a stamped ability's Effect from
// its catalog row, and the Do answers the steps probe with the same
// steps, so the path still names the same step. That is why only the
// Do that IS the item's Effect parks (isItemEffectSteps): a Do called
// from a hand-written closure could not be found again from data, and
// runs its steps back to back as it always has.
//
// An If step's Then is walked the same way, one level down in the path,
// so an intervening-if row (InterveningIf, AtYourEndStepIfACreatureDied
// wrap their steps in one) waits too. Its condition is checked once, as
// the If step begins; a resume inside its Then does not check it again.

// ownsItemEffect is isItemEffectSteps, reached through a variable set at
// init: Do cannot name it directly, because it reads doStepsOf, whose
// doClosurePC is initialised by calling Do (an initialisation cycle).
// Nothing runs a Do on a real item before init has finished.
var ownsItemEffect func(item *game.StackItem, steps []Applier) bool

func init() {
	ownsItemEffect = isItemEffectSteps
	game.ResumeEffectSteps = resumeDoSteps
}

// isItemEffectSteps reports whether steps are the steps of item's own
// Effect — the same backing array, not merely equal values, so a second
// Do with the same primitives is not mistaken for it.
func isItemEffectSteps(item *game.StackItem, steps []Applier) bool {
	if item == nil || len(steps) == 0 {
		return false
	}
	own, ok := doStepsOf(item.Effect)
	return ok && len(own) == len(steps) && &own[0] == &steps[0]
}

// resumeDoSteps is game.ResumeEffectSteps: run item's Effect steps from
// path. An Effect that is no longer a Do (a deploy rewrote the row) has
// nothing to resume.
func resumeDoSteps(g *game.Game, item *game.StackItem, path []int) error {
	if item == nil || len(path) == 0 {
		return nil
	}
	steps, ok := doStepsOf(item.Effect)
	if !ok {
		return nil
	}
	return runDoSteps(NewContext(g, item), steps, path, true)
}

// runDoSteps applies steps in order from the position `from` names (nil
// is the start). When parkable, it parks the rest behind a paused
// resolution instead of running it.
func runDoSteps(ctx *Context, steps []Applier, from []int, parkable bool) error {
	_, err := walkDoSteps(ctx, steps, from, nil, parkable)
	return err
}

// walkDoSteps is runDoSteps one level down: `at` is the path to steps,
// and the result says whether the walk parked.
func walkDoSteps(ctx *Context, steps []Applier, from, at []int, parkable bool) (bool, error) {
	start := 0
	if len(from) > 0 {
		start = from[0]
	}
	for i := start; i < len(steps); i++ {
		// Resuming inside this step's own steps: the step has begun,
		// so neither the pause check nor an If's condition runs again.
		inside := i == start && len(from) > 1
		here := append(at[:len(at):len(at)], i)
		if !parkable {
			if err := steps[i].Apply(ctx); err != nil {
				return false, err
			}
			continue
		}
		if !inside && ctx.Game.ParkEffectStepsForEffect(ctx.Item, here) {
			return true, nil
		}
		iff, ok := steps[i].(If)
		if !ok {
			if err := steps[i].Apply(ctx); err != nil {
				return false, err
			}
			continue
		}
		var sub []int
		if inside {
			sub = from[1:]
		} else if iff.Cond == nil || !iff.Cond.Holds(ctx) {
			continue
		}
		if parked, err := walkDoSteps(ctx, iff.Then, sub, here, true); parked || err != nil {
			return parked, err
		}
	}
	return false, nil
}
