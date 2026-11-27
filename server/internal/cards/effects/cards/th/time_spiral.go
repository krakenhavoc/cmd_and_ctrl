package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Time Spiral — Sorcery {4}{U}{U} (Urza's Saga, S82 OLD b722):
//
//	Exile Time Spiral. Each player shuffles their hand and graveyard into
//	their library, then draws seven cards. You untap up to six lands.
//
// Timetwister's shuffle and draw, the spell exiling itself as it resolves
// and the untap of up to six lands (any lands, chosen as it resolves).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cac55e46-b730-4569-b92c-a4b5922fc20c",
		Name:         "Time Spiral",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.S82OldB722TimeSpiral,
	})
}
