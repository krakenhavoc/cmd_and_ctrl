package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rally Maneuver — Instant {2}{W} (Adventures in the Forgotten Realms):
//
//	"Target creature gets +2/+0 and gains first strike until end of turn.
//	 Up to one other target creature gets +0/+2 and gains lifelink until
//	 end of turn."
//
// Two target clauses: the second is optional and must be a different
// creature (CR 115.3). Each half is one until-end-of-turn record. If the
// first target is illegal at resolution the second still gets its half
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6bbdf771-75d5-4910-bd16-6232e254f7f6",
		Name:         "Rally Maneuver",
		Completeness: effects.CompletenessFull,
		Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0, "first strike"),
			effects.BenefitUntilEOT(1, 0, 2, "lifelink")),
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.Distinct(effects.TargetCreature("up to one other target creature").WithCount(0, 1)),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
				if err := effects.UntilEndOfTurn(ctx, t.ID, nil, "Rally Maneuver — +2/+0 and first strike",
					game.ModifyPTMod(2, 0), game.AddKeywordsMod("first strike")); err != nil {
					return err
				}
			}
			if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
				return effects.UntilEndOfTurn(ctx, t.ID, nil, "Rally Maneuver — +0/+2 and lifelink",
					game.ModifyPTMod(0, 2), game.AddKeywordsMod("lifelink"))
			}
			return nil
		},
	})
}
