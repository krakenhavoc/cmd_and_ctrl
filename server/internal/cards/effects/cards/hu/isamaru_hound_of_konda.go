package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Isamaru, Hound of Konda — Legendary Creature — Dog {W}, 2/2 (Jumpstart 2022):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see grizzly_bears.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cfb756dd-6fe1-430a-8f04-8b10cac62a86",
		Name:         "Isamaru, Hound of Konda",
		Completeness: effects.CompletenessFull,
	})
}
