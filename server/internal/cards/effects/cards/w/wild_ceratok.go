package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wild Ceratok — Creature — Rhino {3}{G}, 4/3 (Guilds of Ravnica, #5525):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4dc82a49-a063-448d-bd2c-a5b69381e27e",
		Name:         "Wild Ceratok",
		Completeness: effects.CompletenessFull,
	})
}
