package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cinder Marsh — Land (Tempest, S82 OLD, #5628):
//
//	"{T}: Add {C}.
//	 {T}: Add {B} or {R}. This land doesn't untap during your next
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
	effects.Register(effects.S82OldB763FreezingLand("6f8cc374-e76c-4bfa-bf20-28dea0bfefbe", "Cinder Marsh", "B", "R"))
}
