package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b466_helpers.go — shared pieces of S82 OLD batch 466 (#5576).

// OldB466SourceDealsOneDamage is "This creature deals 1 damage to any target"
// as the body of a granted "Sacrifice this creature:" ability (Mortarpod,
// Mask of Immolation's shape): the damage comes from the sacrificed creature
// as it last existed (CR 608.2h), so its keywords ride on the damage.
func OldB466SourceDealsOneDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	legal := ctx.LegalTargets()
	if !ok || len(legal) == 0 {
		return nil
	}
	return DealDamage{SourceObject: &ref, Target: legal[0].ID, Amount: 1}.Apply(ctx)
}
