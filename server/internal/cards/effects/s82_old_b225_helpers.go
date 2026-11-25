package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// SteadfastArmasaurDamage is "this creature deals damage equal to its
// toughness to target creature": the source's toughness as it resolves, or
// as it last existed on the battlefield (CR 608.2h), never below zero.
func SteadfastArmasaurDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	info, ok := g.PermanentForEffect(ref)
	if !ok || info.Toughness <= 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: info.Toughness}.Apply(ctx)
	}
	return nil
}

// FirecannonUntapIfDiscarded is Elaborate Firecannon's "you may discard a
// card. If you do, untap this artifact": it untaps only when a card really
// left the hand,.
var FirecannonUntapIfDiscarded = struct{ Then game.DiscardThen }{
	Then: game.RegisterDiscardThen("choose-cards/elaborate-firecannon-untap", func(g *game.Game, r game.DiscardsDone) error {
		if r.Discarded.Count() == 0 {
			return nil
		}
		return g.UntapTargetForEffect(r.Source)
	}),
}
