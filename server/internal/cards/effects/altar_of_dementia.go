package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Altar of Dementia — Artifact {2}:
//
//	"Sacrifice a creature: Target player mills cards equal to the
//	 sacrificed creature's power."
//
// Greater Good's sacrifice-and-read-power body with a targeted mill.
func init() {
	Register(Spec{
		OracleID:     "d64e9152-ef24-4394-aeb0-9c3befc56549",
		Name:         "Altar of Dementia",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a creature: Target player mills cards equal to the sacrificed creature's power.",
			Cost:    SacrificeACreature(),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				victim, ok := firstLegalPlayerTarget(ctx)
				if !ok {
					return nil
				}
				fed, ok := b17PermanentSacrificedToPay(g, item)
				if !ok {
					return nil
				}
				return MillCards{Player: victim, N: departedCreaturePower(g, fed)}.Apply(ctx)
			},
		}},
	})
}
