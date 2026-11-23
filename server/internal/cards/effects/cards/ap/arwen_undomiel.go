package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const arwenLabel = "Arwen Undómiel — put a +1/+1 counter on target creature"

// Arwen Undómiel — Legendary Creature — Elf Noble {G}{U}, 2/2 (The Lord of
// the Rings: Tales of Middle-earth Commander):
//
//	"Whenever you scry, put a +1/+1 counter on target creature.
//	 {4}{G}{U}: Scry 2."
//
// Galadriel of Lothlórien's scry watch with a targeted counter, and Aether
// Theorist's activated scry. Her own scry triggers the first ability, so
// the activation puts a counter on a creature of her controller's choice
// (any creature, not only theirs). A target that left in response gets
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ffbbce1-fe38-40dc-a7c3-2055ae86c6c5",
		Name:         "Arwen Undómiel",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventScry, effects.ByYou, arwenLabel, effects.B36CounterOnChosenAnimal),
				effects.TargetCreature("target creature")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{G}{U}: Scry 2.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 2},
			Cost:    effects.ManaCost("{4}{G}{U}"),
			Effect:  effects.Do(effects.Scry{N: 2}),
		}},
	})
}
