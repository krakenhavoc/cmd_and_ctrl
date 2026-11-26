package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enraged Revolutionary — Creature — Human Warrior {2}{R}, 2/1 (Conspiracy,
// S82 OLD batch 363, #5557):
//
//	"Dethrone (Whenever this creature attacks the player with the most life
//	 or tied for most life, put a +1/+1 counter on it.)"
//
// The shared GapB201Dethrone trigger (CR 702.105a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f4b8e856-17a9-44db-bce6-312a6bd2af32",
		Name:         "Enraged Revolutionary",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.GapB201Dethrone("Enraged Revolutionary — dethrone: put a +1/+1 counter on it"),
		},
	})
}
