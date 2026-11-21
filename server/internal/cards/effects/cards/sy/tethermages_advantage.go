package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tethermage's Advantage — Instant {G} (Reality Fracture, tracker #2795):
//
//	"Target creature gets +2/+2 and gains reach until end of turn.
//	 Untap it."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b4f89885-eeb9-46c7-9820-9758b624ca24",
		Name:         "Tethermage's Advantage",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "reach")),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Tethermage's Advantage — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"reach"}, Label: "Tethermage's Advantage — reach"}).Apply(ctx); err != nil {
					return err
				}
				return effects.UntapTarget{Target: t.ID}.Apply(ctx)
			}
			return nil
		},
	})
}
