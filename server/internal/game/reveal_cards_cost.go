package game

import "github.com/google/uuid"

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
// An ACTIVATED ability only: a mana ability has no stack item to carry
// an announced X (CR 605.3b), and no printed mana ability reveals.
// effects.Register refuses it on one. Build it with effects.RevealX /
// effects.RevealN, never by hand.
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
}

// RevealCardsCountFromX reports the "Reveal X cards" form. Nil-safe.
func RevealCardsCountFromX(c *RevealCardsCost) bool {
	return c != nil && c.CountFromX
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
	return g.RevealCostOptionsForEffect(playerID, sourceID, &rc.RevealCost)
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
	if len(ids) != want {
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
