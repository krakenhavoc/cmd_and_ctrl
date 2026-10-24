package game

import (
	"github.com/google/uuid"
)

// companion.go — companion (CR 702.139, 116.2g, ADR 0172).
//
// A deck may reveal one companion: a card from outside the game whose
// condition the deck fulfils (CR 103.2b). The deck checker judges the
// condition (internal/deck); the engine carries the card:
//
//   - SEATING. A card stamped IsCompanion goes into Player.Companion
//     rather than the library (seatDeckCard). That container is outside
//     the game (ZoneOutside, CR 400.11): no ZoneRef resolves to it and
//     the card index does not cover it, so nothing that moves, finds or
//     targets a card by reference can reach the companion (CR 400.11c).
//   - THE REVEAL. When the game starts, after the starting player is
//     known and before the libraries are shuffled (CR 103.2b before
//     103.3), each revealed companion is announced in turn order from
//     the starting player (revealCompanionsLocked). Declaring it in the
//     deck list is the choice to reveal it (owner decision 7). It is
//     public from then on.
//   - THE SPECIAL ACTION. Once per game, any time its owner has priority
//     with an empty stack in a main phase of their turn, they may pay
//     {3} and put it into their hand (CR 116.2g, 702.139a). A sixth
//     kind of the one special-action verb (special_action.go), added in
//     its three tables and nowhere else. The container emptying is the
//     "once": the card never comes back (CR 702.139c).

// SpecialActionCompanion is CR 116.2g / CR 702.139a: pay {3} and put
// your companion from outside the game into your hand. Its card is in
// Player.Companion (ZoneOutside), and its offer is DERIVED from the card
// being there (CompanionOffer), never declared by a card file.
const SpecialActionCompanion SpecialActionKind = "companion"

// CompanionActionCost is the companion special action's cost (CR
// 116.2g). Generic mana, so mana that "can't be spent to pay generic
// mana costs" (Jegantha's own) cannot pay it.
const CompanionActionCost = "{3}"

// CompanionOffer is the companion special action a card offers, or nil.
// Every revealed companion offers it, whatever it is: the offer comes
// from the rules, not from the card (the card's own abilities do not
// function outside the game). It is asked only in ZoneOutside
// (specialActionZone), so the IsCompanion stamp a companion keeps after
// it reaches a hand offers nothing there.
func CompanionOffer(c Card) *SpecialAction {
	if !c.IsCompanion {
		return nil
	}
	return &SpecialAction{
		Kind:  SpecialActionCompanion,
		Cost:  CompanionActionCost,
		Label: "Put into your hand (companion)",
	}
}

// countCompanions is how many cards in a deck are stamped IsCompanion.
func countCompanions(deck []Card) int {
	n := 0
	for _, c := range deck {
		if c.IsCompanion {
			n++
		}
	}
	return n
}

// seatDeckCard puts one card of a deck being seated where it starts the
// game: a commander in the command zone, a companion or a sideboard
// card outside the game, anything else in the library.
func seatDeckCard(p *Player, c Card) {
	switch {
	case c.IsCommander:
		p.Command.PushTop(c)
	case c.IsCompanion:
		if p.Companion == nil {
			p.Companion = newZone(ZoneOutside, p.ID)
		}
		p.Companion.PushTop(c)
	case c.IsSideboard:
		// ADR 0191 §3: the deck's sideboard, outside the game. Its
		// owner may look at it at any time, so they are a knower of
		// each card from the start; nobody else is.
		if p.Sideboard == nil {
			p.Sideboard = newZone(ZoneOutside, p.ID)
		}
		p.HasSideboard = true
		c.AddKnower(p.ID)
		p.Sideboard.PushTop(c)
	case c.StickerSheet != nil:
		// ADR 0208 §2, CR 123.2: a sticker sheet is not a card. Its
		// pseudo-card leaves the sheet in the revealed pool and goes
		// nowhere itself.
		seatStickerSheet(p, *c.StickerSheet)
	case c.AstrotoriumBack:
		// ADR 0207 §2, CR 717.2: an Attraction card never begins the
		// game in a deck. It starts face down in its owner's Attraction
		// deck in the command zone.
		seatAttractionCard(p, c)
	default:
		p.Library.PushTop(c)
	}
}

// CompanionCard is the companion waiting outside the game for this
// player, or false when there is none (never revealed, or already
// taken). Caller must hold the game's lock (read or write).
func (p *Player) CompanionCard() (Card, bool) {
	if p == nil || p.Companion == nil || len(p.Companion.Cards) == 0 {
		return Card{}, false
	}
	return p.Companion.Cards[0], true
}

// revealCompanionsLocked announces every seat's companion as the game
// begins (CR 103.2b): one EventCompanionRevealed per seat that has one,
// in turn order from the starting player, as the Scryfall ruling orders
// it ("the starting player does so first, and players proceed in turn
// order"). Every seated player becomes a knower of each companion, so
// it is public from here on.
//
// Caller must hold g.mu (write).
func (g *Game) revealCompanionsLocked(startingSeat int) {
	n := len(g.Seats)
	if n == 0 {
		return
	}
	all := make([]uuid.UUID, 0, n)
	for _, p := range g.Seats {
		if p != nil {
			all = append(all, p.ID)
		}
	}
	for i := 0; i < n; i++ {
		p := g.Seats[(startingSeat+i)%n]
		if p == nil || p.Eliminated || p.Companion == nil {
			continue
		}
		for ci := range p.Companion.Cards {
			c := &p.Companion.Cards[ci]
			c.AddKnowersAll(all)
			g.EmitEvent(Event{
				Kind:   EventCompanionRevealed,
				Actor:  p.ID,
				CardID: c.InstanceID,
				Label:  c.Name,
			})
		}
	}
}

// companionLocked is the companion performer (CR 116.2g): the {3} is
// paid, so the card leaves Player.Companion for its owner's hand. It
// stays in the game until the game ends (CR 702.139c) and never returns
// to the container, which is what makes the action once per game.
//
// Every seated player stays a knower: the card was revealed to the
// table, so opponents see it in the hand until it changes zones.
//
// Caller must hold g.mu (write).
func (g *Game) companionLocked(p *Player, cardID uuid.UUID, _ SpecialAction) error {
	if p.Companion == nil {
		return ErrCardNotFound
	}
	c, err := p.Companion.Remove(cardID)
	if err != nil {
		return err
	}
	c.Owner = p.ID
	c.Controller = p.ID
	c.AddKnower(p.ID)
	p.Hand.PushTop(c)
	return nil
}
