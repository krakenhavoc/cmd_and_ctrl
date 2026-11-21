package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Expedition Envoy — Creature — Human Scout Ally {W}, 2/1 (Battle for
// Zendikar, #5547):
//
//	(no rules text)
//
// A vanilla body: no rules text, nothing to register beyond the card
// itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8bc7c769-33e0-4482-95fc-b344b772e174",
		Name:         "Expedition Envoy",
		Completeness: effects.CompletenessFull,
	})
}
