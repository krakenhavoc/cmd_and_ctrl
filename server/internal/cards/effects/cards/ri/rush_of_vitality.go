package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rush of Vitality — Instant {1}{B} (Kaladesh, #5539):
//
//	"Target creature gets +1/+0 and gains lifelink and indestructible until end
//	 of turn."
//
// Three until-end-of-turn effects on the one target, each checked against it as
// the spell resolves (CR 608.2b): a layer 7c pump and a layer 6 grant of both
// keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e460d39-0f18-4a2d-9c8d-3eaba301d285",
		Name:         "Rush of Vitality",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "lifelink", "indestructible")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			const label = "Rush of Vitality"
			if err := (effects.BoostUntilEOT{Target: id, Power: 1, Label: label + " — +1/+0"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"lifelink", "indestructible"}, Label: label + " — lifelink and indestructible"}.Apply(ctx)
		},
	})
}
