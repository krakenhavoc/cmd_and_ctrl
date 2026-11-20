package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sultai Skullkeeper — Creature — Snake Shaman {1}{U}, 2/1 (Fate Reforged,
// #5552):
//
//	"When this creature enters, mill two cards."
//
// Wailing Ghoul's entry trigger: its controller mills two as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "726ba167-9e7b-4d46-a52c-72d5a500aee2",
		Name:         "Sultai Skullkeeper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sultai Skullkeeper — mill two cards", Do(MillCards{N: 2})),
		},
	})
}
