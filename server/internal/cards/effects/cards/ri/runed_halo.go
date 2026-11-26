package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Runed Halo — Enchantment {W}{W} (Core Set 2021, #5511):
//
//	"As this enchantment enters, choose a card name.
//	 You have protection from the chosen card name. (You can't be targeted,
//	 dealt damage, or enchanted by anything with that name.)"
//
// The name is chosen as it enters (CR 614.12), by its controller, and kept on
// the permanent. The protection is a player keyword derived from it
// (game.ProtectionFromTheChosenCardName, as Serra's Emissary's chosen card
// type): a source is stopped when its own name matches, a spell on the stack
// and a permanent alike, ignoring case. Unanswered it grants nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "14cefb15-d833-4c70-80c9-f2d34b6c846b",
		Name:           "Runed Halo",
		Completeness:   effects.CompletenessFull,
		AsEnters:       effects.ChooseCardNameAsEnters("Runed Halo — choose a card name"),
		PlayerKeywords: []string{game.ProtectionFromTheChosenCardName},
	})
}
