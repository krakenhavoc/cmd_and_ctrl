package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "5ffbbce1-fe38-40dc-a7c3-2055ae86c6c5",
		Name:         "Arwen Undómiel",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventScry, ByYou, arwenLabel, b36CounterOnChosenAnimal),
				TargetCreature("target creature")),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}{G}{U}: Scry 2.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 2},
			Cost:    ManaCost("{4}{G}{U}"),
			Effect:  Do(Scry{N: 2}),
		}},
	})
}
