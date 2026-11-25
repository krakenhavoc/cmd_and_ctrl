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
	// PayActionExileFromGraveyard exiles Count cards from the payer's
	// OWN graveyard (ADR 0187 §3): Carrion Rats' "any player may exile
	// a card from their graveyard", Carrion Wurm's three. A card pick,
	// so it takes the {apply, card_ids} answer like a discard, and
	// "Pay" is a payment only from a graveyard of at least Count cards
	// (CR 118.3). Graveyards are public, so the options go to every
	// seat.
	PayActionExileFromGraveyard PayActionKind = "exile_graveyard"
	// PayActionEither is one offer with two or more ways to pay (ADR
	// 0187 §5): Worms of the Earth's "any player may sacrifice two
	// lands of their choice or have this enchantment deal 5 damage to
	// that player. If a player does either, …". It is one condition,
	// not two offers, so the payer is asked once and the answer names
	// the branch (`pay_option`, an index into PayAction.Either). It is
	// a wrapper only: the engine pays the chosen branch, as that
	// PayAction, and a branch is never itself an Either.
	PayActionEither PayActionKind = "either"
	// PayActionTap taps Count untapped permanents the payer controls
	// that match Of (CR 701.26a, #1997): Koskun Falls' "sacrifice this
	// enchantment unless you tap an untapped creature you control". A
	// card pick, so it takes the {apply, card_ids} answer like a
	// sacrifice, and "Pay" is a payment only with at least Count
	// untapped matches (CR 118.3). Tapping is not a {T} cost of the
	// creature, so a creature that came under its controller's control
	// this turn may pay (CR 302.6 restricts only its own {T} abilities).
	PayActionTap PayActionKind = "tap"
	// PayActionMana is a mana payment as a BRANCH of an Either (#4566):
	// Titania, Rugged Rumbler's "Ward—Discard a card or pay {2}". A
	// plain mana pay-unless carries no PayAction at all (PayCost is the
	// cost); this kind exists only so an Either can name mana beside a
	// card payment. Mana is the cost string. It names no cards, and it
	// is paid through the prompt's own mana path (pool, then the
	// auto-tapper), so an unfundable "yes" is a decline as always.
	PayActionMana PayActionKind = "mana"
	// PayActionMinusOneCounter puts one -1/-1 counter on each of Count
	// permanents the payer controls that match Of (#5746): Thelon's
	// Chant's "deals 3 damage to that player unless the player puts a
	// -1/-1 counter on a creature they control". A card pick, so it takes
	// the {apply, card_ids} answer like a sacrifice, and "Pay" is a
	// payment only with at least Count matches that can have a counter
	// put on them (CR 118.3; "can't have counters put on it" refuses one,
	// #4692). The counters are a cost's (CR 118.12: the action is a cost
	// paid as the ability resolves), so they go through the CR 614
	// counter window marked as a cost's, as blight's do: Doubling Season,
	// which names "an effect" (CR 614.16), does not double them, and
	// Vizier of Remedies, which names none, still applies. It is not the
	// keyword action blight (CR 701.68a), so a "whenever a player
	// blights" trigger does not see it. A creature the counter kills
	// still paid.
	PayActionMinusOneCounter PayActionKind = "minus_one_counter"
	// PayActionDiscardRandom discards Count cards from the payer's hand
	// at random (#6187, CR 701.9b: the engine picks, not the payer):
	// Balduvian Horde's "sacrifice it unless you discard a card at
	// random". It names nothing, so the answer is yes or no, and "yes"
	// is a payment only with at least Count cards in hand (CR 118.3).
	// The cards are drawn as the payment is made, through the one discard
	// path as a cost's.
	PayActionDiscardRandom PayActionKind = "discard_random"
	// PayActionHandToLibraryTop puts Count cards the payer names from
	// their hand on top of their library (#6187): Tainted Specter's
	// "unless they put a card from their hand on top of their library".
	// A card pick from the hand, like a discard, but not a discard: no
	// EventDiscardCard, nothing for madness to see.
	PayActionHandToLibraryTop PayActionKind = "hand_to_library_top"
	// PayActionSacrificeTotalPower sacrifices any number of permanents
	// the payer controls matching Of whose total power is TotalPower or
	// greater (#6187): Phyrexian Dreadnought's "sacrifice it unless you
	// sacrifice any number of creatures with total power 12 or greater".
	// A card pick of no fixed size: Count is unused, and a payment is
	// one or more distinct options whose powers (as they are now, a
	// negative one counting as itself) add up to TotalPower.
	PayActionSacrificeTotalPower PayActionKind = "sacrifice_total_power"
)

// NamesNothing reports whether a "Pay" of this payment names no cards
// (energy, life, damage, nothing): the answer is yes or no, and
// whether it is a payment is the payer's state, not a pick (CR 118.3).
func (a *PayAction) NamesNothing() bool {
	if a == nil {
		return false
	}
	switch a.Kind {
	case PayActionEnergy, PayActionLife, PayActionDamage, PayActionNothing, PayActionMana, PayActionDiscardRandom:
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
	case PayActionDiscardRandom:
		return discardRandomWords(a.Count)
	}
	return ""
}

// discardRandomWords is "a card at random" or "N cards at random".
func discardRandomWords(n int) string {
	if n == 1 {
		return "a card at random"
	}
	return strconv.Itoa(n) + " cards at random"
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
	case PayActionDiscardRandom:
		// #6187, CR 118.3: N cards in hand to draw from.
		return a.Count > 0 && len(g.RandomDiscardPoolForEffect(chooser, uuid.Nil, nil)) >= a.Count
	case PayActionMana:
		// #4566: what payCostLocked would pay now, asked of a clone so
		// the answer is the payer's own (canPayCostLocked).
		cost, err := ParseCost(a.Mana)
		return err == nil && g.canPayCostLocked(p, cost, uuid.Nil)
	}
	return false
}

// PayActionPayableForEffect is payActionPayableLocked for callers
// already under g.mu: the legal-move enumerator offers "Pay" only when
// the engine would take it as a payment.
func (g *Game) PayActionPayableForEffect(chooser uuid.UUID, a *PayAction) bool {
	return g.payActionPayableLocked(chooser, a)
}

// payActionCanPayLocked reports whether `chooser` can make payment `a`
// right now, whichever kind it is (CR 118.3): a payment that names
// nothing by its state, a card pick when enough cards qualify, and an
// Either when any one branch can be paid (ADR 0187 §5, so Worms of the
// Earth's offer is always payable, its damage branch being so). Caller
// must hold g.mu with fresh layers.
func (g *Game) payActionCanPayLocked(chooser uuid.UUID, a *PayAction) bool {
	if a == nil {
		return false
	}
	switch {
	case a.Kind == PayActionEither:
		for i := range a.Either {
			if g.payActionCanPayLocked(chooser, &a.Either[i]) {
				return true
			}
		}
		return false
	case a.NamesNothing():
		return g.payActionPayableLocked(chooser, a)
	case a.Kind == PayActionSacrificeTotalPower:
		// #6187: every option together reaches the total, so some set
		// of them does (CR 118.3). A negative power only lowers a sum,
		// so it is left out of the best one.
		total := 0
		for _, id := range g.payActionOptionsLocked(chooser, a) {
			if p := paymentPowerLocked(g, id); p > 0 {
				total += p
			}
		}
		return total >= a.TotalPower
	}
	return len(g.payActionOptionsLocked(chooser, a)) >= a.Count
}

// PayActionCanPayForEffect is payActionCanPayLocked for callers already
// under g.mu: the protocol view marks each branch of an Either payable
// or not, and the enumerator offers only the branches the engine would
// take.
func (g *Game) PayActionCanPayForEffect(chooser uuid.UUID, a *PayAction) bool {
	return g.payActionCanPayLocked(chooser, a)
}

// PayAction is a non-mana pay-unless payment: "discard N cards",
// "sacrifice N permanents of a type", "exile N cards from your
// graveyard" or "put a -1/-1 counter on a creature you control".
type PayAction struct {
	Kind PayActionKind
	// Count is how many cards or permanents one payment takes. A
	// cumulative upkeep multiplies the printed one by its age counters.
	Count int
	// Of narrows a sacrifice ("a land", "a creature"). The zero value is
	// any permanent. Unused by a discard.
	Of PermanentQuery
	// Either lists the ways to pay a PayActionEither (ADR 0187 §5):
	// two or more payments, none of them an Either. Plain data on the
	// frame, like the rest of PayAction. Unused by every other kind.
	Either []PayAction
	// Words is the phrase naming a branch of an Either on the wire and
	// in the move list ("sacrifice two lands", "take 5 damage"). Empty
	// falls back to PayActionPhrase. Unused outside an Either.
	Words string
	// Nonland narrows a discard to nonland cards (#4566): Bullseye,
	// Death Dealer's "discard a nonland card". Only a nonland card in the
	// payer's hand is an option, so a hand of lands can't pay (CR
	// 118.3). It narrows a sacrifice the same way: "sacrifice three other
	// nonland permanents" (Scavenger's Talent). Unused by every other kind.
	Nonland bool `json:",omitempty"`
	// CardType narrows a discard to cards of one card type ("creature"):
	// Crypt Lurker's "discard a creature card". Only a card of that type in
	// the payer's hand is an option, so a hand without one can't pay (CR
	// 118.3). A graveyard exile reads it the same way (#6194). Unused by
	// every other kind. Additive (#522).
	CardType string `json:",omitempty"`
	// Except is a permanent a sacrifice may not take: the "other" of
	// "sacrifice three other nonland permanents", the ability's own
	// source (Scavenger's Talent). The zero value excludes nothing.
	// Unused by every other kind.
	Except *uuid.UUID `json:",omitempty"`
	// Mana is a PayActionMana branch's cost ("{2}"), parsed when it is
	// paid (#4566). Unused by every other kind.
	Mana string `json:",omitempty"`
	// TopOnly is "unless you exile the top creature card of your
	// graveyard" (#6194: Barrow Ghoul, Circling Vultures): a graveyard
	// exile whose cards are not the payer's to pick but the topmost Count
	// that match CardType (CR 404.2, the pile's order). The options are
	// exactly those, so a graveyard with fewer can't pay (CR 118.3).
	// Unused by every other kind. Additive (#522).
	TopOnly bool `json:",omitempty"`
	// Forage marks an Either as the forage keyword action (CR 701.61a,
	// #5184; ForagePayAction): paying either branch is a forage, so the
	// answer emits EventForage once the payment is made. Unused by every
	// other payment.
	Forage bool `json:",omitempty"`
	// TotalPower is the power a PayActionSacrificeTotalPower's
	// sacrificed permanents must add up to at least (#6187). Unused by
	// every other kind. Additive (#522).
	TotalPower int `json:",omitempty"`
}

// paymentPowerLocked is the power a battlefield permanent has now, for
// a total-power payment: the layered value with its counters
// (Card.CurrentPower). Zero for a card that is not on the battlefield.
func paymentPowerLocked(g *Game, id uuid.UUID) int {
	c := findBattlefieldCard(g, id)
	if c == nil {
		return 0
	}
	return c.CurrentPower()
}

// PayActionPhrase is the verb phrase for one payment when its card gives
// no words of its own: "take 5 damage", "pay 5 life", "discard 2 cards",
// "sacrifice 2 permanents", "exile 3 cards from your graveyard". The view
// and the move list use it for the branches of an Either.
func PayActionPhrase(a PayAction) string {
	if a.Words != "" {
		return a.Words
	}
	n := strconv.Itoa(a.Count)
	switch a.Kind {
	case PayActionDamage:
		return "take " + n + " damage"
	case PayActionLife:
		return "pay " + n + " life"
	case PayActionEnergy:
		return "pay " + EnergyPaymentWords(a.Count)
	case PayActionDiscard:
		return "discard " + n + " cards"
	case PayActionSacrifice:
		return "sacrifice " + n + " permanents"
	case PayActionTap:
		return "tap " + n + " untapped permanents"
	case PayActionMinusOneCounter:
		if a.Count == 1 {
			return "put a -1/-1 counter on a permanent you control"
		}
		return "put a -1/-1 counter on each of " + n + " permanents you control"
	case PayActionMana:
		return "pay " + a.Mana
	case PayActionExileFromGraveyard:
		return "exile " + n + " cards from your graveyard"
	case PayActionDiscardRandom:
		return "discard " + discardRandomWords(a.Count)
	case PayActionHandToLibraryTop:
		if a.Count == 1 {
			return "put a card from your hand on top of your library"
		}
		return "put " + n + " cards from your hand on top of your library"
	case PayActionSacrificeTotalPower:
		return "sacrifice permanents with total power " + strconv.Itoa(a.TotalPower) + " or greater"
	case PayActionNothing:
		return "say yes"
	}
	return string(a.Kind)
}

// validEither reports whether an Either is well formed: at least two
// branches, none of them an Either (ADR 0187 §5). A branch's own shape
// is the card's business, as it is for a plain payment.
func validEither(a *PayAction) bool {
	if a == nil || a.Kind != PayActionEither || len(a.Either) < 2 {
		return false
	}
	for i := range a.Either {
		if a.Either[i].Kind == PayActionEither || len(a.Either[i].Either) > 0 {
			return false
		}
	}
	return true
}

// EitherBranch is the branch `option` names on an Either, or nil when
// `a` is not an Either or the index is out of range. The branch is a
// copy.
func EitherBranch(a *PayAction, option int) *PayAction {
	if a == nil || a.Kind != PayActionEither || option < 0 || option >= len(a.Either) {
		return nil
	}
	b := a.Either[option]
	return &b
}

// payActionOptionsLocked is every card or permanent `chooser` could
// name toward `a` right now: their hand, in hand order, for a discard;
// their graveyard, in zone order, for a graveyard exile; the permanents
// they control matching a.Of, in the policy-neutral sacrifice payment
// order (SacrificePaymentOrderForEffect), for a sacrifice. Caller must
// hold g.mu with fresh layers.
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
			// #4566: "discard a nonland card" takes no land.
			if a.Nonland && c.IsLand() {
				continue
			}
			if a.CardType != "" && !c.HasCardType(a.CardType) {
				continue
			}
			out = append(out, c.InstanceID)
		}
	case PayActionHandToLibraryTop:
		// #6187: any card in the payer's hand, in hand order, except one
		// whose exit an effect has already paused (#1445).
		if p.Hand == nil {
			return nil
		}
		for _, c := range p.Hand.Cards {
			if !g.zoneChangePausedLocked(c.InstanceID) {
				out = append(out, c.InstanceID)
			}
		}
	case PayActionExileFromGraveyard:
		// ADR 0187 §3: only the payer's own graveyard, in zone order.
		if p.Graveyard == nil {
			return nil
		}
		for _, c := range p.Graveyard.Cards {
			// #6194: "a creature card" narrows a graveyard exile too.
			if a.CardType != "" && !c.HasCardType(a.CardType) {
				continue
			}
			out = append(out, c.InstanceID)
		}
		if a.TopOnly {
			out = topOfPileOrder(out, a.Count)
		}
	case PayActionSacrifice, PayActionSacrificeTotalPower:
		if g.Battlefield == nil {
			return nil
		}
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			// CR 701.21a: only what you control.
			if c.Controller != chooser || !a.Of.matchesLocked(g, c) {
				continue
			}
			if (a.Except != nil && c.InstanceID == *a.Except) || (a.Nonland && c.IsLand()) {
				continue
			}
			out = append(out, c.InstanceID)
		}
		out = g.SacrificePaymentOrderForEffect(out, uuid.Nil)
	case PayActionTap:
		// #1997: only an untapped permanent can be tapped (CR 701.26a),
		// and only one the payer controls. Battlefield order.
		if g.Battlefield == nil {
			return nil
		}
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.Controller == chooser && !c.Tapped && a.Of.matchesLocked(g, c) {
				out = append(out, c.InstanceID)
			}
		}
	case PayActionMinusOneCounter:
		// #5746: only a permanent the payer controls, and only one that
		// can have a counter put on it (CR 118.3, #4692). Battlefield
		// order. Not a target: hexproof and protection don't stop it.
		if g.Battlefield == nil {
			return nil
		}
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.Controller == chooser && a.Of.matchesLocked(g, c) && !g.cantHaveCountersLocked(c) {
				out = append(out, c.InstanceID)
			}
		}
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
	// An Either is never validated whole: the answer names a branch
	// (resolvePayUnless), and the branch is validated as its own payment.
	if a == nil || a.Kind == PayActionEither {
		return ErrInvalidParam
	}
	if a.Kind == PayActionSacrificeTotalPower {
		// #6187: any number, at least one, adding up to the total.
		if len(picked) == 0 {
			return ErrInvalidParam
		}
	} else if len(picked) != a.Count {
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
	if a.Kind == PayActionSacrificeTotalPower {
		total := 0
		for _, id := range picked {
			total += paymentPowerLocked(g, id)
		}
		if total < a.TotalPower {
			return ErrInvalidParam
		}
	}
	return nil
}

// payActionLocked makes the payment — the discards as one batch, the
// sacrifices as one simultaneous exit — then runs `then`. A discard may
// pause on a CR 903.9 prompt for a commander; `then` runs once the whole
// batch has landed, which is when the cost has been paid.
//
// `then` is data (#4361 S8, pay_action_keyed.go): a registered tail, so a
// payment paused on a prompt leaves a table that is a restore point. The
// zero tail is no continuation.
//
// Caller must hold g.mu in write mode and must have validated `picked`.
func (g *Game) payActionLocked(chooser, source uuid.UUID, a *PayAction, picked []uuid.UUID, then KeyedTail) error {
	switch a.Kind {
	case PayActionEither, PayActionMana:
		// Never paid whole: the chosen branch is paid, as its own
		// PayAction (ADR 0187 §5). A mana branch is paid by the prompt's
		// mana path (answerPayUnlessLocked, #4566). Reaching here is a
		// caller bug.
		return ErrInvalidParam
	case PayActionDiscard:
		return g.discardCardsLocked(chooser, picked, discardOptions{
			cause:  DiscardCauseCost,
			source: source,
			keyed:  then,
		})
	case PayActionExileFromGraveyard:
		// ADR 0187 §3: one batch through the exit primitive, a leg at
		// a time from each leg's continuation, so a leg that stops on
		// a CR 903.9 prompt resumes into the rest and `then` runs once
		// after the LAST card has landed. A commander exiled this way
		// lands like any card (ADR 0115) and its owner is offered the
		// command zone by the CR 903.9a state-based action. Not
		// MustSettleNow and not payExileCardsCostLocked: this payment
		// is made while something resolves, not during a cast.
		return g.routeAllKeyedLocked(zoneRoute{
			Dst:    ZoneExile,
			Actor:  chooser,
			Source: source,
			Cause:  MoveCause{Kind: MoveCauseCost, Controller: chooser},
		}, picked, then)
	case PayActionSacrifice, PayActionSacrificeTotalPower:
		if err := g.payCostSacrificesLocked(picked, nil); err != nil {
			return err
		}
	case PayActionDiscardRandom:
		// #6187, CR 701.9b: the engine draws the cards as the payment is
		// made, from the hand as it is now, and discards them as a cost's
		// through the one discard path.
		pool := g.RandomDiscardPoolForEffect(chooser, uuid.Nil, nil)
		if len(pool) < a.Count {
			return ErrInvalidParam
		}
		ids := g.ChooseAtRandomForEffect(RandomDraw{Player: chooser, Source: source}, pool, a.Count)
		return g.discardCardsLocked(chooser, ids, discardOptions{
			cause:  DiscardCauseCost,
			source: source,
			keyed:  then,
		})
	case PayActionHandToLibraryTop:
		// #6187: each named card to the top of its owner's library, in
		// the order named, through the exit primitive one batch at a
		// time as a graveyard exile's are, so a commander's CR 903.9b
		// offer resumes into the rest.
		// The payer still knows the cards they put there, as
		// payPutOnTopCostLocked's do (ADR 0088 Decision 3).
		knowers := make(map[uuid.UUID]map[uuid.UUID]bool, len(picked))
		for _, id := range picked {
			card, _ := g.cardInZoneLocked(g.findCardZoneLocked(id), id)
			knowers[id] = libraryOrderKnowers(card, nil, chooser, 1)
		}
		err := g.routeAllKeyedLocked(zoneRoute{
			Dst:    ZoneLibrary,
			Actor:  chooser,
			Source: source,
			Cause:  MoveCause{Kind: MoveCauseCost, Controller: chooser},
		}, picked, then)
		for _, id := range picked {
			if z := g.findCardZoneLocked(id); z != nil && z.Kind == ZoneLibrary {
				if c := g.findCardByIDLocked(id); c != nil {
					c.KnownBy = knowers[id]
				}
			}
		}
		return err
	case PayActionTap:
		// #1997: the taps are the payment, each announced as a tap
		// (EventTapCard) through the path a tap cost pays with.
		g.payTapPermanentsCostLocked(chooser, picked)
	case PayActionMinusOneCounter:
		// #5746: one -1/-1 counter on each named permanent, as a cost's
		// (CR 118.12) with the payer as its placer, settled at once as
		// every cost-shaped placement is. The cost is paid whatever the
		// window makes of the counter.
		for _, id := range picked {
			if findBattlefieldCard(g, id) == nil {
				continue
			}
			if _, err := g.payCostCounterLocked(chooser, id, CounterMinusOne, 1); err != nil {
				return err
			}
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
		return g.dealDamageToPlayerTailLocked(source, nil, chooser, a.Count, DamageMarks{}, 0, damageThen{keyed: then})
	}
	return g.runKeyedTailLocked(then, nil, 0)
}

// PayAction is the non-mana payment of a PendingChoicePayUnless —
// "discard a card", "sacrifice two lands" — or nil when the payment is
// mana. A copy: the prompt's own clause cannot be changed through it.
// Read by the protocol view and the legal-move enumerator;
// ResolvePayUnlessWithCards is the one validator.
func (c *PendingChoice) PayAction() *PayAction {
	if c == nil {
		return nil
	}
	// The prompt (#4264) carries its payment as data.
	return copyPayAction(c.PayActionClause)
}
