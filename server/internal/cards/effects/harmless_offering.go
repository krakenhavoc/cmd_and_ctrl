package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Harmless Offering — Sorcery {2}{R} (Foundations, #4465):
//
//	"Target opponent gains control of target permanent you control."
//
// Two clauses, the opponent then the permanent (Clauses). The gift has no
// stated duration, so it lasts for good (CR 611.2a, GainControl with an
// indefinite duration) through a layer-2 control effect, and the
// opponent can't refuse. It needs both: with either target illegal on
// resolution nothing changes hands (CR 608.2b), and a permanent that left
// and came back is a new object that isn't handed over (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f5d32e5f-066b-42cf-8dd1-072a96275313",
		Name:         "Harmless Offering",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetPlayer("target opponent", Opponent()),
			TargetPermanent("target permanent you control", YouControl()),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			opp, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			gift, ok := ctx.ClauseTarget(1)
			if !ok {
				return nil
			}
			return GainControl{
				Target:     gift.ID,
				Controller: opp.ID,
				Duration:   game.IndefiniteDuration(),
				Label:      "Harmless Offering — target opponent gains control of it",
			}.Apply(ctx)
		},
	})
}
