package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "57b75be4-9229-4a48-a59e-a0f812f38637",
		Name:         "Diplomatic Relations",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{Answers: game.AnswerRemove, // a one-sided fight, as Bite Down (#4734)
			Targets: game.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "vigilance"))},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				mine, ok := ctx.ClauseTarget(0)
				if !ok || mine.Kind != game.TargetCard {
					return nil
				}
				const label = "Diplomatic Relations — +1/+0 and vigilance until end of turn"
				if err := (effects.BoostUntilEOT{Target: mine.ID, Power: 1, Label: label}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: mine.ID, Keywords: []string{"vigilance"}, Label: label}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error { return effects.OneSidedBite(ctx.Item, ctx) }),
		),
	})
}
