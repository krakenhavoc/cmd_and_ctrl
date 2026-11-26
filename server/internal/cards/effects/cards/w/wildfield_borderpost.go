package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wildfield Borderpost — Artifact {1}{G}{W} (Alara Reborn, #5586): the Borderpost
// cycle (S82OldB503Borderpost) in green and white.
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB503Borderpost("b8141489-00fa-4aba-85d0-4c883196fece", "Wildfield Borderpost", "{1}{G}{W}", "G", "W"))
}
