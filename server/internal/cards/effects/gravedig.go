package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "c77c3618-bc12-46a8-a4ef-073e359ff3f9",
		Name:          "Gravedig",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Entwine("{2}")},
		Modes: Entwining(ChooseOne(
			ModeDoing("Target player creates a 2/2 black Zombie creature token.",
				TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return CreateToken{Controller: t.ID, Template: BlackZombieToken(), N: 1}.Apply(ctx)
				}),
			ModeDoing("Return target creature card from your graveyard to your hand.",
				TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
				ReturnTheModesGraveyardTargetToHand),
		)),
	})
}
