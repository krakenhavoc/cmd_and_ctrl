package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mistvein Borderpost — {1}{U}{B} Artifact (Archenemy, S82 OLD #5580). The
// Borderpost cycle: see OldB477Borderpost.
func init() {
	effects.Register(effects.OldB477Borderpost("1499af11-274f-4f8d-b0b6-20e1693ebab7", "Mistvein Borderpost", "U", "B"))
}
