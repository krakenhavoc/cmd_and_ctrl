package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tribute Mage — Creature — Human Wizard {2}{U}, 2/2 (EDHREC rank
// 2143):
//
//	"When this creature enters, you may search your library for an
//	 artifact card with mana value 2, reveal that card, put it into
//	 your hand, then shuffle."
//
// The Signet tutor — Trinket Mage's sibling one mana value up. An
// ETB with the S22 search chooser (b20TutorOnETB): "you may" is the
// prompt's decline, the filter is artifact cards at exactly mana
// value 2 read off the library card's printed cost, the pick is
// revealed and goes to hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2b51295-8820-425f-a517-55231123c7de",
		Name:         "Tribute Mage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.B20TutorOnETB(
				"Tribute Mage — search for an artifact card with mana value 2",
				"Tribute Mage — an artifact card with mana value 2",
				func(c game.Card) bool { return c.IsArtifact() && c.ManaValue() == 2 },
			),
		},
	})
}
