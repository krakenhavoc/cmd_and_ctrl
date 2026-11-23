package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fiery Fall — Instant {5}{R}:
//
//	"Fiery Fall deals 5 damage to target creature.
//	 Basic landcycling {1}{R} ({1}{R}, Discard this card: Search your
//	 library for a basic land card, reveal it, put it into your hand,
//	 then shuffle.)"
//
// One damage clause on a creature target, and the shared basic
// landcycling hand ability (Awaken the Inferno).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f014b1a1-20ef-475c-bf83-f1a82e68ab97",
		Name:         "Fiery Fall",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 5)), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DealDamageToTheTarget(5)(ctx.Game, item)
		},
		Activated: []effects.ActivatedAbility{effects.BasicLandcycling("{1}{R}")},
	})
}
