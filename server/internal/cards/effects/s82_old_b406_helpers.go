package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB406ThreatenFirstTarget is "gain control of target creature an
// opponent controls until end of turn. Untap that creature. It gains haste
// until end of turn" as an ability body (Captivating Crew's activated version
// and Smelt-Ward Gatekeepers' enters trigger): the first target, if it is
// still legal as the ability resolves (CR 608.2b), changes control for the
// turn, untaps and gains haste.
func S82OldB406ThreatenFirstTarget(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		target, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		if err := (GainControl{Target: target, Duration: DurationUntilEndOfTurn(ctx), Label: label + " — gain control until end of turn"}).Apply(ctx); err != nil {
			return err
		}
		if err := (UntapTarget{Target: target}).Apply(ctx); err != nil {
			return err
		}
		return GrantKeywordUntilEOT{Target: target, Keywords: []string{"haste"}, Label: label + " — haste until end of turn"}.Apply(ctx)
	}
}

// S82OldB406BloodScrivenerDraw is "If you would draw a card while you have no
// cards in hand, instead you draw two cards and you lose 1 life": Phial of
// Galadriel's draw replacement (the draw grows by one, gated on the drawer's
// empty hand as the draw would happen) whose Replace also costs the controller
// 1 life. The life is lost as the replacement applies, so it is lost once per
// replaced draw, and not at all for a draw made with cards in hand.
func S82OldB406BloodScrivenerDraw(label string) game.ReplacementEffect {
	r := DrawBecomes{Count: func(n int) int { return n + 1 }, Scope: DrawsByController, Label: label}.Build()
	applies, replace := r.AppliesTo, r.Replace
	r.AppliesTo = func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
		if !applies(ev, g, src) {
			return false
		}
		p := g.PlayerByIDForEffect(ev.DrawPlayer)
		return p != nil && (p.Hand == nil || p.Hand.Size() == 0)
	}
	r.Replace = func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
		if err := replace(ev, g, src); err != nil {
			return err
		}
		return g.ChangePlayerLifeForEffect(src.InstanceID, ev.DrawPlayer, -1)
	}
	return r
}

// S82OldB406RevealRandomCardFromHand is "<player> reveals a card at random from
// their hand" (Friendly Fire's pick): one card chosen on the keyed random
// stream, revealed to the table, and returned so a caller can read it. The
// zero id means the hand was empty and nothing was revealed.
func S82OldB406RevealRandomCardFromHand(ctx *Context, player uuid.UUID, reason string) uuid.UUID {
	p := ctx.Game.PlayerByIDForEffect(player)
	if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
		return uuid.Nil
	}
	ids := make([]uuid.UUID, 0, p.Hand.Size())
	for _, c := range p.Hand.Cards {
		ids = append(ids, c.InstanceID)
	}
	pick := ctx.Game.ChooseAtRandomForEffect(RandomDraw(ctx), ids, 1)
	if len(pick) == 0 {
		return uuid.Nil
	}
	ctx.Game.RevealForEffect(game.RevealSpec{Player: player, Source: ctx.Source(), Cards: pick, Reason: reason})
	return pick[0]
}

// S82OldB406InstantOrSorceryYouControlDealtDamage is the condition of "whenever
// an instant or sorcery spell you control deals damage" (Blaze Commando): the
// damage source is an instant or sorcery spell controlled by the ability's
// controller, read while it is mid-resolution, to any recipient.
func S82OldB406InstantOrSorceryYouControlDealtDamage(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Source == uuid.Nil || source == nil {
		return false
	}
	spell, item, ok := g.ResolvingSpellForEffect(ev.Source)
	if !ok || item.Kind != game.StackItemSpell || item.Controller != source.Controller {
		return false
	}
	return spell.IsInstant() || spell.IsSorcery()
}
