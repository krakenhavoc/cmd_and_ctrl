package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ferocious Zheng — Creature — Cat Beast {2}{G}{G}, 4/4 (Global Series Jiang Yanggu & Mu Yanling, #5528):
//
//	(no rules text)
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a45e0854-28c7-41f5-a9fe-5b76a8070c5b",
		Name:         "Ferocious Zheng",
		Completeness: effects.CompletenessFull,
	})
}
