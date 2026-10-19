package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// then_steps.go — a continuation whose later instruction waits for an
// earlier one's prompt (#4162, CR 608.2c), and what one step hands the
// next (#4163). The engine half is game/step_walk.go and
// game/step_record.go.
//
// A primitive's Then, an OnYes or OnAccept, the callback a
// *ThenForEffect method runs: each is a Go closure riding its prompt,
// and it runs when the prompt is answered. When such a continuation
// runs two instructions and the first asks something (a dredge offer on
// a draw, a CR 616 order on counters or life, Khârn the Betrayer's
// question), the second ran before the answer. No walk could stop it,
// because a walk parks a position the engine must find again from data,
// and a closure is not data.
//
// RegisterThenSteps names a sequence of steps with an on-disk key. Its
// Run walks them as a Do walks its own, and parks between them as the
// key and a path, so the rest runs after the answer, also on a table
// restored while the question is open. The closure that calls Run still
// rides its prompt as before (whatever census it was in, it stays in);
// the new part is that the rest of the continuation, once a step of it
// has asked, is data.
//
// The steps are package-level values and capture nothing. What they
// need from the closure that ran them (the number chosen, the creature
// sacrificed, the opponents still to ask) goes through RecordStep before
// Run, and each step reads it with StepRecord. A step must not hand
// anything to the next through a Go variable, for the same reason.

// ThenSteps is a continuation written as steps, registered under an
// on-disk key. It is also an Applier, so it can be one step of another
// sequence.
type ThenSteps struct {
	key   string
	steps []Applier
}

// thenStepsByKey holds every registered continuation's steps. Written
// only by package-level initialisation, read-only afterwards.
var thenStepsByKey = map[string][]Applier{}

// RegisterThenSteps registers steps as a continuation under key. Call it
// once, from a package-level var. Panics on a bad or duplicate key. The
// key is an on-disk identity (ledger line "steps <key>"): never renamed,
// never reused.
func RegisterThenSteps(key string, steps ...Applier) ThenSteps {
	game.RegisterThenStepsKey(key)
	thenStepsByKey[key] = steps
	return ThenSteps{key: key, steps: steps}
}

// Key is the continuation's on-disk key.
func (t ThenSteps) Key() string { return t.key }

// Run runs the steps in order, parking the rest behind a prompt a step
// asked while the resolution is open. The context's item is the one
// whose resolution the continuation belongs to; with none (an AsEnters
// context) the steps run back to back.
func (t ThenSteps) Run(ctx *Context) error {
	return runStepsFrom(ctx, t.steps, nil, parkThenSteps(t.key))
}

// Apply is Run, so a continuation can be a step.
func (t ThenSteps) Apply(ctx *Context) error { return t.Run(ctx) }

// parkThenSteps is a continuation's site.
func parkThenSteps(key string) stepParker {
	return func(ctx *Context, path []int, mark game.StepWalkMark) bool {
		return ctx.Game.ParkThenStepsForEffect(ctx.Item, key, ctx.occurrence, path, mark)
	}
}

// resumeThenSteps is game.ResumeThenSteps: the registered steps from
// path. A key this binary no longer has runs nothing (a restore refuses
// a file naming one, so it is not reached).
func resumeThenSteps(g *game.Game, item *game.StackItem, key string, occ int, path []int) error {
	steps, ok := thenStepsByKey[key]
	if !ok || item == nil || len(path) == 0 {
		return nil
	}
	ctx := NewContext(g, item)
	ctx.occurrence = occ
	return runStepsFrom(ctx, steps, path, parkThenSteps(key))
}

func init() {
	game.ResumeThenSteps = resumeThenSteps
}

// RecordStep files ids and numbers under key, for a later step of the
// same resolution (#4163). It replaces what was there.
func (c *Context) RecordStep(key string, ids []uuid.UUID, ns ...int) {
	c.Game.RecordStepForEffect(c.Item, key, game.StepRecord{IDs: ids, Ns: ns})
}

// StepRecord is what an earlier step filed under key; the zero record
// when none did.
func (c *Context) StepRecord(key string) game.StepRecord {
	r, _ := c.Game.StepRecordForEffect(c.Item, key)
	return r
}
