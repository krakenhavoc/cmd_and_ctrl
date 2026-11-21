package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moment of Defiance — Instant for {2}{B} (The Brothers' War):
//
//	"Target creature gets +2/+1 and gains lifelink until end of turn.
//	Draw a card."
func init() {
	const label = "Moment of Defiance — +2/+1 and lifelink until end of turn"
	p := effects.ForTargets(effects.BenefitUntilEOT(0, 2, 1, "lifelink"))
	p.Draws = 1
	effects.Register(effects.Spec{
		OracleID:     "257f81fb-b3c9-4ebe-9889-dd0ef960bba3",
		Name:         "Moment of Defiance",
		Purpose:      p,
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
					return nil
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 1, Label: label}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"lifelink"}, Label: label}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
			}),
		),
	})
}
