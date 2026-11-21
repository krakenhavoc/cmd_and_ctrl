package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodscale Prowler — Creature — Lizard Warrior {2}{R}, 3/1 (Guildpact,
// #5599):
//
//	"Bloodthirst 1 (If an opponent was dealt damage this turn, this
//	 creature enters with a +1/+1 counter on it.)"
//
// Bloodthirst, the entry replacement and nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "655a584c-5280-4317-b707-101b7e31dcb1",
		Name:         "Bloodscale Prowler",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Bloodthirst(1, "Bloodscale Prowler")},
	})
}
