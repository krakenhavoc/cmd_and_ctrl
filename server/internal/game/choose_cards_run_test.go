package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// choose_cards_run_test.go pins #1743's engine half:
// ChooseCardsRunThenForEffect, "each player chooses from their own
// look" asked as ONE instruction. Each leg is an ordinary choose_cards
// prompt; what the run adds is one continuation after the last answer.

type chooseRunCall struct {
	ran int
	got PromptedPicks
}

// lookLeg looks at the top `n` cards of p's library for p and builds
// an optional "choose up to max" leg over them.
func lookLeg(g *Game, p *Player, source uuid.UUID, n, max int, validate func([]Card) bool) ChooseCardsPrompt {
	return ChooseCardsPrompt{
		Chooser:  p.ID,
		Source:   source,
		Question: "test — choose from your look",
		Cards:    g.LookAtTopOfLibraryForEffect(p.ID, n),
		Min:      0,
		Max:      max,
		Zone:     ZoneLibrary,
		Validate: validate,
	}
}

func chooseCardsRun(t *testing.T, g *Game, legs func() []ChooseCardsPrompt) (*chooseRunCall, int) {
	t.Helper()
	out := &chooseRunCall{}
	var queued int
	g.WithWriteLock(func() {
		var err error
		queued, err = g.ChooseCardsRunThenForEffect(legs(), func(_ *Game, picks PromptedPicks) error {
			out.ran++
			out.got = picks
			return nil
		})
		if err != nil {
			t.Fatalf("ChooseCardsRunThenForEffect: %v", err)
		}
	})
	return out, queued
}

func chooseCardsPromptFor(g *Game, seat uuid.UUID) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceChooseCards && c.Chooser == seat {
			return c
		}
	}
	return nil
}

func answerChooseCards(t *testing.T, g *Game, seat uuid.UUID, cards ...uuid.UUID) {
	t.Helper()
	c := chooseCardsPromptFor(g, seat)
	if c == nil {
		t.Fatalf("seat %s owes no choose_cards prompt", seat)
	}
	if err := g.ResolveChooseCards(c.ID, seat, cards); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
}

// TestAChooseCardsRunWaitsForEveryLeg: three seats, each over its own
// look, answered out of order. The continuation runs once, after the
// third ANSWER, with the seats in leg order.
func TestAChooseCardsRunWaitsForEveryLeg(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	seats := g.Seats[:3]
	source := departureTestSource(g, seats[0].ID, "Explore the Vastlands")
	top := map[uuid.UUID]uuid.UUID{}
	for _, p := range seats {
		topOfLibraryFor(p, "Below", "Instant")
		top[p.ID] = topOfLibraryFor(p, "Top", "Land")
	}

	run, queued := chooseCardsRun(t, g, func() []ChooseCardsPrompt {
		var legs []ChooseCardsPrompt
		for _, p := range seats {
			legs = append(legs, lookLeg(g, p, source, 2, 1, nil))
		}
		return legs
	})
	if queued != 3 {
		t.Fatalf("queued %d prompts, want 3", queued)
	}
	// Each seat is asked about its OWN look and nobody else's.
	for _, p := range seats {
		c := chooseCardsPromptFor(g, p.ID)
		if c == nil {
			t.Fatalf("seat %s was not asked", p.Name)
		}
		for _, id := range c.ChooseCards {
			if z := g.FindCardZoneForEffect(id); z == nil || z.Owner != p.ID {
				t.Errorf("seat %s was offered a card out of another library", p.Name)
			}
		}
	}

	for n, i := range []int{2, 0, 1} {
		if run.ran != 0 {
			t.Fatalf("the continuation ran with %d prompts still open", 3-n)
		}
		p := seats[i]
		if i == 1 {
			answerChooseCards(t, g, p.ID) // chooses nothing
			continue
		}
		answerChooseCards(t, g, p.ID, top[p.ID])
	}

	if run.ran != 1 {
		t.Fatalf("the continuation ran %d times, want exactly 1", run.ran)
	}
	for i, e := range run.got {
		if e.Seat != seats[i].ID {
			t.Errorf("entry %d is seat %s, want leg order (%s)", i, e.Seat, seats[i].Name)
		}
	}
	if !run.got.Picked(seats[0].ID) || !run.got.Picked(seats[2].ID) {
		t.Errorf("an answering seat is reported as choosing nothing: %v", run.got)
	}
	if run.got.Picked(seats[1].ID) {
		t.Errorf("the seat that chose nothing is reported as choosing %v", run.got.By(seats[1].ID))
	}
	if run.got.Count() != 2 {
		t.Errorf("the run chose %d cards, want 2", run.got.Count())
	}
}

// TestAChooseCardsRunLegKeepsItsOwnSetRule: a leg is an ordinary
// choose_cards prompt, so its Validate refuses a set and leaves the
// prompt open — the run does not move on a refused answer.
func TestAChooseCardsRunLegKeepsItsOwnSetRule(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, me.ID, "Explore the Vastlands")
	a := topOfLibraryFor(me, "A", "Land")
	b := topOfLibraryFor(me, "B", "Land")
	topOfLibraryFor(them, "C", "Land")

	notTwoLands := func(picked []Card) bool {
		lands := 0
		for _, c := range picked {
			if c.IsLand() {
				lands++
			}
		}
		return lands <= 1
	}
	run, _ := chooseCardsRun(t, g, func() []ChooseCardsPrompt {
		return []ChooseCardsPrompt{
			lookLeg(g, me, source, 2, 2, notTwoLands),
			lookLeg(g, them, source, 1, 1, nil),
		}
	})

	c := chooseCardsPromptFor(g, me.ID)
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{a, b}); !errors.Is(err, ErrChoiceSetRejected) {
		t.Fatalf("two lands: err = %v, want ErrChoiceSetRejected", err)
	}
	if chooseCardsPromptFor(g, me.ID) == nil {
		t.Fatal("a refused answer took the prompt down")
	}
	answerChooseCards(t, g, me.ID, a)
	answerChooseCards(t, g, them.ID)
	if run.ran != 1 || len(run.got.By(me.ID)) != 1 || run.got.By(me.ID)[0] != a {
		t.Fatalf("run = %d %v, want one run with me choosing A", run.ran, run.got)
	}
}

// TestAChooseCardsRunSkipsALegWithNoCandidates: an empty library is a
// look at nothing, and nothing is not a question. With nobody to ask
// the rest of the card still runs, inline.
func TestAChooseCardsRunSkipsALegWithNoCandidates(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, me.ID, "Explore the Vastlands")
	mine := topOfLibraryFor(me, "Top", "Land")
	g.WithWriteLock(func() { them.Library.Cards = nil })

	run, queued := chooseCardsRun(t, g, func() []ChooseCardsPrompt {
		return []ChooseCardsPrompt{lookLeg(g, me, source, 5, 1, nil), lookLeg(g, them, source, 5, 1, nil)}
	})
	if queued != 1 || chooseCardsPromptFor(g, them.ID) != nil {
		t.Fatalf("queued %d prompts; the seat with an empty library must not be asked", queued)
	}
	answerChooseCards(t, g, me.ID, mine)
	if run.ran != 1 || len(run.got) != 1 || run.got[0].Seat != me.ID {
		t.Fatalf("run = %d %v, want one entry, for the seat that was asked", run.ran, run.got)
	}

	g.WithWriteLock(func() { me.Library.Cards = nil })
	none, queued := chooseCardsRun(t, g, func() []ChooseCardsPrompt {
		return []ChooseCardsPrompt{lookLeg(g, me, source, 5, 1, nil), lookLeg(g, them, source, 5, 1, nil)}
	})
	if queued != 0 || none.ran != 1 || len(none.got) != 0 {
		t.Fatalf("nobody to ask: queued %d, ran %d, got %v — want 0, 1, nothing", queued, none.ran, none.got)
	}
}

// TestAChooseCardsRunLegSettlesWhenItsChooserLeaves: a seat that
// concedes mid-prompt is dropped, its leg settles with nothing, and the
// survivor's answer still finishes the run.
func TestAChooseCardsRunLegSettlesWhenItsChooserLeaves(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, leaver := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, me.ID, "Explore the Vastlands")
	mine := topOfLibraryFor(me, "Mine", "Land")
	topOfLibraryFor(leaver, "Theirs", "Land")

	run, _ := chooseCardsRun(t, g, func() []ChooseCardsPrompt {
		return []ChooseCardsPrompt{lookLeg(g, me, source, 1, 1, nil), lookLeg(g, leaver, source, 1, 1, nil)}
	})
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if run.ran != 0 {
		t.Fatalf("the continuation ran with the survivor's prompt open")
	}
	answerChooseCards(t, g, me.ID, mine)
	if run.ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1 — a departed chooser must not strand the run", run.ran)
	}
	if run.got.Picked(leaver.ID) {
		t.Errorf("the departed seat is reported as choosing %v", run.got.By(leaver.ID))
	}
}

// TestAnUndoAcrossAHalfAnsweredChooseCardsRunReplaysIdentically is the
// clone contract the run shares with the discard and sacrifice runs.
func TestAnUndoAcrossAHalfAnsweredChooseCardsRunReplaysIdentically(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, me.ID, "Explore the Vastlands")
	mine := topOfLibraryFor(me, "Mine", "Land")
	theirs := topOfLibraryFor(them, "Theirs", "Land")

	run, _ := chooseCardsRun(t, g, func() []ChooseCardsPrompt {
		return []ChooseCardsPrompt{lookLeg(g, me, source, 1, 1, nil), lookLeg(g, them, source, 1, 1, nil)}
	})
	var snapshot *Game
	g.WithWriteLock(func() { snapshot = g.cloneLocked() })
	answerChooseCards(t, g, me.ID, mine)
	g.WithWriteLock(func() { g.RestoreFrom(snapshot) })

	answerChooseCards(t, g, me.ID, mine)
	if run.ran != 0 {
		t.Fatalf("the replayed answer ran the continuation %d times, want 0", run.ran)
	}
	answerChooseCards(t, g, them.ID, theirs)
	if run.ran != 1 || run.got.Count() != 2 {
		t.Fatalf("run = %d %v, want one run choosing two cards — a list that grew across the undo would report three",
			run.ran, run.got)
	}
}
