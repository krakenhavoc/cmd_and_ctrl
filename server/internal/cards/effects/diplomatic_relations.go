package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Diplomatic Relations — Instant {2}{G} (Edge of Eternities, #4461):
//
//	"Target creature you control gets +1/+0 and gains vigilance until end
//	 of turn. It deals damage equal to its power to target creature an
//	 opponent controls."
//
// Nature's Way's two clauses, read positionally: slot 0 is your creature,
// which gets +1/+0 and vigilance and then deals damage equal to its power
// (the pump included, read as the spell resolves), slot 1 is the creature
// it hits. Each clause is re-checked on its own (CR 608.2b): a victim that
// left in response leaves the pump and vigilance in place, and a creature
// of yours that left gets nothing and deals nothing. The damage is the
// creature's, so its deathtouch and lifelink apply.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "57b75be4-9229-4a48-a59e-a0f812f38637",
		Name:         "Diplomatic Relations",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{Answers: game.AnswerRemove, // a one-sided fight, as Bite Down (#4734)
			Targets: game.ForTargets(BenefitUntilEOT(0, 1, 0, "vigilance"))},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				mine, ok := ctx.ClauseTarget(0)
				if !ok || mine.Kind != game.TargetCard {
					return nil
				}
				const label = "Diplomatic Relations — +1/+0 and vigilance until end of turn"
				if err := (BoostUntilEOT{Target: mine.ID, Power: 1, Label: label}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: mine.ID, Keywords: []string{"vigilance"}, Label: label}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error { return oneSidedBite(ctx.Item, ctx) }),
		),
	})
}
