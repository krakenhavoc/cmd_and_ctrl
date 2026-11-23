package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nature's Way — Sorcery {1}{G} (Jumpstart):
//
//	"Target creature you control gains vigilance and trample until end of
//	 turn. It deals damage equal to its power to target creature you
//	 don't control."
//
// Two target clauses, read positionally as Bite Down's are: slot 0 is
// your creature, which gains vigilance and trample and then deals
// damage equal to its power as the spell resolves; slot 1 is the
// creature it hits. Each clause is re-checked on its own (CR 608.2b):
// a victim that left in response leaves the keywords in place, and a
// creature of yours that left gains nothing and deals nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c04663a-1023-49a5-a0b9-78f24a72fb1e",
		Name:         "Nature's Way",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{Answers: game.AnswerRemove, // a one-sided fight, as Bite Down (#4734)
			Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "vigilance", "trample"))},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				mine, ok := ctx.ClauseTarget(0)
				if !ok || mine.Kind != game.TargetCard {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: mine.ID, Keywords: []string{"vigilance", "trample"},
					Label: "Nature's Way — vigilance and trample"}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error { return effects.OneSidedBite(ctx.Item, ctx) }),
		),
	})
}
