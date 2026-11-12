package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rally the Righteous — Instant {1}{R}{W}:
//
//	"Radiance — Untap target creature and each other creature that
//	 shares a color with it. Those creatures get +2/+0 until end of
//	 turn."
//
// The creatures are read once, as the spell resolves (radianceSet,
// #5736), and "those creatures" are that same set: each is untapped, in
// the printed order, then all of them get +2/+0 as one effect locked now
// (CR 611.2c), an already-untapped one included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5d4d064e-c4f4-4eca-955c-4539e66d230b",
		Name:         "Rally the Righteous",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump, Targets: game.ForTargets(BenefitUntilEOT(0, 2, 0))},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			ids := radianceSet(ctx, Creature())
			if len(ids) == 0 {
				return nil
			}
			if err := eachRadiant(ctx, ids, func(c *Context, id uuid.UUID) error {
				return UntapTarget{Target: id}.Apply(c)
			}); err != nil {
				return err
			}
			return BoostUntilEOT{Match: b10OneOf(ids), Power: 2, Label: "Rally the Righteous — +2/+0"}.Apply(ctx)
		},
	})
}
