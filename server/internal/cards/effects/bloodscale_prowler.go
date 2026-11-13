package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "655a584c-5280-4317-b707-101b7e31dcb1",
		Name:         "Bloodscale Prowler",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{Bloodthirst(1, "Bloodscale Prowler")},
	})
}
