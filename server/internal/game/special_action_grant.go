package game

import (
	"slices"

	"github.com/google/uuid"
)

// special_action_grant.go — #1391, ADR 0062 amendment 2026-09-24
// ("a special action granted to another card").
//
// Every special action before this file was the card's OWN: printed
// (foretell, suspend, plot) or derived from its face-down state
// (turn_face_up). Fblthp, Lost on the Range is the first permanent
// that gives one to a DIFFERENT card:
//
//	"The top card of your library has plot. The plot cost is equal to
//	 its mana cost.
//	 You may plot nonland cards from the top of your library."
//
// Those two sentences are two separate things, and the grant keeps
// them apart:
//
//   - the ZONE. Plot works from the hand (CR 702.170a). The second
//     sentence adds the top of your library for any plot the card has,
//     including a plot it prints itself. SpecialActionGrant.Zone and
//     .Nonland.
//   - the OFFER. The first sentence gives the top card plot at its own
//     mana cost. On its own it would do nothing, because plot works
//     only from the hand. It matters because the second sentence lets
//     it be used from the library. SpecialActionGrant.CostIsManaCost.
//
// So a Djinn of Fool's Fall (plot {3}{U}, mana cost {4}{U}) on top of
// the library offers TWO plots, and the player chooses between them
// with SpecialActionParams.Cost. A card that prints no plot offers the
// one at its mana cost.
//
// Nothing else changes. The timing table is the kind's (plot is
// sorcery-speed wherever the card is), the performer is the kind's
// (plotLocked routes the card to exile from wherever it is), and the
// cost is paid through the same CR 601.2f pass. The grant is only
// WHERE and AT WHAT PRICE.

// SpecialActionGrant is a special action a permanent gives to OTHER
// cards: its controller may take `Kind` on a card in `Zone`. Pure data,
// like SpecialAction, so it rides CardDef without a closure.
//
// Build one with an effects constructor (effects.PlotFromTopOfLibrary)
// rather than by hand. effects.Register refuses at boot a grant whose
// kind or zone the engine cannot carry out.
type SpecialActionGrant struct {
	// Kind is which special action is granted. Only SpecialActionPlot
	// is built: its performer routes the card to exile from any zone.
	// Foretell's and suspend's performers take the card from a hand.
	Kind SpecialActionKind

	// Zone is the zone the grant opens. ZoneLibrary means the TOP card
	// of the controller's own library ("your library") and no other
	// card in it, because a library is only reachable at its top
	// (CR 401.5). It is the only zone built.
	Zone ZoneKind

	// Nonland limits the grant to nonland cards ("You may plot nonland
	// cards from the top of your library").
	Nonland bool

	// CostIsManaCost also GIVES the card the kind, at a cost equal to
	// its own mana cost ("The top card of your library has plot. The
	// plot cost is equal to its mana cost."). A card with no mana cost
	// is offered nothing: a cost based on the mana cost of an object
	// with no mana cost is unpayable (CR 118.6). {X} counts as 0.
	// Nothing lets the player choose it, and the later plotted cast is
	// free, so paying more would buy nothing.
	CostIsManaCost bool
}

// CatalogSpecialActionGrants is the catalog hook for
// CardDef.SpecialActionGrants, wired by the effects package at init.
// Nil, or a nil return, means the permanent grants nothing.
var CatalogSpecialActionGrants func(oracleID string) []SpecialActionGrant

// SpecialActionGrantBuilt reports whether the engine can carry out a
// grant of this kind in this zone. effects.Register reads it and
// panics at boot on anything else, the same treatment
// SpecialActionKindBuilt gives a declared kind.
func SpecialActionGrantBuilt(kind SpecialActionKind, zone ZoneKind) bool {
	return kind == SpecialActionPlot && zone == ZoneLibrary
}

// specialActionGrantsLocked is every grant made by a permanent `actor`
// controls, in battlefield order.
//
// CatalogAbilityKey rather than CatalogKey: the grant is a static
// ability, so a Fblthp that has lost its abilities grants nothing.
//
// Caller must hold g.mu (read or write).
func (g *Game) specialActionGrantsLocked(actor uuid.UUID) []SpecialActionGrant {
	if g.Battlefield == nil || CatalogSpecialActionGrants == nil || actor == uuid.Nil {
		return nil
	}
	var out []SpecialActionGrant
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != actor {
			continue
		}
		if key := CatalogAbilityKey(*c); key != "" {
			out = append(out, CatalogSpecialActionGrants(key)...)
		}
	}
	return out
}

// specialActionGrantZonesLocked is the zones, beyond the kind's own,
// where `actor` may take `kind` right now. Each zone appears once,
// however many permanents grant it.
//
// Caller must hold g.mu (read or write).
func (g *Game) specialActionGrantZonesLocked(actor uuid.UUID, kind SpecialActionKind) []ZoneKind {
	var out []ZoneKind
	for _, gr := range g.specialActionGrantsLocked(actor) {
		if gr.Kind == kind && gr.Zone != specialActionZone(kind) && !slices.Contains(out, gr.Zone) {
			out = append(out, gr.Zone)
		}
	}
	return out
}

// grantedZoneCard is the one card in `zone` that a grant can reach for
// player p. For the library that is its top card (the last element).
func grantedZoneCard(p *Player, zone ZoneKind) (Card, bool) {
	if p == nil || zone != ZoneLibrary || p.Library == nil || len(p.Library.Cards) == 0 {
		return Card{}, false
	}
	return p.Library.Cards[len(p.Library.Cards)-1], true
}

// SpecialActionsOfferedLocked is every special action `actor` is
// offered on `card` where it sits, in `zone`. It is the game-aware form
// of SpecialActionsOfferedByCard, and the engine, the legal-move
// enumerator and the wire projection all ask through it (#1391):
//
//   - in the kind's OWN zone, the card's own offers, declared or
//     derived. This is the list SpecialActionsOfferedByCard returns,
//     limited to the kinds whose zone this is;
//   - in a zone that a permanent `actor` controls GRANTS a kind in,
//     that kind's offers on the card, if the card is the one the grant
//     reaches (the top of `actor`'s own library) and passes the grant's
//     filter. These are the card's own offers of the kind, moved to the
//     zone, plus the grant's mana-cost offer when it has one and that
//     price is not already offered. Each carries SpecialAction.Zone.
//
// Caller must hold g.mu (read or write).
func (g *Game) SpecialActionsOfferedLocked(actor uuid.UUID, card Card, zone ZoneKind) []SpecialAction {
	var out []SpecialAction
	own := SpecialActionsOfferedByCard(card)
	for _, sa := range own {
		if specialActionZone(sa.Kind) == zone {
			out = append(out, sa)
		}
	}
	top, ok := grantedZoneCard(g.playerByIDLocked(actor), zone)
	if !ok || top.InstanceID != card.InstanceID {
		return out
	}
	// Which kinds the grants open here for this card, and which of
	// those also come with an offer at the card's mana cost. Two
	// Fblthps make one widening and one mana-cost offer, not two.
	var kinds []SpecialActionKind
	manaOffer := map[SpecialActionKind]bool{}
	for _, gr := range g.specialActionGrantsLocked(actor) {
		if gr.Zone != zone || gr.Zone == specialActionZone(gr.Kind) {
			continue
		}
		if gr.Nonland && card.IsLand() {
			continue
		}
		if !slices.Contains(kinds, gr.Kind) {
			kinds = append(kinds, gr.Kind)
		}
		manaOffer[gr.Kind] = manaOffer[gr.Kind] || gr.CostIsManaCost
	}
	for _, kind := range kinds {
		first := len(out)
		for _, sa := range own {
			if sa.Kind == kind {
				sa.Zone = zone
				out = append(out, sa)
			}
		}
		if !manaOffer[kind] {
			continue
		}
		sa, ok := manaCostOffer(kind, card, zone)
		if !ok {
			continue
		}
		if !slices.ContainsFunc(out[first:], func(have SpecialAction) bool { return have.Cost == sa.Cost }) {
			out = append(out, sa)
		}
	}
	return out
}

// manaCostOffer is the offer a CostIsManaCost grant gives `card`: the
// kind at a cost equal to the card's mana cost, with {X} as 0.
//
// No offer when the card has no mana cost (CR 118.6: unpayable) or one
// the engine cannot parse. A mana cost of {0} is a real, free offer. It
// is written "{0}" rather than "" so that SpecialActionParams.Cost can
// still pick it out.
func manaCostOffer(kind SpecialActionKind, card Card, zone ZoneKind) (SpecialAction, bool) {
	parsed, ok := libraryManaCost(card)
	if !ok {
		return SpecialAction{}, false
	}
	parsed.XSlots = 0
	cost := parsed.String()
	if cost == "" {
		cost = "{0}"
	}
	return SpecialAction{
		Kind:  kind,
		Cost:  cost,
		Label: specialActionVerb(kind) + " " + cost,
		Zone:  zone,
	}, true
}

// libraryManaCost is the mana cost of a card outside the stack, or
// false when it has none.
//
// A SPLIT card's mana cost there is both halves combined (CR 709.4b),
// but the deck import materialises face 0 (SetFace), so Card.ManaCost
// holds only the left half. Reading it would plot Fire // Ice for {1}{R}
// instead of {1}{R}{1}{U}, which is cheaper than printed, so the halves
// are summed here. Every other layout's mana cost outside the stack is
// its front face's (CR 712.8a, CR 715.4), which is what Card.ManaCost
// already holds.
func libraryManaCost(card Card) (ParsedCost, bool) {
	costs := []string{card.ManaCost}
	if card.Layout == LayoutSplit && len(card.Faces) > 1 {
		costs = costs[:0]
		for _, f := range card.Faces {
			costs = append(costs, f.ManaCost)
		}
	}
	var out ParsedCost
	printed := false
	for _, s := range costs {
		if s == "" {
			continue
		}
		p, err := ParseCost(s)
		if err != nil {
			return ParsedCost{}, false
		}
		printed = true
		out.Generic += p.Generic
		out.Required = append(out.Required, p.Required...)
		out.XSlots += p.XSlots
		out.HasPhyrexian = out.HasPhyrexian || p.HasPhyrexian
		out.HasSnow = out.HasSnow || p.HasSnow
	}
	return out, printed
}

// specialActionVerb is the word a granted offer's label starts with,
// the same word the keyword's own constructor uses ("Plot {3}{U}").
func specialActionVerb(kind SpecialActionKind) string {
	switch kind {
	case SpecialActionPlot:
		return "Plot"
	case SpecialActionForetell:
		return "Foretell"
	case SpecialActionSuspend:
		return "Suspend"
	case SpecialActionTurnFaceUp:
		return "Turn face up"
	}
	return string(kind)
}

// specialActionOfferLocked is the offer of `kind` on `card` in `zone`
// that `actor` means, or nil when there is none. `cost` picks between
// two offers of the same kind (SpecialActionParams.Cost). Empty takes
// the first.
//
// Caller must hold g.mu (read or write).
func (g *Game) specialActionOfferLocked(actor uuid.UUID, card Card, zone ZoneKind, kind SpecialActionKind, cost string) *SpecialAction {
	for _, sa := range g.SpecialActionsOfferedLocked(actor, card, zone) {
		if sa.Kind != kind {
			continue
		}
		if cost != "" && sa.Cost != cost {
			continue
		}
		out := sa
		return &out
	}
	return nil
}
