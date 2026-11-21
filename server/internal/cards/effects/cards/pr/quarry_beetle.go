package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quarry Beetle — {4}{G} Creature — Insect 4/5 (Jumpstart 2022, #5948):
//
//	"When this creature enters, you may return target land card from your
//	 graveyard to the battlefield."
//
// Gravedigger's optional target, returning a land card from your own
// graveyard to the battlefield untapped; the target is re-checked as the
// trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fff66029-24bb-4354-a650-99a205fd168f",
		Name:         "Quarry Beetle",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("Quarry Beetle — return target land card from your graveyard to the battlefield",
					effects.ReturnTargetedGraveyardCardToBattlefield),
				effects.TargetCardInGraveyard("target land card from your graveyard", effects.Land(), effects.YouOwn())),
				"Quarry Beetle — return it to the battlefield?"),
		},
	})
}
