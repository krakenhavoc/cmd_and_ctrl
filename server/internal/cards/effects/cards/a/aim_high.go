package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aim High — Instant {1}{G} (Innistrad Remastered):
//
//	"Untap target creature. It gets +2/+2 and gains reach until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c169929a-850b-4d46-9f14-a922f4465d85",
		Name:         "Aim High",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "reach")),
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
			if err := (effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Aim High — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"reach"}, Label: "Aim High — reach"}.Apply(ctx)
		},
	})
}
