package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Red Dragon — Creature — Dragon {4}{R}{R}, 4/4 (Adventures in the Forgotten Realms):
//
//	Flying
//	 Fire Breath — When this creature enters, it deals 4 damage to each opponent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ff6e4346-463c-445d-8f72-11cb35dd99ee",
		Name:            "Red Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{DamageEachOpponent: 4},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Red Dragon — Fire Breath: 4 damage to each opponent", effects.Do(effects.DamageEachOpponent{N: 4})),
		},
	})
}
