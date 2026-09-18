package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Anointed Procession — Enchantment {3}{W} (EDHREC rank 365):
//
//	"If an effect would create one or more tokens under your control,
//	 it creates twice that many of those tokens instead."
//
// Parallel Lives in white, word for word, and therefore the same
// declared effect in a different card file: the CR 701.7b creation
// event's counts double (#762).
//
// Two Anointed Processions are ×4 with no prompt — one declared
// effect on two objects, the #792 identical-window skip — and an
// Anointed Procession beside a Parallel Lives is ×4 with one, because
// two DIFFERENT declared effects are ordered by the affected player
// even when every ordering agrees.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "7246d45b-2185-4cdd-981b-5419b7d52bce",
		Name:         "Anointed Procession",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			TokensDoubled("Anointed Procession: double tokens"),
		},
	})
}
