package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Disciple of Bolas — Creature — Human Wizard {3}{B}, 2/2 (issue
// #1117):
//
//	"When this creature enters, sacrifice another creature. You gain
//	 X life and draw X cards, where X is that creature's power."
//
// The sacrifice is chosen AT RESOLUTION (the Harrow / Springbloom
// Druid shape, ADR 0013 §5x): the trigger's Effect opens a sacrifice
// prompt over the controller's OTHER creatures and the payout runs as
// its continuation, so the creature really sacrificed — not
// whichever one happened to be biggest when the trigger was built —
// is the one whose power counts. With no other creature to sacrifice
// the prompt never goes up, the continuation still runs (a
// continuation is the rest of a card that is paused mid-resolution,
// never silently skipped), and it gains no life and draws no cards —
// exactly the printed ability doing nothing, not an error.
//
// "That creature's power" is its last-known information (CR 608.2h,
// departedCreaturePower): counters and an anthem's bonus both count,
// as they did while it was on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f2c8498-5b10-4ec7-8678-598c90987556",
		Name:         "Disciple of Bolas",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Disciple of Bolas — sacrifice another creature; gain life and draw cards equal to its power",
				discipleOfBolasSacrificeThenPayout),
		},
	})
}

// discipleOfBolasSacrificeThenPayout is Disciple of Bolas's whole
// ETB ability: a sacrifice prompt over the controller's other
// creatures, and the life-gain-and-draw as its continuation.
//
// Caller holds g.mu (resolution frame).
func discipleOfBolasSacrificeThenPayout(g *game.Game, item *game.StackItem) error {
	return NewContext(g, item).SacrificeThen(item.Controller,
		sacrificeSpec("another creature", Creature(), NotSelf(item.SourceCardID)),
		"Disciple of Bolas — sacrifice another creature", 1, discipleOfBolasPayout)
}

// discipleOfBolasPayout is the life gain and the draw once the creature has
// really been sacrificed, keyed (#4290): the item rides the run as data, so
// the creature's power is filed as the step record under the same item.
var discipleOfBolasPayout = RunThen("prompt-run/disciple-of-bolas-payout", func(ctx *Context, r game.RunDone) error {
	ids := r.Sacrifices().By(ctx.Controller())
	if len(ids) == 0 {
		// No other creature to sacrifice — the ability does
		// as much as it can, which is nothing.
		return nil
	}
	ctx.RecordStep(sacrificedPowerRecord, nil, departedCreaturePower(ctx.Game, ids[0]))
	return gainThenDrawSacrificedPower.Run(ctx)
})

// sacrificedPowerRecord is the sacrificed creature's last-known power,
// handed from a sacrifice's continuation to the steps after it.
const sacrificedPowerRecord = "sacrificed-power"

// gainThenDrawSacrificedPower is "you gain X life and draw X cards, where
// X is that creature's power" as continuation steps (#4162, CR 608.2c):
// the draw waits for a prompt the life gain asks (a CR 616 order of two
// life-gain replacements). Disciple of Bolas and Disciple of Freyalise.
var gainThenDrawSacrificedPower = RegisterThenSteps("then-steps/gain-then-draw-sacrificed-power",
	Instruction(func(ctx *Context) error {
		if n := ctx.StepRecord(sacrificedPowerRecord).N(); n > 0 {
			return GainLife{Amount: n}.Apply(ctx)
		}
		return nil
	}),
	Instruction(func(ctx *Context) error {
		if n := ctx.StepRecord(sacrificedPowerRecord).N(); n > 0 {
			return DrawCards{N: n}.Apply(ctx)
		}
		return nil
	}),
)
