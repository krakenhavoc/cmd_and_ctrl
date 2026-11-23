package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dealDamageToEachLegalTarget is "deals N damage to each of [the] targets":
// every target still legal as the spell resolves is dealt the damage at
// once, and one that has become illegal takes nothing (CR 608.2b).
func DealDamageToEachLegalTarget(n int) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		var ids []uuid.UUID
		for _, t := range ctx.LegalTargets() {
			ids = append(ids, t.ID)
		}
		return DealDamageToEachThen(ctx, ids, n, RecipientThen{})
	}
}
