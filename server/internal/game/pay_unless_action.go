package game

import (
	"strconv"

	"github.com/google/uuid"
)

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
	// PayActionEnergy is in energy_payment.go (ADR 0129 §3).

	// PayActionLife pays Count life (ADR 0150 §3): "unless any player
	// pays 5 life" (Aether Rift, Dash Hopes), Temporal Extortion's half
	// a life total. Payable only from a life total of at least Count
	// (CR 119.4) and never while the payer's life can't change (CR
	// 119.8), through CanPayLifeLocked, the gate a life cost uses.
	PayActionLife PayActionKind = "life"
	// PayActionDamage is Count damage dealt by the prompt's source to
	// the payer (ADR 0150 §3): "any player may have Browbeat deal 5
	// damage to them". Being dealt damage isn't a resource a player
	// can lack, so "Pay" is always a payment, even when the damage is
	// then prevented or replaced: it goes through the ordinary damage
	// path, and the "yes" branch runs once it has been dealt.
	PayActionDamage PayActionKind = "damage"
	// PayActionNothing is a "yes" that costs nothing (ADR 0150 §3):
	// Distant Memories' "any opponent may have you put that card into
	// your hand". Always payable.
	PayActionNothing PayActionKind = "nothing"
)

// NamesNothing reports whether a "Pay" of this payment names no cards
// (energy, life, damage, nothing): the answer is yes or no, and
// whether it is a payment is the payer's state, not a pick (CR 118.3).
func (a *PayAction) NamesNothing() bool {
	if a == nil {
		return false
	}
	switch a.Kind {
	case PayActionEnergy, PayActionLife, PayActionDamage, PayActionNothing:
		return true
	}
	return false
}

// HalfLifeRoundedUp is "half their life, rounded up" (CR 107.1a), for
// a payment worked out per payer (Temporal Extortion). A total of 0 or
// less is 0.
func HalfLifeRoundedUp(life int) int {
	if life <= 0 {
		return 0
	}
	return (life + 1) / 2
}

// PayActionWords are the words a payment shows on the prompt and in
// the log when the card gives none of its own: "5 life", "6 damage".
// Empty for a card pick or for nothing.
func PayActionWords(a PayAction) string {
	switch a.Kind {
	case PayActionLife:
		return strconv.Itoa(a.Count) + " life"
	case PayActionDamage:
		return strconv.Itoa(a.Count) + " damage"
	case PayActionEnergy:
		return EnergyPaymentWords(a.Count)
	}
	return ""
}

// payActionPayableLocked reports whether `chooser` can make a payment
// that names nothing right now (CR 118.3). Caller must hold g.mu.
func (g *Game) payActionPayableLocked(chooser uuid.UUID, a *PayAction) bool {
	p := g.playerByIDLocked(chooser)
	if p == nil || p.Eliminated || a == nil {
		return false
	}
	switch a.Kind {
	case PayActionEnergy:
		return g.energyPaymentPayableLocked(chooser, a)
	case PayActionLife:
		// CR 119.4 and CR 119.8, through the gate every life cost uses.
		return g.CanPayLifeLocked(p, a.Count)
	case PayActionDamage, PayActionNothing:
		return true
	}
	return false
}

// PayActionPayableForEffect is payActionPayableLocked for callers
// already under g.mu: the legal-move enumerator offers "Pay" only when
// the engine would take it as a payment.
func (g *Game) PayActionPayableForEffect(chooser uuid.UUID, a *PayAction) bool {
	return g.payActionPayableLocked(chooser, a)
}

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
	case PayActionEnergy:
		// ADR 0129 §3 (energy_payment.go): the counters come off the
		// payer through the one path that pays energy.
		if err := g.payEnergyLocked(chooser, a.Count, source); err != nil {
			return err
		}
	case PayActionLife:
		// ADR 0150 §3: the life comes off through the one path that
		// pays life as a cost (CR 119.4, with 119.8 as its backstop).
		if err := g.payLifeAsCostLocked(source, chooser, a.Count); err != nil {
			return err
		}
	case PayActionDamage:
		// ADR 0150 §3: the source deals the damage through the ordinary
		// path, so prevention and replacement apply. The payment has
		// been made once the damage event has settled, however much
		// of it landed.
		return g.dealDamageToPlayerLocked(source, nil, chooser, a.Count, DamageMarks{}, 0, func(g *Game, _ int) error {
			if then == nil {
				return nil
			}
			return then(g)
		})
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
