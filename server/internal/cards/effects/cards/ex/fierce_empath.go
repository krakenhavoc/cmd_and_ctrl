package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fierce Empath — Creature — Elf {2}{G}, 1/1 (EDHREC rank 2137):
//
//	"When this creature enters, you may search your library for a
//	 creature card with mana value 6 or greater, reveal it, put it
//	 into your hand, then shuffle."
//
// The big-creature tutor on a 1/1 body. An ETB with the S22 search
// chooser (the Trinket Mage shape, b20TutorOnETB): "you may" is the
// prompt's decline, the filter is creature cards at mana value six
// or more read off the library card's printed cost, the pick is
// revealed and goes to hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5104053a-d394-4b00-82a4-60fd9f051a6e",
		Name:         "Fierce Empath",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.B20TutorOnETB(
				"Fierce Empath — search for a creature card with mana value 6 or greater",
				"Fierce Empath — a creature card with mana value 6 or greater",
				func(c game.Card) bool { return c.IsCreature() && c.ManaValue() >= 6 },
			),
		},
	})
}
