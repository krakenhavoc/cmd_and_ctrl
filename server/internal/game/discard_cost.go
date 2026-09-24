package game

import "github.com/google/uuid"

// discard_cost.go — #660: discarding as part of an activated
// ability's cost (CR 602.2b), the component `AbilityCost` was
// missing.
//
// Two shapes, because the cards print two (ADR 0062 Decision 2):
//
//   - AbilityCost.DiscardSelf — cycling's "Discard this card"
//     (CR 702.29a). No chooser, no options, no wire payload; the
//     source IS the payment.
//   - AbilityCost.DiscardCards — "Discard a card" (Cryptbreaker),
//     "Discard a creature card" (Fauna Shaman, Survival of the
//     Fittest, Tortured Existence). The activator names the cards at
//     announce.
//
// Both pay through discardCardsLocked with DiscardCauseCost, the ONE
// discard path (#799) with the cause #856 gave it. That is what makes
// a cycled card visible to Marauding Mako, to the CR 614 replacement
// window (#853) and to a future madness — without any of them
// learning that costs exist — and it is what makes CR 601.2h's
// indivisible step a bit on the event (MustSettleNow) rather than a
// policy repeated at each cost site.

// DiscardCost is the general "discard N cards" component of an
// activated ability's cost.
type DiscardCost struct {
	// N is how many cards the clause demands. At least one;
	// effects.Register refuses a zero or negative count, because a
	// component that costs nothing is a card-file mistake rather
	// than a free ability.
	N int

	// Label is the clause as printed, without the verb — "a creature
	// card", "two cards". Shown in the client's picker so the prompt
	// reads like the card rather than like a schema, exactly as
	// SacrificeOther's spec label does.
	Label string

	// Match is the clause's predicate. Nil matches any card in hand
	// ("Discard a card").
	//
	// A func(Card) bool rather than a *TargetSpec: TargetSpec's
	// predicates and specMatchLocked are written against permanents
	// on the battlefield, and paying a cost does not target anyway
	// (CR 601.2h). This is the shape SearchLibrarySpec already uses
	// for the same question one zone over.
	//
	// READ-ONLY, and it runs under g.mu. It sees the card as it
	// sits in the hand: printed characteristics, no layers (a card
	// in a hand has none).
	Match func(Card) bool
}

// Matches reports whether `c` could pay this clause. Nil-safe on the
// predicate: a clause with no predicate takes any card.
func (d *DiscardCost) Matches(c Card) bool {
	if d == nil {
		return false
	}
	if d.Match == nil {
		return true
	}
	return d.Match(c)
}

// DiscardCostOptionsForEffect lists the cards in `playerID`'s hand
// that could pay `cost` right now, in hand order, excluding `sourceID`
// when the source is itself a card in that hand (an ability activated
// from hand cannot pay for itself twice — cycling's "Discard this
// card" is the separate DiscardSelf component).
//
// The view stamps it so the client can skip its picker when the hand
// holds exactly N payable cards, and the legal enumerator solves its
// payment out of it, so the set the bot pays from and the set the
// engine accepts are computed once.
//
// CALLER MUST ALREADY HOLD g.mu (read or write), exactly like
// CounterCostOptionsForEffect next door: both of its callers — the
// view's assembly pass and the legal enumerator — run inside one.
func (g *Game) DiscardCostOptionsForEffect(playerID, sourceID uuid.UUID, cost *DiscardCost) []uuid.UUID {
	if cost == nil || cost.N <= 0 {
		return nil
	}
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Hand == nil || cost == nil {
		return nil
	}
	var out []uuid.UUID
	for i := range p.Hand.Cards {
		c := p.Hand.Cards[i]
		if c.InstanceID == sourceID {
			continue
		}
		if !cost.Matches(c) {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return out
}

// validateDiscardCostLocked resolves the DiscardSelf and DiscardCards
// components into the concrete list of cards to discard, without
// moving anything — the validate-all-then-pay discipline the rest of
// ActivateCatalogAbility uses, so a rejected activation never leaves a
// half-paid cost behind.
//
// What it enforces:
//
//   - DiscardSelf requires the source to be IN A HAND. It is the
//     cycling component, and cycling functions from hand only; a
//     "discard this" cost on a battlefield ability would have nothing
//     to discard. srcZone is the zone the activation was validated
//     against, so this is a free check rather than a second lookup.
//   - Exactly cost.N ids for a DiscardCards clause, each distinct,
//     each in the activator's hand, each matching the clause, and
//     none of them the source of a hand activation.
//   - Ids arriving for a cost with no discard component are rejected
//     rather than ignored, exactly as an unexpected sacrifice_ids is:
//     a client that sends them is confused about which ability it is
//     firing, and dropping the field silently hides that.
//
// Caller must hold g.mu.
func (g *Game) validateDiscardCostLocked(playerID, sourceID uuid.UUID, srcZone ZoneKind, cost AbilityCost, chosen []uuid.UUID) ([]uuid.UUID, error) {
	var out []uuid.UUID
	if cost.DiscardSelf {
		if srcZone != ZoneHand {
			return nil, ErrActivationZoneNotAllowed
		}
		out = append(out, sourceID)
	}
	if cost.DiscardCards == nil {
		if len(chosen) > 0 {
			return nil, ErrInvalidParam
		}
		return out, nil
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return nil, ErrPlayerNotFound
	}
	if len(chosen) != cost.DiscardCards.N {
		return nil, ErrInvalidParam
	}
	seen := make(map[uuid.UUID]bool, len(chosen))
	for _, id := range chosen {
		// One card pays one discard. Naming it twice would let a
		// single card pay "Discard two cards".
		if seen[id] {
			return nil, ErrInvalidParam
		}
		seen[id] = true
		// A hand activation's source is already being paid, or is
		// not payable at all; either way it is not a legal pick here.
		if cost.DiscardSelf && id == sourceID {
			return nil, ErrInvalidParam
		}
		if srcZone == ZoneHand && id == sourceID {
			return nil, ErrInvalidParam
		}
		if !p.Hand.Contains(id) {
			return nil, ErrCardNotFound
		}
		c := g.findHandCardLocked(p, id)
		if c == nil || !cost.DiscardCards.Matches(*c) {
			return nil, ErrInvalidParam
		}
	}
	return append(out, chosen...), nil
}

// findHandCardLocked returns the card with `id` in `p`'s hand, or nil.
// Caller must hold g.mu.
func (g *Game) findHandCardLocked(p *Player, id uuid.UUID) *Card {
	if p == nil || p.Hand == nil {
		return nil
	}
	for i := range p.Hand.Cards {
		if p.Hand.Cards[i].InstanceID == id {
			return &p.Hand.Cards[i]
		}
	}
	return nil
}

// payAbilityDiscardsLocked pays the discard components and, for a
// cycling ability, emits EventCycle (CR 702.29b — activating a
// cycling ability IS cycling the card).
//
// Order matters and is the reverse of what reads naturally: the
// DISCARD happens first, then the cycle event. CR 702.29c says "when
// you cycle this card" triggers see the card in the zone it ended up
// in, which is the graveyard, so emitting the event before the move
// would show every watcher a card still in its owner's hand.
//
// EventCycle carries the cycled card on both Source and CardID —
// Source because the cycling ability is what emitted it, CardID
// because "that card" is what a watcher asks about — matching the
// convention EventDiscardCard and EventCast already use.
//
// Caller must hold g.mu.
func (g *Game) payAbilityDiscardsLocked(playerID, sourceID uuid.UUID, ab ActivatedAbilityShape, cards []uuid.UUID, answers map[uuid.UUID]bool) error {
	if len(cards) > 0 {
		if err := g.discardCardsLocked(playerID, cards, discardOptions{
			cause:            DiscardCauseCost,
			source:           sourceID,
			commanderAnswers: answers,
		}); err != nil {
			return err
		}
	}
	if !ab.Cycling {
		return nil
	}
	g.EmitEvent(Event{
		Kind:   EventCycle,
		Actor:  playerID,
		Source: sourceID,
		CardID: sourceID,
	})
	return nil
}
