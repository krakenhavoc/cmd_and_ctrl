package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Surly Farrier — Creature — Kithkin Citizen {1}{G} (Lorwyn Eclipsed, #4458):
//
//	"{T}: Target creature you control gets +1/+1 and gains vigilance until
//	 end of turn. Activate only as a sorcery."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef67c170-68aa-4ada-8cee-261a4dbb1b5d",
		Name:         "Surly Farrier",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}: Target creature you control gets +1/+1 and gains vigilance until end of turn. Activate only as a sorcery.",
			Cost:         effects.TapCost(),
			Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
			SorcerySpeed: true,
			Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "vigilance")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1,
					Label: "Surly Farrier — +1/+1 until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"vigilance"},
					Label: "Surly Farrier — vigilance until end of turn"}.Apply(ctx)
			},
		}},
	})
}
