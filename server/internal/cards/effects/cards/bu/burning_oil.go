package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burning Oil — Instant {1}{R} (Dark Ascension, #5571):
//
//	"Burning Oil deals 3 damage to target attacking or blocking
//	 creature.
//	 Flashback {3}{W} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Impeccable Timing's body. Flashback is the shared alternative cost (CR
// 702.34a): castable from the graveyard for {3}{W}, exiled as it leaves
// the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "c8e9bec0-43e5-4e84-9e0e-5f3db362f1e4",
		Name:             "Burning Oil",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		Targets:          effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{3}{W}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DealDamageToTheTarget(3)(ctx.Game, item)
		},
	})
}
