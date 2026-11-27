package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b634_helpers.go — Apocalypse, batch 634 (#5614).

// S82OldB634RevealFourTakeSubtype is "reveal the top four cards of your
// library. Put all <subtype> cards revealed this way into your hand and the
// rest on the bottom of your library in any order": every card of the
// creature type the card names is taken, and the controller orders the rest.
// The type is the card's own word (a text-changing effect changes it).
func S82OldB634RevealFourTakeSubtype(name, subtype string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		return S82OldB074RevealTopTakeAllToHandRestBottom(g, item.Controller, ctx.Source(), 4,
			name+" — reveal the top four cards, "+subtype+" cards to your hand and the rest on the bottom of your library",
			HasSubtype(ctx.TextSubtype(subtype)))
	}
}

// S82OldB634Sanctuary is the Apocalypse Sanctuaries' upkeep trigger: "At the
// beginning of your upkeep, if you control a <colour A> or <colour B>
// permanent, <effect>. If you control a <colour A> permanent and a <colour B>
// permanent, <the stronger effect> instead." The intervening if (CR 603.4) is
// checked as it triggers and again on resolution; "instead" is read as it
// resolves (CR 608.2c). `apply` receives whether both colours are controlled.
// A non-nil `targets` is chosen as the trigger goes on the stack.
func S82OldB634Sanctuary(a, b, label string, targets *game.TargetSpec, apply func(ctx *Context, both bool) error) game.TriggeredAbility {
	has := func(ctx *Context, letter string) bool {
		return B18ControlsPermanentOfColor(ctx.Game, ctx.Controller(), ctx.TextColor(letter))
	}
	either := CondBoard{
		Name: "you control a " + a + " or " + b + " permanent",
		Test: func(ctx *Context) bool { return has(ctx, a) || has(ctx, b) },
	}
	t := On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return ev.Actor == source.Controller &&
			(B18ControlsPermanentOfColor(g, source.Controller, source.TextColor(a)) ||
				B18ControlsPermanentOfColor(g, source.Controller, source.TextColor(b)))
	}, label, func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		return apply(ctx, has(ctx, a) && has(ctx, b))
	})
	if targets != nil {
		t = Targeting(t, targets)
	}
	return InterveningIf(either, t)
}

// S82OldB634TargetPlayerLosesLife is Necra Sanctuary's body: the target
// player loses 3 life when both colours are controlled and 1 otherwise. A
// target that is no longer a legal player does nothing (CR 608.2b).
func S82OldB634TargetPlayerLosesLife(ctx *Context, both bool) error {
	p, ok := FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	n := 1
	if both {
		n = 3
	}
	return GainLife{Player: p, Amount: -n}.Apply(ctx)
}

// S82OldB634YouGainLife is Dega Sanctuary's body: you gain 4 life when both
// colours are controlled and 2 otherwise.
func S82OldB634YouGainLife(ctx *Context, both bool) error {
	n := 2
	if both {
		n = 4
	}
	return GainLife{Player: ctx.Controller(), Amount: n}.Apply(ctx)
}

// S82OldB634GerrardUpkeep is "you gain 1 life for each card in target
// opponent's hand", the hand counted as the ability resolves.
func S82OldB634GerrardUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	opp, ok := FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	return GainLife{Player: ctx.Controller(), Amount: CardsInHand(g, opp)}.Apply(ctx)
}
