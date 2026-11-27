package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fieldmist Borderpost — {1}{W}{U} Artifact (Archenemy, S82 OLD #5580). The
// Borderpost cycle: see OldB477Borderpost.
func init() {
	effects.Register(effects.OldB477Borderpost("a292ca5b-2472-4b02-a26b-d1b5b717a34b", "Fieldmist Borderpost", "W", "U"))
}
