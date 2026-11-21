package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coati Scavenger — Creature — Raccoon {2}{G}, 3/2 (The Lost Caverns of
// Ixalan, #4470):
//
//	"Descend 4 — When this creature enters, if there are four or more
//	 permanent cards in your graveyard, return target permanent card from
//	 your graveyard to your hand."
//
// Dundoolin Weaver's return behind a descend 4 intervening if
// (WhenThisEntersIfDescend, CR 603.4): with fewer than four permanent
// cards (CR 110.4) in its controller's graveyard as it enters it does not
// trigger, and the count is read again as the trigger resolves. The
// target is chosen as the trigger goes on the stack (CR 603.3d) and may
// be one of the four cards counted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "535f9bc6-9a07-4850-91eb-c00d06633e7e",
		Name:         "Coati Scavenger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEntersIfDescend(4, "Coati Scavenger — return target permanent card from your graveyard to your hand",
				effects.ReturnTargetGraveyardCardToHand),
				effects.TargetCardInGraveyard("target permanent card from your graveyard", effects.Permanent(), effects.YouOwn())),
		},
	})
}
