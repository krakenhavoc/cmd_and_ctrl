package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Godo's Irregulars — Creature — Human Warrior {R}, 1/1 (Saviors of Kamigawa,
// S82 OLD b583):
//
//	"{R}: This creature deals 1 damage to target creature blocking it."
//
// Goblin Javelineer's target (BlockingSource: a creature blocking this one,
// judged as the ability is put on the stack and again as it resolves) and its
// 1 damage (ArcSpitterPing), on an activated ability for {R}. With no
// blocker there is no legal target and the ability cannot be activated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bcf0ca90-134f-43af-890a-558c5e82974e",
		Name:         "Godo's Irregulars",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature deals 1 damage to target creature blocking it.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.ManaCost("{R}"),
			Targets: effects.BlockingSource(effects.TargetCreature("target creature blocking it")),
			Effect:  effects.ArcSpitterPing,
		}},
	})
}
