package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// DamageXToEachLegalTarget is "<this> deals X damage to each of them":
// every target still legal as the spell resolves (CR 608.2b) is dealt the
// announced X at once. Jaya's Immolating Inferno, Comet Storm. A
// constructor, so a card's OnResolve is a call the XMatters guard reads.
func DamageXToEachLegalTarget() func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		var ids []uuid.UUID
		for _, t := range ctx.LegalTargets() {
			ids = append(ids, t.ID)
		}
		return DealDamageToEachThen(ctx, ids, ctx.X(), RecipientThen{})
	}
}
