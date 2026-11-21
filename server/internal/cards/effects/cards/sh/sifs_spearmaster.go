package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sif's Spearmaster — Creature — God Warrior Hero {2}{R}, 1/4 (Marvel Super
// Heroes Commander, #4455):
//
//	"{T}: This creature deals damage equal to its power to target opponent."
//
// The power is read as the ability resolves, from the last known
// information if the Spearmaster has left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ae33bde-ccc9-40e2-9957-05cd9725f8cf",
		Name:         "Sif's Spearmaster",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals damage equal to its power to target opponent.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					return effects.DealDamage{
						Source: item.SourceCardID,
						Target: t.ID,
						Amount: effects.PowerNowOrLastKnown(g, item.SourceCardID),
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
