package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leopard-Spotted Jiao — Creature — Beast {1}{R}, 3/1 (Global Series Jiang Yanggu & Mu Yanling, #5528):
//
//	(no rules text)
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0056e07b-416b-487e-9e6c-3697db402dd4",
		Name:         "Leopard-Spotted Jiao",
		Completeness: effects.CompletenessFull,
	})
}
