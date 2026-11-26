package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Veinfire Borderpost — Artifact {1}{B}{R} (Alara Reborn, #5586): the Borderpost
// cycle (S82OldB503Borderpost) in black and red.
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB503Borderpost("3f65d1f8-3e46-4e3d-b087-5496a4106e4e", "Veinfire Borderpost", "{1}{B}{R}", "B", "R"))
}
