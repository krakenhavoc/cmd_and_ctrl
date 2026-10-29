package game

import (
	"testing"

	"github.com/google/uuid"
)

// settled_split_test.go — #4294: the myriad split of a token creation
// (token_attack_choices.go). The replacement window has settled when it
// asks, so the questions are the engine's attack split with the settled
// creation as their data, and a table waiting on any of them is a restore
// point.

// A settled creation whose groups name opponents is split like any attack
// split: the groups the replacements left, the one being split and the
// ones placed ride the prompt, so a restore at either question makes the
// same tokens, in the same order, and runs the keyed tail once.
func TestSettledCreationSplitSurvivesARestoreAtEachQuestion(t *testing.T) {
	attackSplitTailSeen = nil
	g := newActiveGameWithSeats(t, 3)
	me, opp1, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	for _, p := range []*Player{opp1, opp2} {
		g.Battlefield.Cards = append(g.Battlefield.Cards, Card{InstanceID: uuid.New(), Name: "Walker", TypeLine: "Planeswalker",
			Owner: p.ID, Controller: p.ID, StartingLoyalty: 4, Counters: map[string]int{CounterLoyalty: 4}})
	}
	tok := func(name string) Card {
		return Card{Name: name, TypeLine: "Token Creature — " + name, Power: 1, Toughness: 1}
	}
	groups := []TokenGroup{
		{Template: tok("Plain"), Count: 1},
		{Template: tok("First"), Count: 2, AttackOpponent: opp1.ID},
		{Template: tok("Second"), Count: 1, AttackOpponent: opp2.ID},
	}
	spec := TokenCreation{Controller: me.ID, Source: uuid.New(), Groups: groups, Together: true, AttackChoicesLabel: "Split Probe"}

	g.WithWriteLock(func() {
		if err := g.CreateTokensKeyedForEffect(spec, KeyedTail{Ref: attackSplitTail, Player: me.ID}); err != nil {
			t.Fatal(err)
		}
	})
	c := payAmountChoice(g)
	if c == nil || c.PickThen != settledSplitKey || c.AttackSplit == nil || c.PayAmount.Max != 2 || c.PayAmount.Goal != 2 {
		t.Fatalf("first question = %+v, want 0..2 for the first opponent's two copies", c)
	}
	if len(g.Battlefield.Cards) != 2 {
		t.Fatalf("%d permanents before the first answer, want only the two planeswalkers", len(g.Battlefield.Cards))
	}

	g = restoreStrictForTest(t, g)
	me = g.Seats[0]
	c = payAmountChoice(g)
	s := c.AttackSplit
	if s == nil || len(s.Targets) != 2 || len(s.Shares) != 0 || s.settled == nil || s.settled.Index != 1 ||
		len(s.settled.Chosen) != 1 || s.settled.Chosen[0].Template.Name != "Plain" || s.Tail.IsZero() {
		t.Fatalf("restored first question = %+v, want the plain group placed and the first opponent's split open", s)
	}
	// Both copies at the first thing offered.
	first := s.Targets[0]
	if err := g.ResolvePayAmount(c.ID, me.ID, 2); err != nil {
		t.Fatal(err)
	}

	g = restoreStrictForTest(t, g)
	me = g.Seats[0]
	c = payAmountChoice(g)
	s = c.AttackSplit
	if s == nil || s.settled == nil || s.settled.Index != 2 || len(s.settled.Chosen) != 2 || s.settled.Chosen[1].Template.Name != "First" ||
		s.settled.Chosen[1].Count != 2 || s.settled.Chosen[1].Attacking != first || s.settled.Chosen[1].AttackOpponent != uuid.Nil {
		t.Fatalf("restored second question = %+v, want the first opponent's two copies placed at %s", s, first)
	}
	if c.PayAmount.Max != 1 || c.PayAmount.Goal != 1 {
		t.Fatalf("second question = %+v, want the second opponent's one copy", c.PayAmount)
	}
	second := s.Targets[1]
	if err := g.ResolvePayAmount(c.ID, me.ID, 0); err != nil {
		t.Fatal(err)
	}
	if payAmountChoice(g) != nil {
		t.Fatal("a third question: every group is placed")
	}

	at := map[string]uuid.UUID{}
	count := map[string]int{}
	for _, card := range g.Battlefield.Cards {
		if card.IsToken() {
			at[card.Name] = card.AttackingTarget
			count[card.Name]++
		}
	}
	if count["Plain"] != 1 || count["First"] != 2 || count["Second"] != 1 {
		t.Fatalf("tokens = %v, want one Plain, two First and one Second", count)
	}
	if at["Plain"] != uuid.Nil || at["First"] != first || at["Second"] != second {
		t.Errorf("attacking = %v, want Plain unattached, First at %s and Second at %s", at, first, second)
	}
	if len(attackSplitTailSeen) != 1 || len(attackSplitTailSeen[0]) != 4 {
		t.Errorf("the tail ran %v, want once with the four tokens", attackSplitTailSeen)
	}
}

// A split whose chooser has left is not asked: every share is 0 and the
// last takes the rest, and the split still has its state to do that with (a
// keyed branch run without a prompt used to be handed none).
func TestSplitOfAChooserWhoLeftAsksNothingAndKeepsItsState(t *testing.T) {
	attackSplitTailSeen = nil
	g := newActiveGameWithSeats(t, 4)
	me := g.Seats[0]
	var targets []uuid.UUID
	for _, s := range g.Seats[1:] {
		targets = append(targets, s.ID)
	}
	me.Eliminated = true
	g.WithWriteLock(func() {
		err := g.StartAttackSplitForEffect(AttackSplit{
			Controller: me.ID, Source: uuid.New(), Count: 2, Label: "Split Probe",
			Template: Card{Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 1, Toughness: 1},
			Targets:  targets, Tail: KeyedTail{Ref: attackSplitTail, Player: me.ID},
		})
		if err != nil {
			t.Fatalf("a split of a chooser who left: %v", err)
		}
	})
	if payAmountChoice(g) != nil {
		t.Fatal("a question was asked of a chooser who left")
	}
	if len(attackSplitTailSeen) != 1 {
		t.Errorf("the tail ran %v, want once", attackSplitTailSeen)
	}
}
