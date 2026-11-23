package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elite Vanguard — Creature — Human Soldier {W}, 2/1 (Magic 2010, #5585):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b3b9a87d-cb95-435c-90b6-037406cab32e",
		Name:         "Elite Vanguard",
		Completeness: effects.CompletenessFull,
	})
}
