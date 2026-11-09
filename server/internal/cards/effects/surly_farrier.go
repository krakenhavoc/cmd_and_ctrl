package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Surly Farrier — Creature — Kithkin Citizen {1}{G} (Lorwyn Eclipsed, #4458):
//
//	"{T}: Target creature you control gets +1/+1 and gains vigilance until
//	 end of turn. Activate only as a sorcery."
func init() {
	Register(Spec{
		OracleID:     "ef67c170-68aa-4ada-8cee-261a4dbb1b5d",
		Name:         "Surly Farrier",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{T}: Target creature you control gets +1/+1 and gains vigilance until end of turn. Activate only as a sorcery.",
			Cost:         TapCost(),
			Targets:      TargetCreature("target creature you control", YouControl()),
			SorcerySpeed: true,
			Purpose:      ForTargets(BenefitUntilEOT(0, 1, 1, "vigilance")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (BoostUntilEOT{Target: id, Power: 1, Toughness: 1,
					Label: "Surly Farrier — +1/+1 until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"vigilance"},
					Label: "Surly Farrier — vigilance until end of turn"}.Apply(ctx)
			},
		}},
	})
}
