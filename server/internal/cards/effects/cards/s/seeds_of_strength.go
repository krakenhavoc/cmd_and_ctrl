package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seeds of Strength — Instant {G}{W} (Ravnica: City of Guilds, S82 OLD batch
// 575, #5600):
//
//	"Target creature gets +1/+1 until end of turn.
//	 Target creature gets +1/+1 until end of turn.
//	 Target creature gets +1/+1 until end of turn."
//
// Three independent target clauses, each +1/+1 as the spell resolves: the same
// creature may be chosen more than once and then gets +1/+1 for each time
// (CR 115.3). A clause whose creature is no longer legal does nothing
// (CR 608.2b) while the others still do.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd3d4d96-1e88-49e1-8276-7f62c22ad91a",
		Name:         "Seeds of Strength",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1), effects.BenefitUntilEOT(1, 1, 1), effects.BenefitUntilEOT(2, 1, 1)), game.AnswerPump),
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.TargetCreature("target creature"),
			effects.TargetCreature("target creature")),
		OnResolve: effects.ResolveSteps(effects.Instruction(func(ctx *effects.Context) error {
			for slot := 0; slot < 3; slot++ {
				t, ok := ctx.ClauseTarget(slot)
				if !ok || t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1, Label: "Seeds of Strength — +1/+1"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		})),
	})
}
