package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bane Alley Blackguard — Creature — Human Rogue {1}{B}, 1/3:
//
//	(no rules text)
//
// A vanilla creature: the printed 1/3 body is the whole card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1134eea5-e931-4cb2-9e30-64c227386d43",
		Name:         "Bane Alley Blackguard",
		Completeness: effects.CompletenessFull,
	})
}
