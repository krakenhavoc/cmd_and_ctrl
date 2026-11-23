package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gravedig — Sorcery {1}{B}:
//
//	"Choose one —
//	 • Target player creates a 2/2 black Zombie creature token.
//	 • Return target creature card from your graveyard to your hand.
//	 Entwine {2} (Choose both if you pay the entwine cost.)"
//
// Entwine (CR 702.42, #3866): {2} more takes both. The token is created
// under the TARGET player's control, whoever that is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "c77c3618-bc12-46a8-a4ef-073e359ff3f9",
		Name:          "Gravedig",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Entwine("{2}")},
		Modes: effects.Entwining(effects.ChooseOne(
			effects.ModeDoing("Target player creates a 2/2 black Zombie creature token.",
				effects.TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return effects.CreateToken{Controller: t.ID, Template: effects.BlackZombieToken(), N: 1}.Apply(ctx)
				}),
			effects.ModeDoing("Return target creature card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
				effects.ReturnTheModesGraveyardTargetToHand),
		)),
	})
}
