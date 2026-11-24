package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eyeblight Cullers — Creature — Elf Warrior {4}{B}, 3/3 (Kaldheim
// Commander, #5506):
//
//	"When this creature dies, create three 1/1 green Elf Warrior creature
//	 tokens, then mill three cards."
//
// One death trigger: the tokens first, then the mill, both for the
// controller of the trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b7d69422-b90b-4149-8c6d-eef636d5c6f5",
		Name:         "Eyeblight Cullers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Eyeblight Cullers — create three 1/1 green Elf Warrior tokens, then mill three cards",
				effects.Do(
					effects.CreateToken{Template: effects.TokenCard("1/1 green Elf Warrior"), N: 3},
					effects.MillCards{N: 3},
				)),
		},
	})
}
