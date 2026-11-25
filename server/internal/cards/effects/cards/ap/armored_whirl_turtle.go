package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armored Whirl Turtle — Creature — Turtle {2}{U}, 0/5 (Global Series Jiang Yanggu & Mu Yanling, #5528):
//
//	(no rules text)
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "51a886f2-9b0a-4964-9d59-99dc1a68a97c",
		Name:         "Armored Whirl Turtle",
		Completeness: effects.CompletenessFull,
	})
}
