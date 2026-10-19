package effects

import (
	"errors"
	"reflect"

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
//
// # Two more sites (#4019)
//
// A step sequence the engine can find again from data is not only an
// item's Effect. Two more are, and they park the same way:
//
//   - a spell's OnResolve written as ResolveSteps: found again through
//     the spell's catalog key (game.CardDef.ResumeResolve), and parked
//     only while the engine runs that spell's catalog hook
//     (game.ParkSpellStepsForEffect);
//   - a modal bullet written as ModeSteps: found again from the bullet
//     itself, which the mode walk already finds from its key and
//     occurrence (game.ResumeBulletSteps), and parked only while the
//     walk runs that occurrence (game.ParkBulletStepsForEffect).
//
// Instruction makes a hand-written function one step, so a body that
// cannot be spelled with primitive values is split at its prompts into
// steps instead: each runs whole, and the walk can stop between them.
//
// A continuation written as steps (RegisterThenSteps, then_steps.go,
// #4162) is a fourth site, found again by its registered key. Since a
// continuation can park while another walk is parked, each walk takes a
// mark as each step begins (game.StepWalkMark) and parks directly above
// it, so the walks resume in the order their instructions are written.

// ownsItemEffect is isItemEffectSteps, reached through a variable set at
// init: Do cannot name it directly, because it reads doStepsOf, whose
// doClosurePC is initialised by calling Do (an initialisation cycle).
// Nothing runs a Do on a real item before init has finished.
var ownsItemEffect func(item *game.StackItem, steps []Applier) bool

func init() {
	ownsItemEffect = isItemEffectSteps
	game.ResumeEffectSteps = resumeDoSteps
	game.ResumeBulletSteps = resumeBulletSteps
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

// stepParker parks the rest of a walk at `path` when the resolution is
// paused on one of its own prompts, and reports whether it did. mark is
// the parked walks as they stood when the step before path began
// (#4162, game.StepWalkMark). Each site has its own; nil means the walk
// cannot park.
type stepParker func(ctx *Context, path []int, mark game.StepWalkMark) bool

// parkItemEffect is the item's own Effect's site.
func parkItemEffect(ctx *Context, path []int, mark game.StepWalkMark) bool {
	return ctx.Game.ParkEffectStepsForEffect(ctx.Item, path, mark)
}

// parkSpellSteps is a spell's OnResolve's site (ResolveSteps).
func parkSpellSteps(ctx *Context, path []int, mark game.StepWalkMark) bool {
	return ctx.Game.ParkSpellStepsForEffect(ctx.Item, path, mark)
}

// parkBulletSteps is a modal bullet's site (ModeSteps). The mode walk
// parks the bullet's position itself (modeWalk.inner), after every step
// walk, so the mark is not needed.
func parkBulletSteps(ctx *Context, path []int, _ game.StepWalkMark) bool {
	return ctx.Game.ParkBulletStepsForEffect(ctx.Item, ctx.Occurrence(), path)
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
	return runStepsFrom(NewContext(g, item), steps, path, parkItemEffect)
}

// runDoSteps applies steps in order from the position `from` names (nil
// is the start). When parkable, it parks the rest behind a paused
// resolution instead of running it.
func runDoSteps(ctx *Context, steps []Applier, from []int, parkable bool) error {
	var park stepParker
	if parkable {
		park = parkItemEffect
	}
	return runStepsFrom(ctx, steps, from, park)
}

// runStepsFrom applies steps in order from the position `from` names (nil
// is the start), parking the rest through `park` behind a paused
// resolution. A nil park runs them back to back.
func runStepsFrom(ctx *Context, steps []Applier, from []int, park stepParker) error {
	_, err := walkDoSteps(ctx, steps, from, nil, park)
	return err
}

// walkDoSteps is runStepsFrom one level down: `at` is the path to steps,
// and the result says whether the walk parked.
func walkDoSteps(ctx *Context, steps []Applier, from, at []int, park stepParker) (bool, error) {
	start := 0
	if len(from) > 0 {
		start = from[0]
	}
	// The parked walks as the last step began (#4162): a walk this
	// one parks goes beneath any a step parked, and above the rest.
	var mark game.StepWalkMark
	if park != nil {
		mark = ctx.Game.StepWalkMarkForEffect()
	}
	for i := start; i < len(steps); i++ {
		// Resuming inside this step's own steps: the step has begun,
		// so neither the pause check nor an If's condition runs again.
		inside := i == start && len(from) > 1
		here := append(at[:len(at):len(at)], i)
		if park == nil {
			if err := steps[i].Apply(ctx); err != nil {
				return false, err
			}
			continue
		}
		if !inside && park(ctx, here, mark) {
			return true, nil
		}
		mark = ctx.Game.StepWalkMarkForEffect()
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
		if parked, err := walkDoSteps(ctx, iff.Then, sub, here, park); parked || err != nil {
			return parked, err
		}
	}
	return false, nil
}

// Instruction is a hand-written instruction as one step of a sequence
// (#4019): it runs whole, and a walk can park before it or after it but
// not inside it. Split a body at the calls that can ask something, one
// Instruction each, and a later instruction waits for an earlier one's
// answer (CR 608.2c).
//
// The function belongs to the catalog row, so a resume finds it again
// from the row. It must not hand anything to the next step through a
// variable: a restore taken between the two would lose it. Read what an
// earlier step did from the game instead.
type Instruction func(ctx *Context) error

func (f Instruction) Apply(ctx *Context) error {
	if f == nil {
		return nil
	}
	return f(ctx)
}

// ResolveSteps is a spell's OnResolve written as a sequence of steps
// (#4019, CR 608.2c): "Electrolyze deals 2 damage divided as you choose
// among one or two targets. Draw a card." is
//
//	OnResolve: ResolveSteps(Instruction(DealDividedDamage), DrawCards{N: 1}),
//
// and the draw waits for a prompt the damage asked (Khârn the
// Betrayer's question, a CR 616 order). The registry finds the steps
// again through the spell's catalog key (game.CardDef.ResumeResolve),
// so a restore taken while the prompt is open still draws after the
// answer. Called any other way (a spliced text, a fused half), the
// steps run back to back as a hand-written OnResolve would.
//
//go:noinline
func ResolveSteps(steps ...Applier) func(*game.StackItem, *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		if item == resolveStepsProbe {
			return doSteps(steps)
		}
		return runStepsFrom(ctx, steps, nil, parkSpellSteps)
	}
}

// resolveStepsProbe is the item the registry hands a ResolveSteps
// closure to ask for its steps; never a real item.
var resolveStepsProbe = &game.StackItem{}

// resolveStepsPC is the code pointer every ResolveSteps closure shares.
var resolveStepsPC = reflect.ValueOf(ResolveSteps()).Pointer()

// resolveStepsOf returns the steps of an OnResolve built by
// ResolveSteps, and false for any other OnResolve.
func resolveStepsOf(f func(*game.StackItem, *Context) error) ([]Applier, bool) {
	if f == nil || reflect.ValueOf(f).Pointer() != resolveStepsPC {
		return nil, false
	}
	var rep doStepsReport
	if err := f(resolveStepsProbe, nil); errors.As(err, &rep) {
		return rep.steps, true
	}
	return nil, false
}

// resumeResolveSteps is the CardDef.ResumeResolve of a spell whose
// OnResolve is ResolveSteps(steps...): the steps from path.
func resumeResolveSteps(steps []Applier) func(*game.Game, *game.StackItem, []int) error {
	return func(g *game.Game, item *game.StackItem, path []int) error {
		if item == nil || len(path) == 0 {
			return nil
		}
		return runStepsFrom(NewContext(g, item), steps, path, parkSpellSteps)
	}
}

// SpreeModeSteps is SpreeMode plus ModeSteps' own body (#4155): a Spree
// bullet with more than one instruction.
func SpreeModeSteps(label, cost string, targets *game.TargetSpec, steps ...Applier) game.ModeOption {
	o := ModeSteps(label, targets, steps...)
	o.Cost = cost
	return o
}

// ModeSteps declares one modal option whose body is a sequence of steps
// (#4019, CR 608.2c), for a bullet with more than one instruction:
// Demonic Pact's "This enchantment deals 4 damage to any target and you
// gain 4 life" is
//
//	ModeSteps("…", TargetAny(), Instruction(demonicPactDamage), GainLife{Amount: 4})
//
// and the life gain waits for a prompt the damage asked. The steps read
// their own occurrence with ctx.Occurrence() (ModeTarget(ctx,
// ctx.Occurrence()) is this bullet's target). The mode walk finds the
// bullet again from its key and occurrence, and the bullet answers the
// steps probe, so a restore taken while the prompt is open still runs
// the rest of the bullet, and then the later bullets, after the answer.
//
//go:noinline
func ModeSteps(label string, targets *game.TargetSpec, steps ...Applier) game.ModeOption {
	return game.ModeOption{
		Label:   label,
		Targets: targets,
		Effect: func(g *game.Game, item *game.StackItem, occurrence int) error {
			if item == modeStepsProbe {
				return doSteps(steps)
			}
			return runStepsFrom(bulletContext(g, item, occurrence), steps, nil, parkBulletSteps)
		},
	}
}

// modeStepsProbe is the item the walk hands a ModeSteps body to ask for
// its steps; never a real item.
var modeStepsProbe = &game.StackItem{}

// modeStepsPC is the code pointer every ModeSteps body shares.
var modeStepsPC = reflect.ValueOf(ModeSteps("", nil).Effect).Pointer()

// modeStepsOf returns the steps of a bullet body built by ModeSteps,
// and false for any other body.
func modeStepsOf(f func(*game.Game, *game.StackItem, int) error) ([]Applier, bool) {
	if f == nil || reflect.ValueOf(f).Pointer() != modeStepsPC {
		return nil, false
	}
	var rep doStepsReport
	if err := f(nil, modeStepsProbe, 0); errors.As(err, &rep) {
		return rep.steps, true
	}
	return nil, false
}

// resumeBulletSteps is game.ResumeBulletSteps: the parked bullet's steps
// from path. A body that is no longer a ModeSteps (a deploy rewrote the
// row) has nothing to resume.
func resumeBulletSteps(g *game.Game, item *game.StackItem, bullet func(*game.Game, *game.StackItem, int) error, occurrence int, path []int) error {
	if item == nil || len(path) == 0 {
		return nil
	}
	steps, ok := modeStepsOf(bullet)
	if !ok {
		return nil
	}
	return runStepsFrom(bulletContext(g, item, occurrence), steps, path, parkBulletSteps)
}

// bulletContext is the Context a bullet's steps run in: the item's,
// with the occurrence they belong to.
func bulletContext(g *game.Game, item *game.StackItem, occurrence int) *Context {
	ctx := NewContext(g, item)
	ctx.occurrence = occurrence + 1
	return ctx
}
