package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82OldB554AttackingKithkinIfPaid is the paid branch of Militia's Pride's
// "you may pay {W}. If you do, create a 1/1 white Kithkin Soldier creature token
// that's tapped and attacking": the token enters tapped attacking whatever the
// triggering creature attacked, which the rebuilt trigger item's event still
// names, and was never declared as an attacker (CR 508.4). An on-disk key:
// never renamed, never reused.
var s82OldB554AttackingKithkinIfPaid = PayUnlessBranches("pay-unless/old-b554-tapped-attacking-kithkin-soldier",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		if ctx.Item == nil || ctx.Item.Trigger == nil {
			return nil
		}
		tmpl := TokenCard("1/1 white Kithkin Soldier")
		tmpl.Tapped = true
		tmpl.AttackingTarget = ctx.Item.Trigger.Event.Target
		return CreateToken{Controller: ctx.Controller(), Template: tmpl, N: 1}.Apply(ctx)
	}, nil)

// S82OldB554MayPayForAnAttackingKithkin is the effect of Militia's Pride's
// attack trigger: its controller may pay {W} as it resolves.
func S82OldB554MayPayForAnAttackingKithkin(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{W}",
		Question: "Militia's Pride — pay {W} to create a tapped and attacking 1/1 Kithkin Soldier?",
		Then:     s82OldB554AttackingKithkinIfPaid,
	}.Apply(NewContext(g, item))
}

// S82OldB554YouClashed is "whenever you clash": game.EventClashed names each
// player who clashed as its Actor (CR 701.30a), whoever's effect began it.
func S82OldB554YouClashed(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventClashed && ev.Actor == source.Controller
}

// s82OldB554FlamekinIfPaid is the paid branch of Rebellion of the Flamekin's
// "you may pay {1}. If you do, create a 3/1 red Elemental Shaman creature
// token. If you won, that token gains haste until end of turn": the clash's
// win event is emitted in the same batch as the trigger's own event, so "you
// won" is read off it. An on-disk key: never renamed, never reused.
var s82OldB554FlamekinIfPaid = PayUnlessBranches("pay-unless/old-b554-flamekin-elemental-shaman",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		if ctx.Item == nil || ctx.Item.Trigger == nil {
			return nil
		}
		ev := ctx.Item.Trigger.Event
		won := false
		for _, e := range ctx.Game.EventsThisTurn() {
			if e.Kind == game.EventClashWon && e.Batch == ev.Batch && e.Actor == ctx.Controller() {
				won = true
				break
			}
		}
		ids, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(TokenCard("3/1 red Elemental Shaman")), 1, game.TokenEntryOptions{})
		if err != nil || !won {
			return err
		}
		for _, id := range ids {
			if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"},
				Label: "Rebellion of the Flamekin — the token gains haste until end of turn"}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}, nil)

// S82OldB554MayPayForAFlamekinShaman is the effect of Rebellion of the
// Flamekin's clash trigger: its controller may pay {1} as it resolves.
func S82OldB554MayPayForAFlamekinShaman(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}",
		Question: "Rebellion of the Flamekin — pay {1} to create a 3/1 red Elemental Shaman?",
		Then:     s82OldB554FlamekinIfPaid,
	}.Apply(NewContext(g, item))
}

// S82OldB554UntapThisAndItGainsShroud is Benthicore's "Untap this creature. It
// gains shroud until end of turn", both halves only while the source is still the
// permanent that was activated (CR 400.7).
func S82OldB554UntapThisAndItGainsShroud(g *game.Game, item *game.StackItem) error {
	if !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	if err := (UntapTarget{Target: item.SourceCardID}).Apply(NewContext(g, item)); err != nil {
		return err
	}
	return ThisGainsKeywordUntilEndOfTurn("shroud", "Benthicore — shroud until end of turn")(g, item)
}

// ANontokenCreatureYouControlAttacked is "whenever a nontoken creature you
// control attacks": an attack declared by the source's controller whose attacker
// is not a token (Mardu Ascendancy, Militia's Pride).
func ANontokenCreatureYouControlAttacked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !AttackDeclaredByYou(ev, source.Controller) {
		return false
	}
	attacker, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !attacker.IsToken()
}
