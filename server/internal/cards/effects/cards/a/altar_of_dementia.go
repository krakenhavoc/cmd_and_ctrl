package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Altar of Dementia — Artifact {2}:
//
//	"Sacrifice a creature: Target player mills cards equal to the
//	 sacrificed creature's power."
//
// Greater Good's sacrifice-and-read-power body with a targeted mill.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d64e9152-ef24-4394-aeb0-9c3befc56549",
		Name:         "Altar of Dementia",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a creature: Target player mills cards equal to the sacrificed creature's power.",
			Cost:    effects.SacrificeACreature(),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				victim, ok := effects.FirstLegalPlayerTarget(ctx)
				if !ok {
					return nil
				}
				fed, ok := effects.B17PermanentSacrificedToPay(g, item)
				if !ok {
					return nil
				}
				return effects.MillCards{Player: victim, N: effects.DepartedCreaturePower(g, fed)}.Apply(ctx)
			},
		}},
	})
}
