package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spidery Grasp — Instant {2}{G} (Game Night 2019, #5515):
//
//	"Untap target creature. It gets +2/+4 and gains reach until end of turn."
//
// Aim High's three steps in order: the untap, the boost and the keyword all
// land on the target if it is still legal as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "021ce403-cbfa-4d69-a6e9-473d58bd477a",
		Name:         "Spidery Grasp",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 4, "reach")),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 4, Label: "Spidery Grasp — +2/+4"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"reach"}, Label: "Spidery Grasp — reach"}.Apply(ctx)
		},
	})
}
