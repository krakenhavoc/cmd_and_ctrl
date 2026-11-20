package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cephalid Looter — Creature — Octopus Rogue {2}{U}, 2/1 (Odyssey, #5613):
//
//	"{T}: Target player draws a card, then discards a card."
//
// Built by odyTargetPlayerLoots: any player may be the target, the
// draw comes first, and the discard is chosen from the hand it made.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "412294bf-a25f-47c2-8043-12eb3e142b9e",
		Name:         "Cephalid Looter",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.OdyTargetPlayerLoots("{T}: Target player draws a card, then discards a card.", 1)},
	})
}
