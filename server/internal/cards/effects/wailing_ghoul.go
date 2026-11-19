package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wailing Ghoul — Creature — Zombie {1}{B}, 1/3 (Foundations Jumpstart, S82,
// #5929):
//
//	"When this creature enters, mill two cards."
//
// Aftermath Analyst's mill-on-enter trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6ce24851-70d2-416c-8733-78928b4eee74",
		Name:         "Wailing Ghoul",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Wailing Ghoul — mill two cards", Do(MillCards{N: 2})),
		},
	})
}
