package game

import (
	"fmt"

	"github.com/google/uuid"
)

// reveal_cards_cost.go — "Reveal N <quality> cards from your hand" as
// the cost of an ACTIVATED ability (CR 602.2b, CR 701.20; #2598, ADR
// 0020's 2026-10-08 amendment). Martyr of Bones' "{1}, Reveal X black
// cards from your hand, Sacrifice this creature:" and the rest of the
// Martyr cycle.

// RevealCardsCost is the reveal component of an AbilityCost. It is the
// either/or branch's RevealCost (reveal_cost.go) one owner over: the
// candidate walk and the quality (Subtype / Color) are RevealCost's
// own, so the validator, the view and the bot enumerator read the same
// set the spell path does.
//
// What an ability adds is the COUNT. A cast's reveal names exactly one
// card; the Martyr cycle names X, the number the activator announces
// with the activation (CR 602.2b), which the effect reads back with
// ctx.X() ("You gain three times X life"). So the clause is either a
// fixed N or CountFromX, never both, and CountFromX makes
// AbilityCost.DemandsX true exactly as DiscardCost.CountFromX does —
// there is no {X} symbol, so the mana cost is the printed one whatever
// X is. X may be zero: revealing no cards pays "X cards" at X=0
// (CR 107.3a), a legal and empty payment.
//
// Revealing moves nothing (CR 701.20b): the cards stay in the hand. It
// is not a discard and not a sacrifice, so no discard or dies payoff
// sees it. The whole table sees the cards (RevealForEffect) and they
// stay known while they sit there.
//
// The N and X forms are an ACTIVATED ability's only: a mana ability has
// no stack item to carry an announced X (CR 605.3b). A mana ability
// carries the AnyNumber form alone (Rosheen, Roaring Prophet, #6085), as
// ManaAbilityShape.RevealCards. effects.Register refuses the rest. Build
// it with effects.RevealX / effects.RevealN / effects.RevealAnyNumber,
// never by hand.
type RevealCardsCost struct {
	// RevealCost carries the quality the cards must have. Behold is
	// meaningless here (a battlefield pick is not a hand reveal) and
	// Register refuses it.
	RevealCost
	// N is the fixed count ("Reveal two cards"). At least one unless
	// CountFromX, whose count is the announced X and whose N is 0.
	N int
	// CountFromX makes the count the announced X.
	CountFromX bool
	// Label is the clause as printed, without the verb — "X black
	// cards". Shown in the client's picker.
	Label string

	// SharesColor is "that share a color" (#6208): Illuminated Folio's
	// "{1}, {T}, Reveal two cards from your hand that share a color".
	// Every card named must have a colour in common with every other one
	// — a relation AMONG the picks, which is why it is here and not a
	// quality on RevealCost: a per-card walk has nothing to ask. A card
	// with no colour shares none (CR 105.2c, CR 202.2), so a colourless
	// card can never pay, and neither can a hand of one colourless card
	// and one blue one.
	//
	// It is TapOthersCost.SharesCreatureType's reading one zone over: a
	// clause of N 2, where "every pair shares a colour" and "one colour
	// common to all" are the same thing, judged pairwise. The option walk
	// drops a candidate no other candidate shares a colour with, so a
	// hand that cannot pay offers nothing and the client's picker greys
	// the row; the validator then checks the named set.
	// RevealCardsPaymentForEffect is the enumerator's one payment, since
	// the first N of the walk may be a pair that shares nothing.
	//
	// DATA, not a predicate, so the ADR 0041 closure ratchet gains no
	// route. A fixed count only: no printed card reveals X cards that
	// share a colour, and Register refuses the combination.
	SharesColor bool

	// AnyNumber is "Reveal any number of <quality> cards from your hand"
	// (#6085): Rosheen, Roaring Prophet's "{T}: Reveal any number of cards
	// with {X} in their mana cost in your hand. Add {C}{C} for each card
	// revealed this way." The count is the activator's, from zero to every
	// candidate, and N is 0. On a MANA ability only
	// (ManaAbilityShape.RevealCards): the reveal is part of the ability's
	// effect, and a mana ability resolves as it is activated (CR 605.3b),
	// so the cards are named with the activation, as a colour is named up
	// front (ManaAbilityParams.Colors, #1443). Never beside N, CountFromX
	// or SharesColor; effects.Register refuses the combinations.
	AnyNumber bool `json:",omitempty"`
}

// RevealCardsAnyNumber reports the "any number of" form (#6085).
// Nil-safe.
func RevealCardsAnyNumber(c *RevealCardsCost) bool {
	return c != nil && c.AnyNumber
}

// ErrRevealedColorMismatch is the refusal for a reveal payment whose
// cards do not share a colour (#6208, CR 202.2). It wraps
// ErrInvalidParam, as every other refusal of a reveal payment does, and
// names the rule so the client can say why.
var ErrRevealedColorMismatch = fmt.Errorf("%w: the cards you reveal must share a color", ErrInvalidParam)

// RevealCardsCountFromX reports the "Reveal X cards" form. Nil-safe.
func RevealCardsCountFromX(c *RevealCardsCost) bool {
	return c != nil && c.CountFromX
}

// RevealCardsShareColor reports the "that share a color" form (#6208).
// Nil-safe.
func RevealCardsShareColor(c *RevealCardsCost) bool {
	return c != nil && c.SharesColor
}

// cardsShareColor reports whether two CARDS have a colour in common,
// read after continuous effects (EffectiveColors, CR 202.2). A
// colourless card shares no colour with anything, itself included
// (CR 105.2c) — sharesColor's rule for two permanents, over the card
// values a hand walk hands out.
func cardsShareColor(a, b Card) bool {
	for _, color := range a.EffectiveColors() {
		if b.HasColor(color) {
			return true
		}
	}
	return false
}

// RevealCardsOptionsForEffect lists the cards `playerID` could name to
// pay `rc` when activating `sourceID`: the matching cards in their
// hand, in hand order, other than the source itself (an ability
// activated from a hand cannot pay with its own card; "Reveal this
// card" is a different cost). Nil-safe.
//
// ONE walk, read by the validator, the view and the bot enumerator.
// Caller must hold g.mu (read or write).
func (g *Game) RevealCardsOptionsForEffect(playerID, sourceID uuid.UUID, rc *RevealCardsCost) []uuid.UUID {
	if rc == nil {
		return nil
	}
	ids := g.RevealCostOptionsForEffect(playerID, sourceID, &rc.RevealCost)
	if RevealCardsShareColor(rc) {
		ids = g.revealWithColorPartnerLocked(ids)
	}
	return ids
}

// revealWithColorPartnerLocked keeps the candidates that share a colour
// with at least one other candidate: the necessary condition for a pair
// of them to pay a SharesColor clause (#6208), so a hand whose cards all
// share nothing offers none and the ability is not activatable at all
// (CR 118.3). tapOthersWithSharingPartnerLocked's walk, over cards in a
// hand.
//
// Caller must hold g.mu.
func (g *Game) revealWithColorPartnerLocked(ids []uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range ids {
		a, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		for _, other := range ids {
			if other == id {
				continue
			}
			if b, ok := g.LookupCardForEffect(other); ok && cardsShareColor(a, b) {
				out = append(out, id)
				break
			}
		}
	}
	return out
}

// RevealCardsPaymentForEffect is ONE legal payment of a FIXED reveal
// clause out of the engine's own candidate walk — what the legal-move
// enumerator offers, in place of "the first N of the walk" (#544).
//
// For a plain clause the two are the same thing, and this returns the
// first N. For a SharesColor clause they are not: the first two
// candidates each have a partner somewhere in the hand without being
// each other's. So the payment is the first N candidates that all carry
// one colour, tried in WUBRG order — every card of one colour shares
// that colour with every other, which is exactly what the validator
// checks pairwise.
//
// Nil when the hand cannot pay, so the ability is not offered at all,
// and nil for the X form, whose count is the announcement's.
//
// Caller must hold g.mu (read or write).
func (g *Game) RevealCardsPaymentForEffect(playerID, sourceID uuid.UUID, rc *RevealCardsCost) []uuid.UUID {
	if rc == nil || rc.CountFromX || rc.AnyNumber || rc.N < 1 {
		return nil
	}
	options := g.RevealCardsOptionsForEffect(playerID, sourceID, rc)
	if !RevealCardsShareColor(rc) {
		if len(options) < rc.N {
			return nil
		}
		return options[:rc.N:rc.N]
	}
	for _, color := range AllColors {
		group := make([]uuid.UUID, 0, rc.N)
		for _, id := range options {
			c, ok := g.LookupCardForEffect(id)
			if !ok || !c.HasColor(color) {
				continue
			}
			group = append(group, id)
			if len(group) == rc.N {
				return group
			}
		}
	}
	return nil
}

// validateRevealCardsCostLocked checks the cards named to pay the
// ability's reveal component without revealing anything — the
// validate-all-then-pay discipline. Exactly N (or the announced X)
// distinct cards, each a candidate RevealCardsOptionsForEffect lists.
// Ids for an ability with no reveal component are refused rather than
// ignored, as a stray discard_ids is.
//
// Caller must hold g.mu.
func (g *Game) validateRevealCardsCostLocked(playerID, sourceID uuid.UUID, rc *RevealCardsCost, ids []uuid.UUID, x int) error {
	if rc == nil {
		if len(ids) > 0 {
			return ErrInvalidParam
		}
		return nil
	}
	want := rc.N
	if rc.CountFromX {
		want = x
	}
	// #6085: "any number of" takes every count, zero included; each card
	// is still a distinct candidate, checked below.
	if !rc.AnyNumber && len(ids) != want {
		return ErrInvalidParam
	}
	options := make(map[uuid.UUID]bool)
	for _, id := range g.RevealCardsOptionsForEffect(playerID, sourceID, rc) {
		options[id] = true
	}
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		// One card reveals once: naming it twice would let a single
		// card pay "Reveal two cards".
		if seen[id] || !options[id] {
			return ErrInvalidParam
		}
		seen[id] = true
	}
	// #6208: "that share a color" — every named pair has a colour in
	// common (CR 202.2), which for the printed count of two is the whole
	// clause. A colourless card shares none (CR 105.2c), and a card the
	// engine cannot find shares nothing, so a walk that cannot say what
	// a pick is never admits more than printed.
	if RevealCardsShareColor(rc) {
		for i := range ids {
			for j := i + 1; j < len(ids); j++ {
				a, aok := g.LookupCardForEffect(ids[i])
				b, bok := g.LookupCardForEffect(ids[j])
				if !aok || !bok || !cardsShareColor(a, b) {
					return ErrRevealedColorMismatch
				}
			}
		}
	}
	return nil
}

// payRevealCardsCostLocked reveals the named cards to the table
// (CR 701.20a). Validated at announce; nothing moves. Zero ids (X = 0)
// reveal nothing and emit nothing.
//
// Caller must hold g.mu.
func (g *Game) payRevealCardsCostLocked(playerID, sourceID uuid.UUID, rc *RevealCardsCost, ids []uuid.UUID) {
	if rc == nil || len(ids) == 0 {
		return
	}
	reason := "reveal " + rc.Label + " from hand"
	if c, ok := g.LookupCardForEffect(sourceID); ok {
		reason = c.Name + " — " + reason
	}
	g.RevealForEffect(RevealSpec{Player: playerID, Source: sourceID, Reason: reason, Cards: ids})
}
