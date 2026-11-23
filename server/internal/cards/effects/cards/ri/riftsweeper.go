package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Riftsweeper — Creature — Elf Shaman {1}{G}, 2/2 (Future Sight, #5594):
//
//	"When this creature enters, choose target face-up exiled card. Its
//	 owner shuffles it into their library."
//
// Pull from Eternity's target clause (a face-up card in exile, which a
// suspended card is) on an enters trigger, and the shared "the owners
// shuffle them into their libraries" body: the card goes to its OWNER's
// library and that owner shuffles.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ba0dce5-e767-47cc-9967-174e52bce120",
		Name:         "Riftsweeper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Riftsweeper — shuffle target face-up exiled card into its owner's library",
					effects.TrekB9ShuffleTargetsIntoOwnersLibraries),
				effects.TargetCardInExile("target face-up exiled card", effects.FaceUp())),
		},
	})
}
