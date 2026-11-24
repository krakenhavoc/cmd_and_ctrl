package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ezzaroot Channeler — Creature — Treefolk Druid {5}{G}, 4/6 (Commander
// 2021):
//
//	"Reach
//	 Creature spells you cast cost {X} less to cast, where X is the amount of
//	 life you gained this turn.
//	 {T}: You gain 2 life."
//
// A cost reduction counted per cast (CR 601.2f) from the turn's life-gained
// tally: it spends against generic mana only and stops at zero, as every
// reduction does, and counts the life the caster has gained so far this
// turn, so the Channeler's own {T} ability pays for the next creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3e72ec1f-9941-42e6-b832-58b79b44066b",
		Name:            "Ezzaroot Channeler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		CostModifiers: []game.CostModifier{
			effects.CostsLessEach(func(q game.CostQuery) int { return effects.B15LifeGainedThisTurn(q.Game, q.Controller) },
				"Creature spells you cast cost {X} less to cast, where X is the amount of life you gained this turn.",
				effects.YourSpell(), effects.CreatureSpell()),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: You gain 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect:  effects.Do(effects.GainLife{Amount: 2}),
		}},
	})
}
