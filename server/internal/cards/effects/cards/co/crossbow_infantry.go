package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crossbow Infantry — Creature — Human Soldier Archer {1}{W}, 1/1 (Eighth
// Edition, S82 OLD batch 611, #5607):
//
//	"{T}: This creature deals 1 damage to target attacking or blocking
//	 creature."
//
// Quilled Sliver's granted ping as the creature's own ability: the target must
// be attacking or blocking as it is chosen and again as the ability resolves
// (CR 608.2b). A tap ability, so summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd99d791-3305-4195-ad12-96d958a28764",
		Name:         "Crossbow Infantry",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to target attacking or blocking creature.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
			Purpose: game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 1)), Answers: game.AnswerRemove},
			Effect:  effects.DealDamageToTheTarget(1),
		}},
	})
}
