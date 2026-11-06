package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "ada82f93-dc93-4a95-a13f-42b5942b3413",
		Name:         "Magic Damper",
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 1, "hexproof")),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target := FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			if err := (BoostUntilEOT{Target: target, Power: 1, Toughness: 1, Label: "Magic Damper — +1/+1"}).Apply(ctx); err != nil {
				return err
			}
			if err := (GrantKeywordUntilEOT{Target: target, Keywords: []string{"hexproof"}, Label: "Magic Damper — hexproof"}).Apply(ctx); err != nil {
				return err
			}
			return UntapTarget{Target: target}.Apply(ctx)
		},
	})
}
