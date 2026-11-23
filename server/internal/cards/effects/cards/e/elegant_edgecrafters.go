package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elegant Edgecrafters — Creature — Elf Artificer {4}{G}{G}, 3/4 (Kaladesh, #5539):
//
//	"This creature can't be blocked by creatures with power 2 or less.
//	 Fabricate 2 (When this creature enters, put two +1/+1 counters on
//	 it or create two 1/1 colorless Servo artifact creature tokens.)"
//
// The block restriction is Enraged Ceratok's, judged as blockers are
// declared (CR 509.1b) on each would-be blocker's current power.
// Fabricate 2 is the shared keyword constructor (CR 702.123a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eeb3d282-97df-4df2-8bea-9432e21b6d62",
		Name:         "Elegant Edgecrafters",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerLE(2), "creatures with power 2 or less"),
		},
		Triggered: []game.TriggeredAbility{effects.Fabricate("Elegant Edgecrafters", 2)},
	})
}
