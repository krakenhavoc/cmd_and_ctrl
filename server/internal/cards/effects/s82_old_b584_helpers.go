package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b584_helpers.go — S82 OLD batch 584 (#5601): Saviors of Kamigawa
// channel cards. Append-only.

// OldB584SourceDealsXToTarget is "<this> deals X damage to target <creature>":
// the first target still legal at resolution takes X (CR 608.2b), dealt by the
// ability's source as last known (CR 608.2h).
func OldB584SourceDealsXToTarget() func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: ctx.X()}.Apply(ctx)
			}
		}
		return nil
	}
}

// OldB584SourceDealsXToEach is "It deals X damage to each <creature matching>"
// from a hand-zone channel cost, X announced with the cost (CR 602.2b).
func OldB584SourceDealsXToEach(match CardPredicate) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		return DamageEachMatching(ctx, And(Creature(), match), ctx.X())
	}
}
