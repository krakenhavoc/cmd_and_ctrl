package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "71ae3bc0-28da-4e91-bffd-a49fe545b7a9",
		Name:         "Treasure Hunter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.WhenThisEnters("Treasure Hunter — return an artifact card to your hand",
						effects.ReturnFirstLegalGraveyardTargetToHand),
					"Treasure Hunter — return an artifact card from your graveyard to your hand?"),
				effects.TargetCardInGraveyard("target artifact card from your graveyard", effects.YouOwn(), effects.Artifact())),
		},
	})
}
