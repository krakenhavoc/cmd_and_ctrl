package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Return of the Nightstalkers — Sorcery {5}{B}{B} (Portal Second Age, S82 older sets batch 735, #5625):
//
//	"Return all Nightstalker permanent cards from your graveyard to the battlefield. Then destroy all Swamps you control."
//
// The cards enter together; the Swamps are destroyed once they have, so a returned Swamp is destroyed too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6082f66e-3641-4b09-bb04-86bf29ee1c6f",
		Name:         "Return of the Nightstalkers",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.S82OldB735ReturnNightstalkers,
	})
}
