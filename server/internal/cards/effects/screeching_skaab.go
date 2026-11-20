package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Screeching Skaab — Creature — Zombie {1}{U}, 2/1 (Dark Ascension, #5571):
//
//	"When this creature enters, mill two cards."
//
// Wailing Ghoul's enters trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c339982a-45d6-4c1a-aefc-c72427b40744",
		Name:         "Screeching Skaab",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Screeching Skaab — mill two cards", Do(MillCards{N: 2})),
		},
	})
}
