package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fire at Will — Instant {R/W}{R/W}{R/W} (Eventide, S82 OLD batch 520, #5590):
//
//	"Fire at Will deals 3 damage divided as you choose among one, two, or three
//	 target attacking or blocking creatures."
//
// Mogg Mob's divided damage with an attacking-or-blocking creature clause: the
// division is announced with the targets as the spell is cast (CR 601.2d), each
// target at least 1 and the shares summing to 3. A target that is no longer
// attacking or blocking, or has left, takes nothing and its share is not
// redistributed (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ca8409ab-d438-441a-861f-c41e307d5621",
		Name:         "Fire at Will",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.TargetCreature("target attacking or blocking creatures", effects.AttackingOrBlocking()).
			WithCount(1, 3).Dividing(effects.Divide(3)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DealDividedDamage(ctx)
		},
	})
}
