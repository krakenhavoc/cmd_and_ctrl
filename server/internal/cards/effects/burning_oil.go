package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "c8e9bec0-43e5-4e84-9e0e-5f3db362f1e4",
		Name:             "Burning Oil",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		Targets:          TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{W}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DealDamageToTheTarget(3)(ctx.Game, item)
		},
	})
}
