package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Magic Damper — Instant {U} (Final Fantasy, #4462):
//
//	"Target creature you control gets +1/+1 and gains hexproof until end
//	 of turn. Untap it."
//
// Acrobatic Leap's shape on a creature you control: the still-legal
// target (CR 608.2b) gets the boost and hexproof, then untaps. A target
// that left in response gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ada82f93-dc93-4a95-a13f-42b5942b3413",
		Name:         "Magic Damper",
		Purpose:      effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "hexproof")), game.AnswerProtect|game.AnswerPump),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target := effects.FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: target, Power: 1, Toughness: 1, Label: "Magic Damper — +1/+1"}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"hexproof"}, Label: "Magic Damper — hexproof"}).Apply(ctx); err != nil {
				return err
			}
			return effects.UntapTarget{Target: target}.Apply(ctx)
		},
	})
}
