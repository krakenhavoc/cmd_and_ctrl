package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// meld.go — the printed meld ability (CR 701.42, 712.4; ADR 0145). The
// engine half is game/meld.go.
//
// Every printed meld ability is one sentence with the same tail:
//
//	"If you both own and control <this> and a <type> named <partner>,
//	 exile them, then meld them into <combined back face>."
//
// The head is printed text and varies — Urza, Lord Protector's is an
// activated ability, Graf Rats' a beginning-of-combat trigger whose
// condition is also an intervening "if" (CR 603.4), Mishra, Claimed by
// Gix wants both attacking — so a card writes its own condition with
// YouOwnAndControlThisAndANamed and its effect with MeldWith.

// MeldWith is "exile them, then meld them" said by this permanent's
// ability, with the condition in front of it checked as the ability
// resolves: the source must still be on the battlefield, owned and
// controlled by the ability's controller, and so must a permanent of
// PartnerType named Partner. With either missing nothing happens — the
// sentence is conditional, and a meld ability whose partner died in
// response does not exile its own source.
//
// The partner is found by name, which is what the card says: a Clone
// copying The Mightstone and Weakstone is "an artifact named The
// Mightstone and Weakstone" and is exiled, and then the two cannot be
// melded and stay in exile (CR 701.42c). When more than one permanent
// qualifies, a real meld card is preferred, then the first on the
// battlefield.
type MeldWith struct {
	Partner     string
	PartnerType string
}

func (m MeldWith) Apply(ctx *Context) error {
	src := ctx.Source()
	partner := meldPartnerFor(ctx.Game, src, ctx.Controller(), m.Partner, m.PartnerType)
	if partner == uuid.Nil {
		return nil
	}
	return ctx.Game.MeldForEffect(src, partner, ctx.Controller())
}

// YouOwnAndControlThisAndANamed is the printed condition of a meld
// trigger, as an intervening "if" (CR 603.4): the source and a
// permanent of partnerType named partner are both on the battlefield,
// both owned and controlled by the source's controller. Graf Rats,
// Titania, Voice of Gaea and Gisela, the Broken Blade check it twice,
// as the trigger fires and as it resolves; MeldWith rechecks the same
// thing at resolution, so a card wires this one into the trigger's
// condition only.
func YouOwnAndControlThisAndANamed(partner, partnerType string) func(g *game.Game, source *game.Card) bool {
	return func(g *game.Game, source *game.Card) bool {
		if source == nil {
			return false
		}
		return meldPartnerFor(g, source.InstanceID, source.Controller, partner, partnerType) != uuid.Nil
	}
}

// meldPartnerFor finds the permanent a meld ability of `src` would meld
// with, or uuid.Nil when the condition fails: `src` is on the
// battlefield owned and controlled by `controller`, and so is a
// permanent of `partnerType` named `partner`.
func meldPartnerFor(g *game.Game, src, controller uuid.UUID, partner, partnerType string) uuid.UUID {
	if g.Battlefield == nil {
		return uuid.Nil
	}
	ownsAndControls := func(c *game.Card) bool {
		return c.Owner == controller && c.Controller == controller && !c.FaceDown
	}
	found := false
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID == src {
			found = ownsAndControls(c)
			break
		}
	}
	if !found {
		return uuid.Nil
	}
	best := uuid.Nil
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID == src || !ownsAndControls(c) || !c.HasCardType(partnerType) || !game.CardNameMatches(*c, partner) {
			continue
		}
		if c.IsMeldCard() {
			return c.InstanceID
		}
		if best == uuid.Nil {
			best = c.InstanceID
		}
	}
	return best
}

// whenThisAttacksTokensTappedAndAttacking is the Hanweir pair's attack
// trigger, on Hanweir Garrison and on the permanent it melds into:
// "Whenever this creature attacks, create N <template> tokens that are
// tapped and attacking." The tokens attack the player or permanent the
// source attacked, read off the attack EVENT at trigger time and
// carried on the item (CR 506.4's defending player is fixed at
// declaration), and were never declared as attackers, so "whenever a
// creature attacks" does not fire for them (CR 508.4) — the source's
// own trigger included. A defender that has left the game by
// resolution gets no tokens rather than tokens attacking nobody.
func whenThisAttacksTokensTappedAndAttacking(label, template string, n int) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventAttack},
		AppliesTo: ThisAttacked,
		Key:       label,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Player = b17DefendingPlayer(g, ev)
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			defender := item.Params.Player
			if defender == uuid.Nil {
				return nil
			}
			tmpl := TokenCard(template)
			tmpl.Tapped = true
			tmpl.AttackingTarget = defender
			return CreateToken{Controller: item.Controller, Template: tmpl, N: n}.Apply(NewContext(g, item))
		},
	}
}
