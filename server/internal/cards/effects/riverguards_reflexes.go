package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Riverguard's Reflexes — Instant {1}{W} (Lorwyn Eclipsed, #4458):
//
//	"Target creature gets +2/+2 and gains first strike until end of turn.
//	 Untap it."
//
// Three steps in printed order on the one target, each skipped by the
// same CR 608.2b legality check (a target that left or became illegal
// makes the whole spell do nothing).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0bdebad7-a05e-439a-b5a9-241a2d09f943",
		Name:         "Riverguard's Reflexes",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose: game.Purpose{
			Answers: game.AnswerPump | game.AnswerCombatGrant,
			Targets: game.ForTargets(BenefitUntilEOT(0, 2, 2, "first strike")),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Riverguard's Reflexes — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"first strike"},
				Label: "Riverguard's Reflexes — first strike"}).Apply(ctx); err != nil {
				return err
			}
			return UntapTarget{Target: id}.Apply(ctx)
		},
	})
}
