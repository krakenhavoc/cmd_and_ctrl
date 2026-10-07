package game

import (
	"fmt"

	"github.com/google/uuid"
)

// opening_hand.go — CR 103.6: actions a card in an opening hand lets
// its owner take before the game begins (ADR 0133). The one shape the
// catalog has is "you may begin the game with this on the battlefield"
// (CR 103.6a): the Leylines, Leyline Axe, and Gemstone Caverns with its
// two riders. Chancellor-style "you may reveal this card from your
// opening hand" is a different shape (a reveal that arms a delayed
// trigger) and is not here.
//
// # When
//
// The mulligan window closes when every live seat has kept
// (settleMulliganLocked). That is the moment CR 103.6 describes — the
// starting hands are final — so it is where each seat is asked, in turn
// order from the starting player (CR 103.6), once per qualifying card
// in hand order.
//
// The questions are queued as ordinary Confirm prompts together, not one
// after another, and the first step's entry hooks run as they always
// did. That is a deliberate narrowing of "before the game begins" and
// ADR 0133 says what it costs: the table is parked on the prompts the
// way it is on any open choice (ChoiceBlocksTable), so no one acts
// before they are answered, and nothing in the catalog can see the
// difference — the untap and upkeep entry of turn one have no
// permanents to act on yet, and a permanent that arrives afterwards
// enters untapped like any other. What the choice does buy is that no
// answer can be left dangling by a seat that departs mid-prompt: the
// continuation is the prompt's own branch, and a dropped prompt simply
// leaves the card in hand.
//
// # What it never does
//
// It never runs for a card whose spec declares no OpeningHandAction, so
// a game with none in any opening hand queues nothing and behaves
// exactly as before. It is never offered twice: the window closes once.

// OpeningHandAction is a CR 103.6 action a card offers from its owner's
// opening hand (Spec.OpeningHand): "you may begin the game with this on
// the battlefield". The zero value is the plain Leyline clause.
//
// Plain data on purpose: the three riders printed on real cards are
// fields rather than closures, so the catalog entry is restorable and
// the ADR 0041 closure ratchet gains no route. A card that prints a
// fourth rider adds a fourth field here.
type OpeningHandAction struct {
	// NotStartingPlayer is Gemstone Caverns' "and you're not the
	// starting player" (Game.StartingSeat): the action is not offered
	// to the seat that takes the first turn.
	NotStartingPlayer bool

	// EntersWithCounters are counters the card has as the game begins —
	// Gemstone Caverns' "with a luck counter on it". Put on after the
	// card has entered rather than as it enters: nothing can respond to
	// an opening-hand action, so the two are indistinguishable, and a
	// counter on a permanent that is already there cannot be doubled by
	// a replacement that is not there yet.
	EntersWithCounters map[string]int

	// ExileFromHand is "if you do, exile a card from your hand": how
	// many cards its owner then exiles from what is left in their
	// hand, chosen by them. Zero is every Leyline. A hand with fewer
	// cards exiles what it has (CR 701.17b-style: do as much as you
	// can).
	ExileFromHand int
}

// CatalogOpeningHand returns a card's opening-hand action, or nil when
// it offers none. Installed at init to read CardDef.OpeningHand; a
// game-package test can replace it without importing the catalog.
var CatalogOpeningHand func(key string) *OpeningHandAction

// offerOpeningHandActionsLocked asks every seat, in turn order from the
// starting seat (CR 103.6), whether to take each opening-hand action
// its hand offers. Called once, from the close of the mulligan window,
// after every hand is final.
//
// Caller must hold g.mu.
func (g *Game) offerOpeningHandActionsLocked() {
	if CatalogOpeningHand == nil {
		return
	}
	n := len(g.Seats)
	for k := 0; k < n; k++ {
		seat := g.Seats[(g.StartingSeat+k)%n]
		if seat == nil || seat.Eliminated || seat.Hand == nil {
			continue
		}
		// Read the hand before queueing anything: the prompts hold the
		// card's identity, not a position in a slice that a later answer
		// reshapes.
		type offer struct {
			id     uuid.UUID
			name   string
			action *OpeningHandAction
		}
		var offers []offer
		for i := range seat.Hand.Cards {
			c := &seat.Hand.Cards[i]
			act := CatalogOpeningHand(catalogKeyOf(c))
			if act == nil || (act.NotStartingPlayer && seat.Seat == g.StartingSeat) {
				continue
			}
			offers = append(offers, offer{id: c.InstanceID, name: c.Name, action: act})
		}
		for _, o := range offers {
			g.queueOpeningHandOfferLocked(seat.ID, o.id, o.name, o.action)
		}
	}
}

// openingHandQuestion is the prompt's header, the card's own words.
func openingHandQuestion(name string, act *OpeningHandAction) string {
	q := fmt.Sprintf("Begin the game with %s on the battlefield?", name)
	if act.ExileFromHand > 0 {
		q += " If you do, you exile a card from your hand."
	}
	return q
}

// queueOpeningHandOfferLocked queues one card's Confirm prompt. The
// answer "yes" runs takeOpeningHandActionLocked; "no" leaves the card in
// hand, to be drawn and cast like any other.
//
// The prompt is private to its chooser (PendingChoice.private): the
// card is in a hidden zone, and naming it on every seat's wire would
// tell the table what is in an opening hand that its owner may yet keep
// there. The table still sees that the seat owes an answer, which it
// has to — the game waits on it.
//
// Caller must hold g.mu.
func (g *Game) queueOpeningHandOfferLocked(seat, cardID uuid.UUID, name string, act *OpeningHandAction) {
	id := g.QueueConfirmForEffect(ConfirmPrompt{
		Chooser:      seat,
		Source:       cardID,
		Question:     openingHandQuestion(name, act),
		AcceptLabel:  "Begin the game with it",
		DeclineLabel: "Keep it in my hand",
		OnAccept: func(g *Game) error {
			return g.takeOpeningHandActionLocked(seat, cardID, act)
		},
	})
	if _, c := g.findChoiceLocked(id); c != nil {
		c.private = true
	}
}

// takeOpeningHandActionLocked puts the card onto the battlefield from
// its owner's hand and does what the rest of the sentence says. The card
// must still be in that hand: a prompt answered after the hand changed
// (a departed seat's, a rider that exiled it) does nothing rather than
// conjure a card.
//
// The entry is the ordinary hand door (PutFromHandOntoBattlefieldThen
// ForEffect), so the CR 614 window, the layer timestamp and the card's
// own enters hooks are the same as any other arrival, and a card that is
// not a permanent card is refused there (CR 110.4) rather than here.
//
// Caller must hold g.mu.
func (g *Game) takeOpeningHandActionLocked(seat, cardID uuid.UUID, act *OpeningHandAction) error {
	p := g.playerByIDLocked(seat)
	if p == nil || p.Hand == nil || !p.Hand.Contains(cardID) {
		return nil
	}
	return g.PutFromHandOntoBattlefieldThenForEffect(cardID, HandEntryOptions{}, func(g *Game, entered uuid.UUID) error {
		if entered == uuid.Nil {
			return nil
		}
		for name, n := range act.EntersWithCounters {
			if err := g.AddCounterForEffect(entered, name, n); err != nil {
				return err
			}
		}
		return g.queueOpeningHandExileLocked(seat, entered, act.ExileFromHand)
	})
}

// queueOpeningHandExileLocked is "exile a card from your hand": its
// owner picks `n` of what is left in their hand, and those cards go to
// exile. Nothing is asked of an empty hand, and a hand of exactly `n`
// still asks (the prompt does not short-circuit a forced pick, for the
// reason QueueChooseCardsForEffect gives).
//
// Caller must hold g.mu.
func (g *Game) queueOpeningHandExileLocked(seat, source uuid.UUID, n int) error {
	if n <= 0 {
		return nil
	}
	p := g.playerByIDLocked(seat)
	if p == nil || p.Hand == nil || len(p.Hand.Cards) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(p.Hand.Cards))
	for i := range p.Hand.Cards {
		ids = append(ids, p.Hand.Cards[i].InstanceID)
	}
	if n > len(ids) {
		n = len(ids)
	}
	g.QueueChooseCardsForEffect(ChooseCardsPrompt{
		Chooser:  seat,
		Source:   source,
		Question: "Exile a card from your hand",
		Cards:    ids,
		Min:      n,
		Max:      n,
		Zone:     ZoneHand,
		Then: func(g *Game, picked []uuid.UUID) error {
			for _, id := range picked {
				if err := g.ExileCardForEffect(id); err != nil {
					return err
				}
			}
			return nil
		},
	})
	return nil
}
