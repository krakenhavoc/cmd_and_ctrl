package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chorus of Might — Instant {3}{G} (Return to Ravnica, S82 OLD):
//
//	"Until end of turn, target creature gets +1/+1 for each creature you
//	 control and gains trample."
//
// The bonus counts the creatures its controller controls as the spell
// resolves, the target included if it is theirs, and is fixed from then on
// (CR 611.2c). With the target illegal the whole spell does nothing (CR
// 608.2b).
//
// No simplification.
func init() {
	purpose := effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "trample"))
	purpose.Answers = game.AnswerPump | game.AnswerCombatGrant
	effects.Register(effects.Spec{
		OracleID:     "0765c18b-25a0-4f79-aa1e-7adba7f5c91e",
		Name:         "Chorus of Might",
		Completeness: effects.CompletenessFull,
		Purpose:      purpose,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target := effects.FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			n := effects.CountControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsCreature() })
			if err := (effects.BoostUntilEOT{Target: target, Power: n, Toughness: n, Label: "Chorus of Might — +1/+1 for each creature you control"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"trample"}, Label: "Chorus of Might — trample until end of turn"}.Apply(ctx)
		},
	})
}
