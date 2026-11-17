package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b17_helpers.go — shared bodies for March of the Machine's Sieges, batch 17
// (#5477): the two "power and toughness are each equal to the number of ..."
// back faces, and Grandmother Ravi Sengir's "any other target" damage.

// s82MomB17CountPT is a layer 7a characteristic-defining ability: "<this>'s
// power and toughness are each equal to <count>" (CR 604.3, CR 613.4a).
func s82MomB17CountPT(label string, count func(g *game.Game, source *game.Card) int) game.StaticAbility {
	return game.StaticAbility{
		Label:     label,
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7A_CDA,
		AppliesTo: selfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := count(g, source)
			c.Power, c.Toughness = n, n
		},
	}
}

// s82MomB17AnyOtherTarget is TargetAny with the label "any other target"; the
// "other" is enforced at resolution (see tarkirAnyOtherTarget).
func s82MomB17DamageAnyOtherTarget(n int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if len(item.Targets) == 0 {
			return nil
		}
		t := item.Targets[0]
		if t.ID == item.SourceCardID || !ctx.IsTargetLegal(t) {
			return nil
		}
		return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
	}
}

// bounceUpToOneTarget is "return up to one target creature to its owner's
// hand": the one target, if still legal, goes back. Shared by Bottomless Pool
// and Invasion of Xerex.
func bounceUpToOneTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		return BounceToHand{Target: t.ID}.Apply(ctx)
	}
	return nil
}
