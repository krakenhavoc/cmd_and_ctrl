package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cabal Evangel — Creature — Human Cleric {1}{B}, 2/2 (Dominaria, #5530):
//
//	(no rules text)
//
// No rules text: a vanilla creature, registered so it does not wear the
// "unimplemented" badge.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a4fed66f-ef3c-4547-8612-46e65d63ef1f",
		Name:         "Cabal Evangel",
		Completeness: effects.CompletenessFull,
	})
}
