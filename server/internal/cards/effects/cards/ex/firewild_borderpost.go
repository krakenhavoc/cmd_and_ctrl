package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Firewild Borderpost — Artifact {1}{R}{G} (Alara Reborn, #5586): the Borderpost
// cycle (S82OldB503Borderpost) in red and green.
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB503Borderpost("312a5a75-9b68-4e4b-97fe-3ab644e3e493", "Firewild Borderpost", "{1}{R}{G}", "R", "G"))
}
