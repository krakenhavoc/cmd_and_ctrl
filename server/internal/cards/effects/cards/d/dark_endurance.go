package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dark Endurance — Instant {1}{B} (Edge of Eternities, #4461):
//
//	"This spell costs {1} less to cast if it targets a blocking creature.
//	 Target creature gets +2/+0 and gains indestructible until end of
//	 turn."
//
// The discount is Run Behind's CostsLessIfItTargets, over a blocking
// creature (CR 509.1g) instead of an attacking one, judged against the
// announced target as the cost is determined (CR 601.2f). Both halves of
// the effect land on the creature while it is still a legal target
// (CR 608.2b); one that left in response gets neither.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ded4009d-be23-4e36-a7a3-9e07230712ee",
		Name:         "Dark Endurance",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerPump,
			Targets: game.ForTargets(effects.BenefitUntilEOT(0, 2, 0, "indestructible"))},
		Targets: effects.TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(1, "This spell costs {1} less to cast if it targets a blocking creature.",
				effects.BlockingCreature()),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			const label = "Dark Endurance — +2/+0 and indestructible until end of turn"
			if err := (effects.BoostUntilEOT{Target: id, Power: 2, Label: label}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"indestructible"}, Label: label}.Apply(ctx)
		},
	})
}
