package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Screeching Skaab — Creature — Zombie {1}{U}, 2/1 (Dark Ascension, #5571):
//
//	"When this creature enters, mill two cards."
//
// Wailing Ghoul's enters trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c339982a-45d6-4c1a-aefc-c72427b40744",
		Name:         "Screeching Skaab",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Screeching Skaab — mill two cards", effects.Do(effects.MillCards{N: 2})),
		},
	})
}
