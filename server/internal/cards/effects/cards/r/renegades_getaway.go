package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Renegade's Getaway — {2}{B} Instant (Aether Revolt, #5536):
//
//	"Target permanent gains indestructible until end of turn. Create a 1/1
//	 colorless Servo artifact creature token."
//
// With the target gone the whole spell is countered on resolution
// (CR 608.2b), so no Servo is made.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b77479e-8d49-46f6-b774-d8caf151c15c",
		Name:         "Renegade's Getaway",
		Completeness: effects.CompletenessFull,
		Purpose:      renegadesGetawayPurpose(),
		Targets:      effects.TargetPermanent("target permanent"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.GrantKeywordUntilEOT{
							Target:   t.ID,
							Keywords: []string{"indestructible"},
							Label:    "Renegade's Getaway — indestructible until end of turn",
						}.Apply(ctx)
					}
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.CreateToken{Template: effects.TokenCard("1/1 colorless Servo artifact"), N: 1}.Apply(ctx)
			}),
		),
	})
}

func renegadesGetawayPurpose() game.Purpose {
	p := effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible"))
	p.Answers = game.AnswerSurvive
	p.Tokens, p.CreatureTokens = 1, 1
	return p
}
