package game

import (
	"github.com/google/uuid"
)

// sideboard.go — a deck's sideboard as the player's outside-the-game
// pool (ADR 0191).
//
// ADR 0172 built ZoneOutside for one thing, the revealed companion, and
// said what a casual player's "outside the game" is was a decision it
// did not take. The owner took it on 2026-10-14: a deck's sideboard is
// that pool. This file is the other half of the same idea:
//
//   - THE CONTAINER. Player.Sideboard is a second *Zone of kind
//     ZoneOutside beside Player.Companion, built the same way and for
//     the same reason. No ZoneRef resolves to it and the card index does
//     not cover it, so no move verb, effect or admin tool can reach a
//     card there (CR 400.11c). The ONLY door is an effect that brings a
//     card in: OutsideCandidatesForEffect reads it, RevealOutsideForEffect
//     shows it, BringFromOutsideForEffect moves it into a hand.
//   - SEATING. A card stamped IsSideboard goes there (seatDeckCard).
//   - THE PROMPT. "You may reveal a sorcery card you own from outside the
//     game" is a choose_cards prompt over the owner's sideboard with the
//     zone re-check pointed at it (pickStillInPickZoneLocked). It is not
//     a new kind: every rule about the payload, the bounds and the
//     enumerator's reach is already written once.
//
// Nothing here reads Player.Companion. The companion special action is
// still the only way out of that container, and a sideboard card is
// never offered the {3} action.

// MaxSideboardSize is the most cards a seat may own outside the game
// (ADR 0191 §5): 15, the sideboard a constructed deck is allowed and the
// size #2352 plans for Standard.
const MaxSideboardSize = 15

// ChooseFromOutside is the choose_cards destination of a pick over the
// chooser's own sideboard: the named cards come in from outside the game
// (ADR 0191 §6). The bot prices the pick as a take, the way it prices a
// tutor.
const ChooseFromOutside ChooseDestination = "outside"

// countSideboard is how many cards in a deck are stamped IsSideboard.
func countSideboard(deck []Card) int {
	n := 0
	for _, c := range deck {
		if c.IsSideboard {
			n++
		}
	}
	return n
}

// SideboardCards is the cards this player owns outside the game, in
// sideboard order, or nil when there are none. The slice is the zone's
// own: read it, never write through it. Caller must hold the game's
// lock (read or write).
func (p *Player) SideboardCards() []Card {
	if p == nil || p.Sideboard == nil {
		return nil
	}
	return p.Sideboard.Cards
}

// LookupOutsideCardForEffect finds a card in ANY seat's sideboard. It is
// read-only and exists for the surfaces that SHOW or CHECK a card a
// prompt names (the view, the enumerator, a prompt's set-level
// Validate); nothing that moves, targets or affects a card may call it,
// which is CR 400.11c. LookupCardForEffect deliberately does not see
// these cards.
//
// Caller must hold g.mu (read or write).
func (g *Game) LookupOutsideCardForEffect(cardID uuid.UUID) (Card, bool) {
	for _, p := range g.Seats {
		if p == nil || p.Sideboard == nil {
			continue
		}
		for _, c := range p.Sideboard.Cards {
			if c.InstanceID == cardID {
				return c, true
			}
		}
	}
	return Card{}, false
}

// OutsideCandidatesForEffect is the cards `player` owns outside the game
// that satisfy `match` (nil matches every card), in sideboard order.
// Nil when the pool is closed to them:
//
//   - the seat is gone or eliminated;
//   - another player is making their decisions (ADR 0167 §5, CR 723.4):
//     "information about cards outside the game stays with the
//     controlled player", so a controller is offered nothing, and the
//     effect finds nothing to bring in. Weaker than printed, never
//     stronger.
//
// Caller must hold g.mu (read or write).
func (g *Game) OutsideCandidatesForEffect(player uuid.UUID, match func(Card) bool) []uuid.UUID {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated || p.Sideboard == nil {
		return nil
	}
	if g.ControllerOfPlayerLocked(player) != player {
		return nil
	}
	var out []uuid.UUID
	for _, c := range p.Sideboard.Cards {
		if match == nil || match(c) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// OutsidePickPrompt is the queue-side description of "choose cards you
// own from outside the game".
type OutsidePickPrompt struct {
	// Player owns the cards and answers. Required.
	Player uuid.UUID
	// Source is the card asking.
	Source uuid.UUID
	// Question is the prompt's header, written the way the card is.
	Question string
	// Match narrows the candidates ("a sorcery card"); nil offers every
	// card in the sideboard.
	Match func(Card) bool
	// Min / Max bound the pick, clamped to the candidates. Zero Max
	// means one.
	Min, Max int
	// Validate is a legality check on the picked SET, with
	// ChooseCardsPrompt.Validate's shape and rules (never called for an
	// empty pick, never handed a *Game). "Exactly two cards with
	// different names" is Min 0, Max 2 and a Validate that wants two.
	Validate func(picked []Card) bool
	// Then receives the picks, in the order the chooser submitted them,
	// and runs with g.mu held. It may queue further choices.
	Then func(g *Game, picked []uuid.UUID) error
}

// QueueOutsidePickForEffect asks `p.Player` to choose from the cards
// they own outside the game and reports whether a prompt went up. It
// reports false, and queues nothing, when the pool is closed
// (OutsideCandidatesForEffect) or nothing in it matches: "you may
// reveal a card from outside the game" with no such card is a clause
// that does nothing, and a prompt nobody could answer with anything but
// "nothing" is a click the player had no choice about.
//
// The candidates are the OWNER's own cards, so the chooser may look at
// every one of them: the owner is a knower from seating
// (seatDeckCard). Nobody else is, and the wire withholds the prompt's
// options from every other seat (filterPendingChoices).
//
// Caller must hold g.mu.
func (g *Game) QueueOutsidePickForEffect(p OutsidePickPrompt) bool {
	cands := g.OutsideCandidatesForEffect(p.Player, p.Match)
	if len(cands) == 0 {
		return false
	}
	max := p.Max
	if max <= 0 {
		max = 1
	}
	g.QueueChooseCardsForEffect(ChooseCardsPrompt{
		Chooser:     p.Player,
		FromPlayer:  p.Player,
		Source:      p.Source,
		Question:    p.Question,
		Cards:       cands,
		Min:         p.Min,
		Max:         max,
		Zone:        ZoneOutside,
		Validate:    p.Validate,
		Then:        p.Then,
		Destination: ChooseFromOutside,
	})
	return true
}

// pickStillOutsideLocked is pickStillInPickZoneLocked for a pick over
// the chooser's sideboard: is `id` still in the pick's owner's
// container? Caller must hold g.mu (read or write).
func (g *Game) pickStillOutsideLocked(choice *PendingChoice, id uuid.UUID) bool {
	p := g.playerByIDLocked(choice.FromPlayer)
	if p == nil || p.Sideboard == nil {
		return false
	}
	return p.Sideboard.Contains(id)
}

// RevealOutsideForEffect shows cards `spec.Player` owns outside the game
// to the whole table: every seated player becomes a knower of each, and
// one grouped EventRevealCards run announces them (RevealForEffect's
// contract, for cards that sit outside every zone it can find). Nothing
// moves. A card that is not in the player's sideboard is skipped.
//
// The knowledge is sticky, which is what "entitled to remember" means:
// the revealed card that stays outside is still known to every seat.
//
// Returns the RevealSeq of the run, or 0 when nothing was revealed.
// Caller must hold g.mu.
func (g *Game) RevealOutsideForEffect(spec RevealSpec) uint64 {
	p := g.playerByIDLocked(spec.Player)
	if p == nil || p.Sideboard == nil || len(spec.Cards) == 0 {
		return 0
	}
	seats := make([]uuid.UUID, 0, len(g.Seats))
	for _, s := range g.Seats {
		if s != nil {
			seats = append(seats, s.ID)
		}
	}
	var seq uint64
	for _, id := range spec.Cards {
		found := false
		for i := range p.Sideboard.Cards {
			if p.Sideboard.Cards[i].InstanceID == id {
				p.Sideboard.Cards[i].AddKnowersAll(seats)
				found = true
				break
			}
		}
		if !found {
			continue
		}
		if seq == 0 {
			// RevealForEffect's key: the Seq the first event of the run
			// will be stamped with, so a replay is byte-identical.
			seq = g.eventSeq + 1
		}
		g.EmitEvent(Event{
			Kind:      EventRevealCards,
			Actor:     spec.Player,
			Source:    spec.Source,
			CardID:    id,
			Label:     spec.Reason,
			OldZone:   ZoneOutside,
			RevealSeq: seq,
			RevealTo:  spec.To,
		})
	}
	return seq
}

// BringFromOutsideForEffect puts a card `playerID` owns outside the game
// into their hand: the one thing that ever leaves the container, and
// the "effects that bring them in" CR 400.11c allows. The card keeps the
// knowers it has (a revealed card stays known to the table), and the
// move is announced as a zone change from outside the game.
//
// ErrCardNotFound when the card is not in that player's sideboard.
// Caller must hold g.mu.
func (g *Game) BringFromOutsideForEffect(playerID, cardID uuid.UUID) error {
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Sideboard == nil {
		return ErrCardNotFound
	}
	c, err := p.Sideboard.Remove(cardID)
	if err != nil {
		return ErrCardNotFound
	}
	c.Owner = p.ID
	c.Controller = p.ID
	c.AddKnower(p.ID)
	p.Hand.PushTop(c)
	g.EmitEvent(Event{
		Kind:    EventZoneMove,
		Actor:   p.ID,
		CardID:  c.InstanceID,
		OldZone: ZoneOutside,
		NewZone: ZoneHand,
	})
	return nil
}
