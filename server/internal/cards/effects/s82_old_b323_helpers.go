package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b323_helpers.go — helpers for Magic Origins, batch 323 (S82 older
// sets, #5549).

// OldB323DamageToTheBlocker is "<this> deals N damage to that creature" for a
// per-blocker trigger (CR 509.3d): the creature that blocked is the block
// event's CardID. The source deals it as the object it was as the trigger went
// on the stack, so one that has left the battlefield still deals it (CR 608.2h).
func OldB323DamageToTheBlocker(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		ctx := NewContext(g, item)
		ref, ok := ctx.SourceRef()
		blocker := item.Trigger.Event.CardID
		if !ok || blocker == uuid.Nil {
			return nil
		}
		return DealDamage{SourceObject: &ref, Target: blocker, Amount: n}.Apply(ctx)
	}
}
