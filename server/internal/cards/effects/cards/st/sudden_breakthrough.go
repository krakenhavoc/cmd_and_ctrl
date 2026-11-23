package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sudden Breakthrough — Instant {1}{R} (Strixhaven: School of Mages,
// #5502):
//
//	"Target creature gets +2/+0 and gains first strike until end of turn.
//	 Create a Treasure token. (It's an artifact with "{T}, Sacrifice this
//	 token: Add one mana of any color.")"
//
// Charge the Sanctum's +2/+0 and first strike, then the catalog Treasure.
// A target that is gone as the spell resolves fizzles the whole spell,
// Treasure included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5576cd4-6b21-4b29-902c-9c1fbdd7b2be",
		Name:         "Sudden Breakthrough",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0, "first strike")),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: id, Power: 2, Label: "Sudden Breakthrough — +2/+0"}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"first strike"}, Label: "Sudden Breakthrough — first strike"}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(ctx)
		},
	})
}
