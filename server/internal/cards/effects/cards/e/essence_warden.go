package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Essence Warden — Creature — Elf Shaman {G}, 1/1 (EDHREC rank 1180):
//
//	"Whenever another creature enters, you gain 1 life."
//
// Soul Warden in green, word for word. ANY creature, under anyone's
// control; "another" excludes its own entry only; one trigger per
// creature, so a five-token batch is five triggers, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6ca2a89e-7032-4864-b4e9-66f3178f90ab",
		Name:         "Essence Warden",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AnotherCreatureEntered, "Essence Warden — you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
