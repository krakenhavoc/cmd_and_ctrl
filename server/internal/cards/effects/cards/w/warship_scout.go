package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warship Scout — Creature — Human Scout {R}, 2/1 (Avatar: The Last
// Airbender Eternal, #5075):
//
// No rules text: a vanilla creature, registered so it does not wear the
// "unimplemented" badge.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f07e6ba7-42e8-4b4d-9831-00a856059060",
		Name:         "Warship Scout",
		Completeness: effects.CompletenessFull,
	})
}
