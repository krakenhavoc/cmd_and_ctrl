package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knight of New Benalia — Creature — Human Knight {1}{W}, 3/1 (Dominaria, #5530):
//
//	(no rules text)
//
// No rules text: a vanilla creature, registered so it does not wear the
// "unimplemented" badge.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75f28e54-17ce-49f9-85b2-9f9a04e1d86f",
		Name:         "Knight of New Benalia",
		Completeness: effects.CompletenessFull,
	})
}
