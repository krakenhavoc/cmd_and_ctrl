package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Transit Mage — Creature — Human Wizard {2}{U}, 2/2 (Aetherdrift,
// #4464):
//
//	"When this creature enters, you may search your library for an
//	 artifact card with mana value 4 or 5, reveal it, put it into your
//	 hand, then shuffle."
//
// Tribute Mage's tutor (b20TutorOnETB) two mana values up and a band two
// wide: "you may" is the prompt's decline, the filter reads the library
// card's printed mana value, and the pick is revealed and goes to hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "991efbf1-299d-4892-aeab-8ff9827f7168",
		Name:         "Transit Mage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			b20TutorOnETB(
				"Transit Mage — search for an artifact card with mana value 4 or 5",
				"Transit Mage — an artifact card with mana value 4 or 5",
				func(c game.Card) bool { return c.IsArtifact() && (c.ManaValue() == 4 || c.ManaValue() == 5) },
			),
		},
	})
}
