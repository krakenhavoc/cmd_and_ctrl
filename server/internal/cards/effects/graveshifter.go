package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Graveshifter — Creature — Shapeshifter {3}{B}, 2/2 (Lorwyn Eclipsed,
// #4458):
//
//	"Changeling (This card is every creature type.)
//	 When this creature enters, you may return target creature card from
//	 your graveyard to your hand."
//
// Witch's Cottage's optional, targeted enters trigger with the return to
// the hand: the controller is asked whether to use it, picks a creature
// card in their own graveyard as it goes on the stack (CR 603.3d), and a
// card that left the graveyard in response is not returned (CR 608.2b).
// Changeling is the keyword the engine honours (CR 702.73a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1eebc02a-9a64-49b0-8322-c5ce671457ce",
		Name:            "Graveshifter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"changeling"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				Optional(
					WhenThisEnters("Graveshifter — return target creature card from your graveyard to your hand",
						returnFirstLegalGraveyardTargetToHand),
					"Graveshifter — return target creature card from your graveyard to your hand?"),
				TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn()),
			),
		},
	})
}
