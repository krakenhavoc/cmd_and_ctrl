package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gearbane Orangutan — Creature — Ape {2}{R}, 2/2 (Murders at Karlov
// Manor, #4469):
//
//	"Reach
//	 When this creature enters, choose one —
//	 • Destroy up to one target artifact.
//	 • Sacrifice an artifact. If you do, put two +1/+1 counters on this
//	   creature."
//
// Apothecary Stomper's modal enters trigger: the mode is chosen as the
// trigger goes on the stack, with the first bullet's target. "Up to one"
// lets that bullet go on the stack with no target (CR 601.2c), and a
// target gone by resolution is skipped (CR 608.2b).
//
// The second bullet is not optional: as it resolves its controller
// sacrifices an artifact they control (mustSacrificeOne), and the
// counters are that sacrifice's continuation, so they come only once an
// artifact has really gone ("if you do"). With no artifact there is
// nothing to sacrifice and no counters. The counters go on the
// Orangutan only while it is still the object that entered (CR 400.7).
//
// No simplification.
func init() {
	trigger := WhenThisEnters("Gearbane Orangutan — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = ChooseOne(
		ModeDoing("Destroy up to one target artifact.",
			TargetPermanent("up to one target artifact", Artifact()).WithCount(0, 1),
			DestroyTheModesTarget),
		ModeDoing("Sacrifice an artifact. If you do, put two +1/+1 counters on this creature.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return mustSacrificeOne(ctx.Game, item, "Gearbane Orangutan — sacrifice an artifact", Artifact(), gearbaneOrangutanSacrificed)
			}),
	)
	Register(Spec{
		OracleID:        "1f3e7420-86aa-4847-bc2d-21f6bae470c8",
		Name:            "Gearbane Orangutan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}

// gearbaneOrangutanSacrificed puts the two counters on the Orangutan once
// an artifact has really gone, keyed (#4290).
var gearbaneOrangutanSacrificed = RegisterSacrificedOne("prompt-run/gearbane-orangutan-sacrifice", func(ctx *Context, _ uuid.UUID) error {
	return CounterOnThis{Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
})
