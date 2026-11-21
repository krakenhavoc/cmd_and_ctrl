package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Defense Force Aggressor — Creature — Klingon Warrior {R}, 2/1 (Star Trek
// Commander, #4974):
//
//	(no rules text)
//
// A vanilla body is still registered so it resolves, enters and attacks
// without the manual sandbox path; its cost, types and P/T ride the card
// data from Scryfall.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f1a39bf-7597-4655-b7af-4b1332edf529",
		Name:         "Defense Force Aggressor",
		Completeness: effects.CompletenessFull,
	})
}
