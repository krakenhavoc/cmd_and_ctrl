package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nightscape Master — Creature — Zombie Wizard {2}{B}{B}, 2/2 (Invasion, S82 #5616):
//
//	"{U}{U}, {T}: Return target creature to its owner's hand.
//	 {R}{R}, {T}: This creature deals 2 damage to target creature."
//
// Two tap abilities, so only one can be used a turn: a bounce of any creature
// and Prodigal Sorcerer's ping for 2 aimed at a creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e5f9bd6b-0aa6-4f61-b101-ec194af6d632",
		Name:         "Nightscape Master",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{U}{U}, {T}: Return target creature to its owner's hand.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    effects.Plus(effects.ManaCost("{U}{U}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.BounceChosenTarget,
			},
			{
				Label:   "{R}{R}, {T}: This creature deals 2 damage to target creature.",
				Cost:    effects.Plus(effects.ManaCost("{R}{R}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(2),
			},
		},
	})
}
