package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dukhara Scavenger — Creature — Crocodile {5}{B}, 4/6 (Kaladesh, #5539):
//
//	"When this creature enters, you may put target artifact or creature card
//	 from your graveyard on top of your library."
//
// Midnight Scavengers' shape: the target is chosen as the trigger goes on the
// stack and the "you may" is asked as it resolves (CR 603.5), since topdecking
// a card is not always a gain. A card that left the graveyard in response is
// an illegal target and is not moved (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9906b19-912e-4a9d-888f-3c895bbed2c5",
		Name:         "Dukhara Scavenger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.Targeting(
					effects.WhenThisEnters("Dukhara Scavenger — you may put target artifact or creature card from your graveyard on top of your library", effects.PutChosenTargetOnTopOfLibrary),
					effects.TargetCardInGraveyard("target artifact or creature card from your graveyard",
						effects.Or(effects.Artifact(), effects.Creature()), effects.YouOwn())),
				"Dukhara Scavenger — put the card on top of your library?"),
		},
	})
}
