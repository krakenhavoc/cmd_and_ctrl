package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Auratouched Mage — Creature — Human Wizard {5}{W}, 3/3 (Planechase 2012,
// S82 OLD, #5569):
//
//	"When this creature enters, search your library for an Aura card that could
//	 enchant it. If this creature is still on the battlefield, put that Aura
//	 card onto the battlefield attached to it. Otherwise, reveal the Aura card
//	 and put it into your hand. Then shuffle."
//
// Light-Paws' search into the battlefield attached to the source (CR 303.4f), with
// the Mage checked as the ability resolves: still there, the Aura enters attached;
// gone, the card is revealed and goes to your hand. The library is shuffled either
// way. An Aura "could enchant it" when its enchant clause admits the Mage, judged
// on the Mage on the battlefield, or on the card as it last existed once it has
// left (CR 608.2h). The search may find nothing (CR 701.23b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c7ffc6c-88be-4127-a673-fba8c1aa5fa1",
		Name:         "Auratouched Mage",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Auratouched Mage — search your library for an Aura card that could enchant it", effects.S82OldB431AuratouchedMageSearch),
		},
	})
}
