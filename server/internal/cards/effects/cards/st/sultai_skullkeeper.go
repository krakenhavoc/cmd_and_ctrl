package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sultai Skullkeeper — Creature — Snake Shaman {1}{U}, 2/1 (Fate Reforged,
// #5552):
//
//	"When this creature enters, mill two cards."
//
// Wailing Ghoul's entry trigger: its controller mills two as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "726ba167-9e7b-4d46-a52c-72d5a500aee2",
		Name:         "Sultai Skullkeeper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sultai Skullkeeper — mill two cards", effects.Do(effects.MillCards{N: 2})),
		},
	})
}
