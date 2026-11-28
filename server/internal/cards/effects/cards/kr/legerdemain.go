package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Legerdemain — {2}{U}{U} Sorcery (Tempest, S82 OLD batch 769, #5628):
//
//	"Exchange control of target artifact or creature and another target
//	 permanent that shares one of those types with it. (This effect lasts
//	 indefinitely.)"
//
// Shifting Loyalties' exchange over two clauses, with a relation of its own: the
// second pick is another permanent that is an artifact when the first is, or a
// creature when the first is (SharesAnArtifactOrCreatureTypeWithTarget). A type
// they share that is not artifact or creature (two enchantment creatures that
// share only enchantment) does not count. Judged as the spell is cast (CR
// 601.2c) and again as it resolves (CR 608.2b); the exchange is all or nothing
// (CR 701.12a) and has no duration (CR 611.2a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "107d5f67-1395-444f-b612-c5cc770df956",
		Name:         "Legerdemain",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
			effects.Distinct(effects.SharesAnArtifactOrCreatureTypeWithTarget(
				effects.TargetPermanent("another target permanent that shares one of those types with it", effects.Or(effects.Artifact(), effects.Creature())), 0)),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			a, okA := ctx.ClauseTarget(0)
			b, okB := ctx.ClauseTarget(1)
			if !okA || !okB || len(ctx.LegalTargets()) != 2 {
				return nil
			}
			return effects.ExchangeControl{A: a.ID, B: b.ID, Label: "Legerdemain — exchange control"}.Apply(ctx)
		},
	})
}
