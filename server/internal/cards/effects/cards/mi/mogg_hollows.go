package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mogg Hollows — Land (Tempest, S82 OLD, #5628):
//
//	"{T}: Add {C}.
//	 {T}: Add {R} or {G}. This land doesn't untap during your next
//	 untap step."
//
// The Tempest "freezing" cycle: a free colourless half and a coloured half whose
// rider (Rule 605.3b: it happens as the mana is added, with no stack) marks the
// land as one that sits out its controller's next untap step. The marker is the
// engine's own next-untap skip (CR 502.3), so it lapses after that one step and
// does nothing to a land that was untapped for some other reason before it.
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB763FreezingLand("1745fd57-467c-45f9-a46e-b9a2af87ec87", "Mogg Hollows", "R", "G"))
}
