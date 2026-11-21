package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Expanding Ooze — Creature — Ooze {1}{B}{G}, 3/3 (Modern Horizons 3):
//
//	"{B}{G}: Adapt 1.
//	 Whenever this creature attacks, put a +1/+1 counter on target modified
//	 creature you control. (Equipment, Auras you control, and counters are
//	 modifications.)"
func init() {
	const label = "Expanding Ooze — put a +1/+1 counter on target modified creature you control"
	attack := Targeting(WheneverThisAttacks(label, b36CounterOnChosenAnimal),
		TargetCreature("target modified creature you control", YouControl(), modifiedCreature))
	Register(Spec{
		OracleID:     "4feeb493-8db5-41d3-bd27-486052e61f28",
		Name:         "Expanding Ooze",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{B}{G}: Adapt 1.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{B}{G}"),
			Effect:  func(g *game.Game, item *game.StackItem) error { return adaptN(g, item, 1) },
		}},
		Triggered: []game.TriggeredAbility{attack},
	})
}
