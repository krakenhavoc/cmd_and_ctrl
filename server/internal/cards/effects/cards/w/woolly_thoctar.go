package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Woolly Thoctar — {R}{G}{W} Creature — Beast 5/4 (Shards of Alara, S82 OLD b513):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see grizzly_bears.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "220bfa3b-0e17-4350-a967-cfcf05b754ef",
		Name:         "Woolly Thoctar",
		Completeness: effects.CompletenessFull,
	})
}
