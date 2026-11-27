package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fyndhorn Elder — Creature — Elf Druid {2}{G} 1/1 (Eighth Edition, S82 older sets batch 620, #5607):
//
//	"{T}: Add {G}{G}."
//
// Glorious Sunrise's {T}: Add {G}{G}{G} granted row, printed on the creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "507bdb1a-90b4-4fd3-a1a3-e8be316e97f3",
		Name:         "Fyndhorn Elder",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}{G}",
			Label:    "{T}: Add {G}{G}",
			Answers:  game.AnswerValue,
		}},
	})
}
