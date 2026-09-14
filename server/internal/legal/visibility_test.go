package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// visibility_test.go holds the half of ADR 0033 §3 that lives in this
// package rather than in protocol.
//
// A Policy's Input has two halves and only one of them is filtered.
// Input.View goes through protocol.FilterViewFor, which is where all
// the redaction machinery is and where the aiseat package proves the
// result is byte-identical to a human's frame. Input.Moves does not:
// aiseat's runner builds it with legal.EnumerateFor straight off the
// authoritative *game.Game, and legal.Move.Label is free-form text
// assembled by reading Card.Name out of whatever zone the move
// touches. Since sub-PR 2 the same labels ride the wire as the
// viewer's own `legal_moves`, so a label that names a card the seat
// has not seen leaks to a human client too.
//
// Almost every branch enumerates over the seat's OWN cards or over a
// public zone, where there is nothing to withhold. Two do not: the
// discard clause reads FromPlayer's hand, and the search clause reads
// a library. Both are safe today by accident of their callers —
// QueueDiscardFromRevealedHand reveals the hand to the chooser before
// queueing, queueSearchChoiceLocked marks the searcher a knower of
// every match — and neither guarantee is written down anywhere the
// author of the next coercive-discard card would look.
//
// So the enumerator holds the line itself, and these are the tests
// that say so.

// discardChoice queues a discard_from_hand choice with the given
// chooser over fromPlayer's hand, WITHOUT revealing anything. That is
// the shape no card produces today; it is the shape the next one
// might.
func discardChoice(t *testing.T, g *game.Game, chooser, from uuid.UUID, count int) {
	t.Helper()
	g.WithWriteLock(func() {
		g.QueueChoiceForEffect(game.PendingChoice{
			Kind:       game.PendingChoiceDiscardFromHand,
			Chooser:    chooser,
			FromPlayer: from,
			Count:      count,
			Reason:     "Coercive discard",
		})
	})
}

func handNames(p *game.Player) []string {
	out := make([]string, 0, len(p.Hand.Cards))
	for _, c := range p.Hand.Cards {
		out = append(out, c.Name)
	}
	return out
}

func labelsOf(moves []legal.Move) string {
	var b strings.Builder
	for _, m := range moves {
		b.WriteString(m.Label)
		b.WriteByte('\n')
	}
	return b.String()
}

// TestDiscardMovesDoNotNameAnUnrevealedHand: a seat told to pick a
// card out of somebody else's hand, with nothing revealed to it, gets
// moves it can act on and labels that name nothing.
//
// The IDs stay in Params — they have to, the action is "discard this
// one" — and that is the correct line rather than a compromise. The
// chooser is being asked to pick from a hand it is looking at the
// backs of; an index into those backs is the whole mechanic. What it
// must not get is the front of the card.
func TestDiscardMovesDoNotNameAnUnrevealedHand(t *testing.T) {
	g := newTable(t)
	chooser, victim := g.Seats[0], g.Seats[1]
	discardChoice(t, g, chooser.ID, victim.ID, 1)

	moves := legal.EnumerateFor(g, chooser.ID)
	if len(moves) == 0 {
		t.Fatal("no moves offered for an owed discard choice")
	}
	labels := labelsOf(moves)
	for _, name := range handNames(victim) {
		if strings.Contains(labels, name) {
			t.Errorf("a move label named %q out of an unrevealed hand:\n%s", name, labels)
		}
	}
	if !strings.Contains(labels, "Coercive discard: discard a card") {
		t.Errorf("expected the anonymous discard label, got:\n%s", labels)
	}
}

// The control, and the more important of the two: redaction must be
// driven by the knower set, not by a blanket refusal to name anything
// in a hand. Thoughtseize reveals, and a bot that cannot read the
// hand it just made an opponent reveal is a worse bot for no reason.
func TestDiscardMovesNameARevealedHand(t *testing.T) {
	g := newTable(t)
	chooser, victim := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		g.QueueDiscardFromRevealedHand(chooser.ID, victim.ID, uuid.Nil, 1, "Thoughtseize")
	})

	moves := legal.EnumerateFor(g, chooser.ID)
	if len(moves) == 0 {
		t.Fatal("no moves offered for an owed discard choice")
	}
	labels := labelsOf(moves)
	named := 0
	for _, name := range handNames(victim) {
		if strings.Contains(labels, name) {
			named++
		}
	}
	if named != len(victim.Hand.Cards) {
		t.Errorf("the chooser was shown the hand but only %d of %d cards are named:\n%s",
			named, len(victim.Hand.Cards), labels)
	}
}

// The other seat at the table is not the chooser and is owed nothing,
// so it should be offered no answer to this choice at all — the pool
// never reaches its move list in the first place. Belt and braces
// against a future refactor that enumerates choices by kind rather
// than by chooser.
func TestDiscardMovesAreNotOfferedToNonChoosers(t *testing.T) {
	g := newTable(t)
	chooser, victim, other := g.Seats[0], g.Seats[1], g.Seats[2]
	g.WithWriteLock(func() {
		g.QueueDiscardFromRevealedHand(chooser.ID, victim.ID, uuid.Nil, 1, "Thoughtseize")
	})

	labels := labelsOf(legal.EnumerateFor(g, other.ID))
	if strings.Contains(labels, "Thoughtseize") {
		t.Errorf("a non-chooser was offered an answer to somebody else's choice:\n%s", labels)
	}
	for _, name := range handNames(victim) {
		if strings.Contains(labels, name) {
			t.Errorf("a non-chooser's move labels named %q from a hand revealed to somebody else:\n%s", name, labels)
		}
	}
}

// A seat's own hand is its own: mulligan and discard-to-hand-size
// labels must keep naming cards, or the redaction has overshot into
// hiding a player's hand from themselves.
func TestOwnHandIsStillNamedInMoveLabels(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[0]
	discardChoice(t, g, seat.ID, seat.ID, 1)

	labels := labelsOf(legal.EnumerateFor(g, seat.ID))
	named := 0
	for _, name := range handNames(seat) {
		if strings.Contains(labels, name) {
			named++
		}
	}
	if named != len(seat.Hand.Cards) {
		t.Errorf("the seat can only see %d of its own %d hand cards in its move labels:\n%s",
			named, len(seat.Hand.Cards), labels)
	}
}
