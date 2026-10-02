package game

import "github.com/google/uuid"

// pay_unless_action.go — ADR 0108 §5 decision 4 (owner decision 3,
// #1888): the two payments a pay-unless prompt may ask for besides mana.
//
//	Echo—Discard a card.                  Deepcavern Imp, Rakdos Headliner
//	Echo—Sacrifice two lands.             Skizzik Surger
//	Cumulative upkeep—Sacrifice a land.   Polar Kraken
//	Cumulative upkeep—Discard a card.     Vexing Sphinx
//	Cumulative upkeep—Sacrifice a creature. Phyrexian Soulgorger
//
// "Sacrifice it unless you [do something]" is CR 118.12a: the same thing
// as "you may [do something]; if you don't, sacrifice it", and the action
// is a cost paid as the ability resolves (CR 118.12). So it is the SAME
// prompt as a mana pay-unless — one question, asked of the payer, with
// the payment on the "yes" — and only what pays differs. The payer
// chooses which cards to discard or which permanents to sacrifice, and
// names them on the answer (`card_ids`), as a waterbend payment names
// its taps (#1311). One decision, not two: CR 702.24a says a cumulative
// upkeep's choices are all made and then the whole set of costs is paid
// or none is.
//
// A payment the payer cannot make — fewer cards in hand, fewer matching
// permanents than the count — cannot be chosen (CR 118.3), so "pay" is
// not offered and a "yes" without the picks is a decline, exactly as an
// unfundable mana "yes" is.
//
// PayAction is plain data, not a predicate: it lives on a pending
// prompt, and a func there would be a new closure route the ADR 0041
// ratchet refuses. The sacrifice filter is ADR 0107's PermanentQuery.

// PayActionKind names what a non-mana payment does.
type PayActionKind string

const (
	// PayActionDiscard discards Count cards from the payer's hand
	// (CR 701.9).
	PayActionDiscard PayActionKind = "discard"
	// PayActionSacrifice sacrifices Count permanents the payer controls
	// that match Of (CR 701.21).
	PayActionSacrifice PayActionKind = "sacrifice"
)

// PayAction is a non-mana pay-unless payment: "discard N cards" or
// "sacrifice N permanents of a type".
type PayAction struct {
	Kind PayActionKind
	// Count is how many cards or permanents one payment takes. A
	// cumulative upkeep multiplies the printed one by its age counters.
	Count int
	// Of narrows a sacrifice ("a land", "a creature"). The zero value is
	// any permanent. Unused by a discard.
	Of PermanentQuery
}

// payActionOptionsLocked is every card or permanent `chooser` could
// name toward `a` right now: their hand, in hand order, for a discard;
// the permanents they control matching a.Of, in the policy-neutral
// sacrifice payment order (SacrificePaymentOrderForEffect), for a
// sacrifice. Caller must hold g.mu with fresh layers.
func (g *Game) payActionOptionsLocked(chooser uuid.UUID, a *PayAction) []uuid.UUID {
	if a == nil {
		return nil
	}
	p := g.playerByIDLocked(chooser)
	if p == nil {
		return nil
	}
	var out []uuid.UUID
	switch a.Kind {
	case PayActionDiscard:
		if p.Hand == nil {
			return nil
		}
		for _, c := range p.Hand.Cards {
			out = append(out, c.InstanceID)
		}
	case PayActionSacrifice:
		if g.Battlefield == nil {
			return nil
		}
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			// CR 701.21a: only what you control.
			if c.Controller == chooser && a.Of.matchesLocked(g, c) {
				out = append(out, c.InstanceID)
			}
		}
		out = g.SacrificePaymentOrderForEffect(out, uuid.Nil)
	}
	return out
}

// PayActionOptionsForEffect is payActionOptionsLocked for callers
// already under g.mu: the protocol view (the picker) and the legal-move
// enumerator (the bot's payment).
func (g *Game) PayActionOptionsForEffect(chooser uuid.UUID, a *PayAction) []uuid.UUID {
	return g.payActionOptionsLocked(chooser, a)
}

// validatePayActionLocked checks that `picked` is one payment of `a` by
// `chooser`: exactly Count distinct cards, each one of the live options.
// Caller must hold g.mu.
func (g *Game) validatePayActionLocked(chooser uuid.UUID, a *PayAction, picked []uuid.UUID) error {
	if a == nil || len(picked) != a.Count {
		return ErrInvalidParam
	}
	legal := make(map[uuid.UUID]bool)
	for _, id := range g.payActionOptionsLocked(chooser, a) {
		legal[id] = true
	}
	seen := make(map[uuid.UUID]bool, len(picked))
	for _, id := range picked {
		// One card pays once: naming a land twice is not "two lands".
		if seen[id] || !legal[id] {
			return ErrInvalidParam
		}
		seen[id] = true
	}
	return nil
}

// payActionLocked makes the payment — the discards as one batch, the
// sacrifices as one simultaneous exit — then runs `then`. A discard may
// pause on a CR 903.9 prompt for a commander; `then` runs once the whole
// batch has landed, which is when the cost has been paid.
//
// Caller must hold g.mu in write mode and must have validated `picked`.
func (g *Game) payActionLocked(chooser, source uuid.UUID, a *PayAction, picked []uuid.UUID, then func(g *Game) error) error {
	switch a.Kind {
	case PayActionDiscard:
		return g.discardCardsLocked(chooser, picked, discardOptions{
			cause:  DiscardCauseCost,
			source: source,
			then: func(g *Game, _ []uuid.UUID) error {
				if then == nil {
					return nil
				}
				return then(g)
			},
		})
	case PayActionSacrifice:
		if err := g.payCostSacrificesLocked(picked, nil); err != nil {
			return err
		}
	}
	if then == nil {
		return nil
	}
	return then(g)
}

// PayAction is the non-mana payment of a PendingChoicePayUnless —
// "discard a card", "sacrifice two lands" — or nil when the payment is
// mana. A copy: the prompt's own clause cannot be changed through it.
// Read by the protocol view and the legal-move enumerator;
// ResolvePayUnlessWithCards is the one validator.
func (c *PendingChoice) PayAction() *PayAction {
	if c == nil || c.payUnlessResume == nil || c.payUnlessResume.action == nil {
		return nil
	}
	a := *c.payUnlessResume.action
	return &a
}
