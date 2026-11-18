package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lantern Flare — Instant {1}{W}:
//
//	"Cleave {X}{R}{W} (You may cast this spell for its cleave cost. If
//	 you do, remove the words in square brackets.)
//	 Lantern Flare deals X damage to target creature or planeswalker and
//	 you gain X life. [X is the number of creatures you control.]"
//
// Hard-cast, X is the number of creatures you control as the spell
// resolves (CR 608.2h). Cleaved, the sentence is gone and X is the {X}
// of the cleave cost, announced as it is cast (CR 107.3b), which the
// engine accepts because the alternative cost includes {X}.
//
// The life is gained by the spell's controller even if the target is
// gone, but then every target is illegal and the spell does not resolve
// (CR 608.2b), as on any one-target spell.
func init() {
	Register(Spec{
		OracleID:     "06a1c404-d76f-4206-8740-7ad7488fac33",
		Name:         "Lantern Flare",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		AlternativeCosts: []game.AlternativeCost{
			Cleave("{X}{R}{W}", nil),
		},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				x := ctx.X()
				if !ctx.PaidAltCost("cleave") {
					x = countYouControl(ctx, Creature())
				}
				if x <= 0 {
					return nil
				}
				var target game.TargetRef
				if len(ctx.Item.Targets) > 0 {
					target = ctx.Item.Targets[0]
				}
				if err := (DealDamage{Source: ctx.Source(), Target: target.ID, Amount: x}).Apply(ctx); err != nil {
					return err
				}
				return GainLife{Amount: x}.Apply(ctx)
			}),
		),
	})
}
