package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marshaling the Troops — Sorcery {1}{G} (Portal Three Kingdoms, S82 OLD b714, #5622):
//
//	"Tap any number of untapped creatures you control. You gain 4 life for each creature tapped this way."
//
// Devout Invocation's pick with life instead of tokens: the creatures are
// chosen as the spell resolves, any number, none included, and only a
// creature actually tapped gains life (MarshalingTheTroopsResolve).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f82c628c-c01a-4a1c-af9a-e296ba50f12f",
		Name:         "Marshaling the Troops",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.MarshalingTheTroopsResolve,
	})
}
