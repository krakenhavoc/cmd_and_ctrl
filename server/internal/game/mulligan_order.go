package game

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// Mulligans in turn order (CR 103.5, issue #2237).
//
// After the starting player is determined, each player in turn order,
// starting with the starting player, decides whether to keep or
// mulligan. Everyone who mulliganed redraws, and the process repeats
// in turn order for those players only. A player who kept is done.
//
// State: Player.MulliganDecided is true for a seat that has answered in
// the current round. The seat that decides next is derived, never
// stored: the first seat from StartingSeat, in seat order, that is not
// eliminated, has not kept and has not answered this round. When no
// such seat is left the round is over; the seats that mulliganed (the
// ones that have not kept) get a fresh round, and when every live seat
// has kept the window closes.
//
// Engine simplification: a mulligan redraws at once instead of after
// the round. The redraw uses the seat's own shuffle stream and nobody
// else's decision reads it, so only the moment the new hand is dealt
// differs from the paper order, not what anyone can know.

// ErrNotYourMulligan is returned by KeepHand and Mulligan when another
// seat is the one deciding (CR 103.5), or this seat has already
// answered in the current round.
var ErrNotYourMulligan = errors.New("game: it is not your turn to decide on your opening hand")

// ErrNoMoreMulligans is returned by Mulligan inside the window once the
// seat's counted mulligans have reached the starting hand size: its
// opening hand would be zero cards (CR 103.5, ADR 0168 decision 1).
var ErrNoMoreMulligans = errors.New("game: no further mulligans: your opening hand would be zero cards")

// The London mulligan (CR 103.5, ADR 0168).
//
// A mulligan inside the window always draws OpeningHandSize. The cost
// is paid when the seat keeps: it puts N cards from its hand on the
// bottom of its library in an order it chooses, where N is its
// mulligans taken less the free first one of a multiplayer game
// (CR 103.5c, CR 800.6; a game that began with more than two players,
// CR 100.1b). The rule's own order bottoms as part of each mulligan;
// bottoming once at keep leads to the same games (ADR 0168 decision 2).
//
// The keep raises Brainstorm's pair of prompts: pick N of the hand,
// then order them on the bottom. Player.mulliganBottoming is true from
// the keep until the cards are down, and the window does not close
// while any live seat is still bottoming. The other seats go on
// deciding meanwhile: a seat that kept is not the decider any more.

// freeMulligansLocked is how many of a seat's mulligans cost nothing:
// one in a multiplayer game (CR 103.5c), none in a two-player game
// unless the table's FreeFirstMulligan house rule gives it one too
// (#3387). Caller must hold g.mu.
func (g *Game) freeMulligansLocked() int {
	if g.Settings.FreeFirstMulligan || g.startingPlayerCountLocked() > 2 {
		return 1
	}
	return 0
}

// countedMulligansLocked is p's mulligans that count toward the cards
// it bottoms and the mulligans it may take (CR 103.5c).
func (g *Game) countedMulligansLocked(p *Player) int {
	return max(0, p.MulligansTaken-g.freeMulligansLocked())
}

// MulliganBottomCountLocked is how many cards p puts on the bottom of
// its library if it keeps now: its counted mulligans, at most its hand.
// Caller must hold g.mu (read is enough); exported for the view
// builder and the legal-move enumerator, which already hold it.
func (g *Game) MulliganBottomCountLocked(p *Player) int {
	n := g.countedMulligansLocked(p)
	if p.Hand != nil {
		n = min(n, len(p.Hand.Cards))
	}
	return n
}

// MulliganNextBottomLocked is how many cards p would put on the bottom
// if it took one more mulligan and then kept: what the mulligan costs.
func (g *Game) MulliganNextBottomLocked(p *Player) int {
	return min(OpeningHandSize, max(0, p.MulligansTaken+1-g.freeMulligansLocked()))
}

// MulliganAllowedLocked reports whether p may take another mulligan
// inside the window: its counted mulligans have not yet reached the
// starting hand size (CR 103.5). Caller must hold g.mu.
func (g *Game) MulliganAllowedLocked(p *Player) bool {
	return g.countedMulligansLocked(p) < OpeningHandSize
}

// queueMulliganBottomLocked asks p, which has just kept, to put n cards
// from its hand on the bottom of its library in any order. A hand of
// exactly n skips the pick and asks only for the order, and one card
// has no order (PutInLibraryInChosenOrderThenForEffect places it).
// When the cards are down, the seat stops gating the window.
//
// Caller must hold g.mu.
func (g *Game) queueMulliganBottomLocked(p *Player, n int) {
	if p.Hand == nil || n <= 0 {
		return
	}
	hand := make([]uuid.UUID, 0, len(p.Hand.Cards))
	for i := range p.Hand.Cards {
		hand = append(hand, p.Hand.Cards[i].InstanceID)
	}
	n = min(n, len(hand))
	if n == 0 {
		return
	}
	seat := p.ID
	p.mulliganBottoming = true
	cards := "1 card"
	if n > 1 {
		cards = fmt.Sprintf("%d cards", n)
	}
	order := func(g *Game, picked []uuid.UUID) error {
		return g.PutInLibraryInChosenOrderThenForEffect(PutInLibrarySpec{
			Chooser:   seat,
			Cards:     picked,
			From:      ZoneHand,
			Placement: LibraryPlaceBottom,
			Reason:    "London mulligan — in what order do they go on the bottom? (the first is nearest the top)",
			// The cards are down: this seat stops gating the window.
			Then: func(g *Game) error {
				if q := g.playerByIDLocked(seat); q != nil {
					q.mulliganBottoming = false
				}
				g.settleMulliganLocked()
				return nil
			},
		})
	}
	if n == len(hand) {
		_ = order(g, hand)
		return
	}
	id := g.QueueChooseCardsForEffect(ChooseCardsPrompt{
		Chooser:  seat,
		Question: fmt.Sprintf("London mulligan — put %s from your hand on the bottom of your library", cards),
		Cards:    hand,
		Min:      n,
		Max:      n,
		Zone:     ZoneHand,
		Then:     order,

		Destination: ChooseLibraryBottom,
	})
	// The hand is a hidden zone: only its owner sees what it is asked.
	if _, c := g.findChoiceLocked(id); c != nil {
		c.private = true
	}
}

// MulliganDeciderLocked returns the seat that decides next, or -1 when
// no seat has an open decision. Caller must hold g.mu (read is enough); exported for the
// view builder, which already holds it.
func (g *Game) MulliganDeciderLocked() int {
	n := len(g.Seats)
	if !g.MulligansOpen || g.OpeningRoll != nil || n == 0 {
		return -1
	}
	for k := 0; k < n; k++ {
		i := (g.StartingSeat + k) % n
		p := g.Seats[i]
		if p == nil || p.Eliminated || p.HandKept || p.MulliganDecided {
			continue
		}
		return i
	}
	return -1
}

// MulliganDecider is the seat that has to answer keep or mulligan now,
// or -1 when none does (no window, or the opening roll is open). Takes
// the read lock.
func (g *Game) MulliganDecider() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.MulliganDeciderLocked()
}

// settleMulliganLocked advances the mulligan window after any change
// to it: a keep, a mulligan, or a departure. It starts the next round
// for the seats that mulliganed, and closes the window when every live
// seat has kept. Safe to call at any time; it does nothing outside the
// window. Caller must hold g.mu.
func (g *Game) settleMulliganLocked() {
	if g.State != StateActive || !g.MulligansOpen || g.OpeningRoll != nil {
		return
	}
	if g.MulliganDeciderLocked() >= 0 {
		return
	}
	allKept, bottoming := true, false
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if !p.HandKept {
			allKept = false
			p.MulliganDecided = false
		}
		if p.mulliganBottoming {
			bottoming = true
		}
	}
	if !allKept {
		// Next round: the seats that mulliganed, in turn order again.
		return
	}
	if bottoming {
		// ADR 0168 decision 4: the hands are not final until every
		// keeper's cards are on the bottom. The last one to finish
		// settles the window again.
		return
	}
	g.MulligansOpen = false
	// CR 103.6: the starting hands are final, so the seats holding a
	// card that lets them begin the game with it on the battlefield are
	// asked, in turn order from the starting player (ADR 0133). The
	// prompts hold the table the way any open choice does.
	g.offerOpeningHandActionsLocked()
	// First-step entry happens here, not at Start: the cursor has been
	// parked on Untap with NoPriority since Start, waiting for everyone
	// to commit. Run the hook now so the starting seat's auto-untap
	// fires and the cursor advances to Upkeep. (S13.)
	g.runStepEntryHooksLocked()
}
