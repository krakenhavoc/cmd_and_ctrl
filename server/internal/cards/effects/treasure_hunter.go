package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Treasure Hunter — Creature — Human {2}{W}, 2/2 (Exodus, #5626):
//
//	"When this creature enters, you may return target artifact card from
//	 your graveyard to your hand."
//
// Auramancer's trigger with an artifact card: an optional enters
// trigger (CR 603.5) targeting an artifact card in your own graveyard,
// returned by the shared returnFirstLegalGraveyardTargetToHand. A card
// that left the graveyard in response is not returned (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "71ae3bc0-28da-4e91-bffd-a49fe545b7a9",
		Name:         "Treasure Hunter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				Optional(
					WhenThisEnters("Treasure Hunter — return an artifact card to your hand",
						returnFirstLegalGraveyardTargetToHand),
					"Treasure Hunter — return an artifact card from your graveyard to your hand?"),
				TargetCardInGraveyard("target artifact card from your graveyard", YouOwn(), Artifact())),
		},
	})
}
