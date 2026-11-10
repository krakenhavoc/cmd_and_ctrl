package game

import (
	"testing"

	"github.com/google/uuid"
)

// clash_test.go — #5424, CR 701.30: the clash itself, without a card.
// The cards built on it are tested in cards/effects/clash_cards_test.go.

// clashTestTop puts a card with the given mana cost on top of p's library.
func clashTestTop(p *Player, name, cost string) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = cost
	p.Library.PushTop(c)
	return c.InstanceID
}

func clashTestPrompts(g *Game) []*PendingChoice {
	var out []*PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceOptionPick {
			out = append(out, c)
		}
	}
	return out
}

func clashTestWinners(g *Game) []uuid.UUID {
	var out []uuid.UUID
	for _, ev := range g.Events {
		if ev.Kind == EventClashWon {
			out = append(out, ev.Actor)
		}
	}
	return out
}

// clashTestTold is a tail that records what a clash told it.
var (
	clashTestToldAmount = -1
	clashTestTold       = RegisterTail("test/clash-told", func(_ *Game, o TailOutcome) error {
		clashTestToldAmount = o.Amount
		return nil
	})
)

// TestClashAsksInAPNAPOrderWhoeverStartedIt — the clasher who is not the
// active player started the clash; the active player still decides first
// (CR 701.30c, CR 101.4), and the tail is told its own player lost.
func TestClashAsksInAPNAPOrderWhoeverStartedIt(t *testing.T) {
	g := newActiveGame(t)
	active, other := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := clashTestTop(other, "Other's One", "{U}")
	theirs := clashTestTop(active, "Active's Three", "{2}{U}")
	clashTestToldAmount = -1

	g.WithWriteLock(func() {
		if err := g.ClashForEffect(uuid.Nil, other.ID, active.ID, KeyedTail{Ref: clashTestTold, Player: other.ID}); err != nil {
			t.Fatalf("ClashForEffect: %v", err)
		}
	})
	open := clashTestPrompts(g)
	if len(open) != 1 || open[0].Chooser != active.ID {
		t.Fatalf("the first decision is not the active player's alone (%d open)", len(open))
	}
	if err := g.ResolveOptionPick(open[0].ID, active.ID, 1); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if clashTestToldAmount != -1 || len(clashTestWinners(g)) != 0 {
		t.Fatal("the clash finished before every decision was in")
	}
	open = clashTestPrompts(g)
	if len(open) != 1 || open[0].Chooser != other.ID {
		t.Fatalf("the second decision is not the other clasher's (%d open)", len(open))
	}
	if err := g.ResolveOptionPick(open[0].ID, other.ID, 0); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}

	if active.Library.Cards[0].InstanceID != theirs {
		t.Error("the active player's card was not put on the bottom")
	}
	if other.Library.Cards[other.Library.Size()-1].InstanceID != mine {
		t.Error("the other clasher's card did not stay on top")
	}
	if w := clashTestWinners(g); len(w) != 1 || w[0] != active.ID {
		t.Errorf("winners = %v, want the player with the higher mana value", w)
	}
	if clashTestToldAmount != 0 {
		t.Errorf("the tail was told %d, want 0 (its player lost)", clashTestToldAmount)
	}
}

// TestClashWithTwoEmptyLibrariesHasNoWinner — nothing to reveal, nothing
// to ask, nobody wins, and the rest of the card still runs.
func TestClashWithTwoEmptyLibrariesHasNoWinner(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	clashTestToldAmount = -1

	g.WithWriteLock(func() {
		me.Library.Cards, opp.Library.Cards = nil, nil
		if err := g.ClashForEffect(uuid.Nil, me.ID, opp.ID, KeyedTail{Ref: clashTestTold, Player: me.ID}); err != nil {
			t.Fatalf("ClashForEffect: %v", err)
		}
	})
	if len(clashTestPrompts(g)) != 0 {
		t.Error("a player with no card to place was asked where to put it")
	}
	if len(clashTestWinners(g)) != 0 {
		t.Error("a clash with no card revealed had a winner")
	}
	if clashTestToldAmount != 0 {
		t.Errorf("the tail was told %d, want 0", clashTestToldAmount)
	}
}

// TestClashFinishesWhenAClasherLeaves — a clasher who concedes with their
// decision open leaves the card where it is, and the clash still finishes:
// the other clasher is asked, their card moves and the tail runs.
func TestClashFinishesWhenAClasherLeaves(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	active, other := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clashTestTop(active, "Active's One", "{U}")
	mine := clashTestTop(other, "Other's Three", "{2}{U}")
	clashTestToldAmount = -1

	g.WithWriteLock(func() {
		if err := g.ClashForEffect(uuid.Nil, other.ID, active.ID, KeyedTail{Ref: clashTestTold, Player: other.ID}); err != nil {
			t.Fatalf("ClashForEffect: %v", err)
		}
	})
	if err := g.Concede(active.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	open := clashTestPrompts(g)
	if len(open) != 1 || open[0].Chooser != other.ID {
		t.Fatalf("the remaining clasher was not asked (%d open)", len(open))
	}
	if err := g.ResolveOptionPick(open[0].ID, other.ID, 1); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if other.Library.Cards[0].InstanceID != mine {
		t.Error("the remaining clasher's card was not put on the bottom")
	}
	if w := clashTestWinners(g); len(w) != 1 || w[0] != other.ID {
		t.Errorf("winners = %v, want the remaining clasher, who revealed the higher card", w)
	}
	if clashTestToldAmount != 1 {
		t.Errorf("the tail was told %d, want 1", clashTestToldAmount)
	}
}
