package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Idyllic Tutor — Sorcery {2}{W} (EDHREC rank 751):
//
//	"Search your library for an enchantment card, reveal it, put it
//	 into your hand, then shuffle."
//
// The enchantress deck's Demonic Tutor. Type-filtered search to hand,
// revealed, through the S22 chooser.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "57c9ff89-dd30-467b-bb3e-499eeea8cb94",
		Name:         "Idyllic Tutor",
		Completeness: CompletenessFull,
		OnResolve:    b06TutorToHand("Idyllic Tutor — an enchantment card", func(c game.Card) bool { return c.IsEnchantment() }),
	})
}
