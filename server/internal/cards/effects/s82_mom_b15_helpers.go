package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b15_helpers.go — the back-face effects of the March of the
// Machine Phyrexian transformers of S82 batch 15 that no other card
// shares.

// blightsowerThallidCondition is "when this creature transforms into
// Blightsower Thallid or dies": one ability, two conditions.
func blightsowerThallidCondition(ev game.Event, source *game.Card, c game.Characteristic, g *game.Game) bool {
	switch ev.Kind {
	case game.EventTransform:
		return ThisTransformedInto(1)(ev, source, c, g)
	case game.EventLTB:
		return ThisDied(ev, source, c, g)
	}
	return false
}

// chromeHostHulkReshape sets the target's base power and toughness to
// 5/5 until end of turn (layer 7b), under any pumps and counters.
func chromeHostHulkReshape(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return untilEndOfTurn(ctx, id, nil, "Chrome Host Hulk — base power and toughness 5/5 until end of turn",
		game.SetBasePTMods(5, 5)...)
}

// maladyInvokerShrink is -0/-X where X is the source's power as the
// ability resolves, last-known if it has left the battlefield (CR 608.2h).
func maladyInvokerShrink(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	info, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	return BoostUntilEOT{Target: id, Toughness: -info.Power, Label: "Malady Invoker — -0/-X until end of turn"}.Apply(ctx)
}

// orderOfTheAlabasterHostShrink gives the blocking creature -1/-1 until
// end of turn. EventBlock names the blocker in CardID; the trigger's
// object is that blocker.
func orderOfTheAlabasterHostShrink(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Object == nil {
		return nil
	}
	o := item.Trigger.Object
	if _, ok := g.PermanentForEffect(game.ObjectRef{ID: o.ID, Epoch: o.Epoch}); !ok {
		return nil
	}
	return BoostUntilEOT{Target: o.ID, Power: -1, Toughness: -1, Label: "Order of the Alabaster Host — -1/-1 until end of turn"}.Apply(NewContext(g, item))
}
