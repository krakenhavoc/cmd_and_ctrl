package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Structural Distortion — Sorcery {3}{R} (Shadows over Innistrad, S82 OLD b293):
//
//	"Exile target artifact or land. Structural Distortion deals 2 damage to
//	 that permanent's controller."
//
// The controller is read before the exile (last-known information, CR
// 608.2h), so the damage lands on whoever controlled the permanent as it
// left. A target that is illegal on resolution does nothing at all (CR
// 608.2b), the damage included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bf0b6800-8e6b-4ccd-8e4d-96e61342f981",
		Name:         "Structural Distortion",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact or land", effects.Or(effects.Artifact(), effects.Land())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					continue
				}
				controller := c.Controller
				if err := (effects.ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: controller, Amount: 2}.Apply(ctx)
			}
			return nil
		},
	})
}
