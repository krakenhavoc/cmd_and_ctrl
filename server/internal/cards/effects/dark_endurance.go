package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ded4009d-be23-4e36-a7a3-9e07230712ee",
		Name:         "Dark Endurance",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerPump,
			Targets: game.ForTargets(BenefitUntilEOT(0, 2, 0, "indestructible"))},
		Targets: TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(1, "This spell costs {1} less to cast if it targets a blocking creature.",
				BlockingCreature()),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			const label = "Dark Endurance — +2/+0 and indestructible until end of turn"
			if err := (BoostUntilEOT{Target: id, Power: 2, Label: label}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Target: id, Keywords: []string{"indestructible"}, Label: label}.Apply(ctx)
		},
	})
}
