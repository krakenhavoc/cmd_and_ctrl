package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Centaur Healer — Creature — Centaur Cleric {1}{G}{W}, 3/3 (Return to
// Ravnica, #5566):
//
//	"When this creature enters, you gain 3 life."
//
// Koala-Sheep's enters trigger: it uses the stack and its controller gains
// the life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff435f84-73ab-4a16-8af6-6576af4e4edb",
		Name:         "Centaur Healer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Centaur Healer — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
}
