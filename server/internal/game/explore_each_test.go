package game

import (
	"testing"

	"github.com/google/uuid"
)

// explore_each_test.go — #5425: the explore replacement "explores, then
// explores again" (CR 614.5) and CR 701.44d, several permanents told to
// explore at once.

// An explore's rest is words on its tails and prompts; a rest with nothing
// new in it is spelled as it was before #5425, so a table saved by an
// older binary resumes.
func TestExploreRestRoundTripsThroughItsWords(t *testing.T) {
	a, b := ObjectRef{ID: uuid.New(), Epoch: 3}, ObjectRef{ID: uuid.New()}
	for name, r := range map[string]exploreRest{
		"nothing":          {},
		"more":             {More: 2},
		"a tail":           {Then: exploreDoneTail.key},
		"more and a tail":  {More: 1, Then: exploreDoneTail.key},
		"again":            {Again: 3},
		"a queue":          {Queue: []ObjectRef{a, b}},
		"everything":       {More: 2, Then: exploreDoneTail.key, Again: 1, Queue: []ObjectRef{a}},
		"again and a tail": {Again: 1, Then: exploreDoneTail.key},
	} {
		got := exploreRestFromWord(r.word())
		if got.More != r.More || got.Then != r.Then || got.Again != r.Again || len(got.Queue) != len(r.Queue) {
			t.Errorf("%s: word %q read back as %+v, want %+v", name, r.word(), got, r)
			continue
		}
		for i := range r.Queue {
			if got.Queue[i] != r.Queue[i] {
				t.Errorf("%s: queue[%d] = %+v, want %+v", name, i, got.Queue[i], r.Queue[i])
			}
		}
		if fromTail := exploreRestFromTail(r.More, r.tailWord()); fromTail.Then != r.Then || fromTail.Again != r.Again || len(fromTail.Queue) != len(r.Queue) {
			t.Errorf("%s: tail word %q read back as %+v", name, r.tailWord(), fromTail)
		}
	}
	if w := (exploreRest{More: 2, Then: "some/key"}).word(); w != "2|some/key" {
		t.Errorf("a rest with nothing new is spelled %q, want the pre-#5425 %q", w, "2|some/key")
	}
	if w := (exploreRest{More: 2, Then: "some/key"}).tailWord(); w != "some/key" {
		t.Errorf("its tail word is %q, want the key alone", w)
	}
	if w := (exploreRest{}).word(); w != "" {
		t.Errorf("an empty rest is spelled %q, want nothing", w)
	}
}

func exploreAgainProbe(t *testing.T, g *Game, owner *Player) uuid.UUID {
	t.Helper()
	const key = testEffectKeyPrefix + "explore-again-probe"
	RegisterExploreAgainKey(key)
	id := pushColouredCreature(g, owner, "Again Probe", []string{"G"})
	g.WithWriteLock(func() {
		c := findBattlefieldCard(g, id)
		c.OracleID = key
	})
	c, _ := battlefieldCardByID(g, id)
	if catalogAbilityKeyOf(&c) != key {
		t.Skip("a bare card has no catalog key in this package; the replacement is covered by the catalog's tests")
	}
	return id
}

// The explores a replacement adds are not replaced again, and they are
// explores: each emits EventExplored.
func TestExploreAgainReplacementDoublesOneInstruction(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	exploreAgainProbe(t, g, me)
	id := pushColouredCreature(g, me, "Explorer", []string{"G"})
	me.Library.Cards = nil
	c, _ := battlefieldCardByID(g, id)
	g.WithWriteLock(func() {
		if err := g.ExploreForEffect(uuid.Nil, ObjectRef{ID: id, Epoch: c.ObjectEpoch}, me.ID); err != nil {
			t.Fatalf("ExploreForEffect: %v", err)
		}
	})
	if got := countExplored(g, id); got != 2 {
		t.Errorf("explored %d times, want 2", got)
	}
}

// CR 701.44d: the controller chooses one, it explores, and the process
// repeats; with one left there is nothing to choose.
func TestExploreEachAsksWhichExploresNext(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := []uuid.UUID{
		pushColouredCreature(g, me, "One", []string{"G"}),
		pushColouredCreature(g, me, "Two", []string{"G"}),
		pushColouredCreature(g, me, "Three", []string{"G"}),
	}
	me.Library.Cards = nil
	refs := make([]ObjectRef, len(ids))
	for i, id := range ids {
		c, _ := battlefieldCardByID(g, id)
		refs[i] = ObjectRef{ID: id, Epoch: c.ObjectEpoch}
	}
	g.WithWriteLock(func() {
		if err := g.ExploreEachForEffect(uuid.Nil, refs, me.ID); err != nil {
			t.Fatalf("ExploreEachForEffect: %v", err)
		}
	})
	pick := func() *PendingChoice {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == PendingChoiceOwnPermanents {
				return c
			}
		}
		return nil
	}
	first := pick()
	if first == nil || len(first.ChooseCards) != 3 || first.ChooseMin != 1 || first.ChooseMax != 1 {
		t.Fatalf("first choice = %+v, want exactly one of three", first)
	}
	if countExplored(g, ids[0])+countExplored(g, ids[1])+countExplored(g, ids[2]) != 0 {
		t.Fatal("a permanent explored before the choice")
	}
	if err := g.ResolveOwnPermanents(first.ID, me.ID, []uuid.UUID{ids[2]}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if countExplored(g, ids[2]) != 1 || countExplored(g, ids[0]) != 0 || countExplored(g, ids[1]) != 0 {
		t.Fatal("the chosen permanent did not explore first, alone")
	}
	second := pick()
	if second == nil || len(second.ChooseCards) != 2 {
		t.Fatalf("second choice = %+v, want one of the two left", second)
	}
	if err := g.ResolveOwnPermanents(second.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if pick() != nil {
		t.Error("asked which explores next with one left")
	}
	for i, id := range ids {
		if got := countExplored(g, id); got != 1 {
			t.Errorf("permanent %d explored %d times, want 1", i, got)
		}
	}
}

// CR 701.44c: a permanent that is no longer the object the instruction
// named still explores. It cannot be offered by a battlefield pick, so it
// goes first.
func TestExploreEachADepartedPermanentStillExplores(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	here := pushColouredCreature(g, me, "Here", []string{"G"})
	gone := uuid.New()
	me.Library.Cards = nil
	c, _ := battlefieldCardByID(g, here)
	g.WithWriteLock(func() {
		if err := g.ExploreEachForEffect(uuid.Nil, []ObjectRef{{ID: here, Epoch: c.ObjectEpoch}, {ID: gone}}, me.ID); err != nil {
			t.Fatalf("ExploreEachForEffect: %v", err)
		}
	})
	if countExplored(g, gone) != 1 || countExplored(g, here) != 1 {
		t.Errorf("explored: departed %d, present %d; want 1 each", countExplored(g, gone), countExplored(g, here))
	}
	for _, ch := range g.PendingChoices {
		if ch != nil && ch.Kind == PendingChoiceOwnPermanents {
			t.Error("asked a question with one permanent on the battlefield")
		}
	}
}

func TestExploreEachOfNothingDoesNothing(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		if err := g.ExploreEachForEffect(uuid.Nil, nil, g.Seats[0].ID); err != nil {
			t.Fatalf("ExploreEachForEffect: %v", err)
		}
	})
	if len(g.PendingChoices) != 0 {
		t.Error("an explore of no permanents asked something")
	}
}
