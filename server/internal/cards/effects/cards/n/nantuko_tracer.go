package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nantuko Tracer — Creature — Insect Druid {1}{G}, 2/1 (Judgment, #5611):
//
//	"When this creature enters, you may put target card from a graveyard
//	 on the bottom of its owner's library."
//
// An optional targeted enters trigger; the card goes to the bottom of its
// OWNER's library (CR 401.3 puts it there for its owner, whichever
// graveyard it was in).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4fd567cb-0b8e-41dd-a2b7-87a196de294a",
		Name:         "Nantuko Tracer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Optional(effects.WhenThisEnters("Nantuko Tracer — put target card from a graveyard on the bottom of its owner's library", effects.OldB621CardToBottomOfLibrary),
				"Nantuko Tracer — put target card from a graveyard on the bottom of its owner's library?"),
				effects.TargetCardInGraveyard("target card from a graveyard")),
		},
	})
}
