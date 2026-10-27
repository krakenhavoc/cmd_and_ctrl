package game

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// search_prompt_keyed_test.go — #4253 slice 5: a library-search prompt
// whose spec holds no closure set rule and no closure continuation is
// written with the table (pendingChoices[].searchPrompt), so a table
// waiting on it is a restore point. The catalog's restore tests are in
// effects/search_prompt_restore_test.go.

// testSearchRuleAtMost is a registered set rule for these tests: at most N
// cards. Under the test/ prefix, so never in the ledger.
var testSearchRuleAtMost = RegisterSearchRule("test/search-at-most-n", func(picked []Card, n int) bool {
	return len(picked) <= n
})

func isBasicNamed(name string) func(Card) bool {
	return func(c Card) bool { return c.Name == name }
}

// strictSearchRestore round-trips g through JSON and a strict restore,
// after checking the open search prompt was written.
func strictSearchRestore(t *testing.T, g *Game) *Game {
	t.Helper()
	snap := throughJSON(t, g.CaptureSnapshot())
	if !snap.Restorable() {
		t.Fatalf("the open search prompt is not a restore point: %+v", snap.Continuations)
	}
	written := false
	for _, c := range snap.PendingChoices {
		if c.Kind == PendingChoiceSearchLibrary && c.PickThen == SearchPromptKey && c.SearchPrompt != nil {
			written = true
		}
	}
	if !written {
		t.Fatal("the search prompt was not written as data")
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	return restored
}

// The prompt holds the spec without its predicate: the predicate listed
// the candidates and is read again nowhere. What the answer needs of it,
// that the search names a quality (so it may find nothing, CR 701.23b),
// is kept as a fact, live and restored alike.
func TestASearchPromptKeepsTheQualityAndDropsThePredicate(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := searchLibraryOf(me, "Island", "Swamp", "Island")
	runSearch(t, g, SearchLibrarySpec{Player: me.ID, Pred: isBasicNamed("Island"), Dest: ZoneHand, Shuffle: true})
	c := openSearch(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	// #4361 S4: the frame's type has no predicate to hold.
	if !c.searchResume.quality {
		t.Fatal("the live prompt does not record that the search names a quality")
	}
	if !slices.Equal(c.SearchCards, []uuid.UUID{ids[0], ids[2]}) {
		t.Fatalf("candidates %v, want the two Islands", c.SearchCards)
	}

	g = strictSearchRestore(t, g)
	me = g.Seats[0]
	c = openSearch(g, me.ID)
	if c == nil || c.SearchMin != 0 {
		t.Fatalf("the restored prompt %+v, want one that may find nothing", c)
	}
	if got := g.SearchFloorLocked(c); got != 0 {
		t.Fatalf("the restored floor is %d, want 0 for a stated quality", got)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
		t.Fatalf("finding nothing on the restored prompt: %v", err)
	}
	if searchEvents(g) != 1 {
		t.Errorf("%d search events, want 1", searchEvents(g))
	}
}

// An "and/or graveyard" search recomputes its floor from the spec on the
// restored table: an unconditional one still must find (CR 701.23d), so
// the quality bool is what decides it there.
func TestARestoredGraveyardSearchKeepsItsFloor(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	searchLibraryOf(me, "Island", "Swamp")
	gy := uuid.New()
	me.Graveyard.PushTop(Card{InstanceID: gy, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	runSearch(t, g, SearchLibrarySpec{Player: me.ID, Dest: ZoneHand, AlsoGraveyard: true})
	g = strictSearchRestore(t, g)
	me = g.Seats[0]
	c := openSearch(g, me.ID)
	if got := g.SearchFloorLocked(c); got != 1 {
		t.Fatalf("the restored unconditional graveyard search's floor is %d, want 1", got)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("finding nothing: %v, want ErrInvalidParam", err)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{gy}); err != nil {
		t.Fatalf("taking the graveyard card: %v", err)
	}
}

// A registered set rule is written by key with its number, judged on the
// restored table, and is a stated quality.
func TestASearchPromptWithARegisteredRuleIsARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := searchLibraryOf(me, "Island", "Swamp", "Mountain")
	runSearch(t, g, SearchLibrarySpec{Player: me.ID, Dest: ZoneHand, Limit: 3, Rule: testSearchRuleAtMost.WithN(1)})

	g = strictSearchRestore(t, g)
	me = g.Seats[0]
	c := openSearch(g, me.ID)
	if c.SearchMin != 0 {
		t.Fatalf("SearchMin %d: a set rule is a stated quality", c.SearchMin)
	}
	if g.SearchPickLegalLocked(c, ids[:2]) || !g.SearchPickLegalLocked(c, ids[:1]) {
		t.Fatal("the enumerator's window does not judge by the restored rule")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, ids[:2]); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("two cards against an at-most-one rule: %v, want ErrInvalidParam", err)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, ids[:1]); err != nil {
		t.Fatalf("one card: %v", err)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the pick did not reach the hand")
	}
}

// A search with a registered "then" runs it from the restored prompt.
func TestARestoredSearchPromptRunsItsKeyedThen(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := searchLibraryOf(me, "Island", "Swamp")
	var got []uuid.UUID
	ref := RegisterTail("test/search-prompt-then", func(_ *Game, o TailOutcome) error {
		got = append([]uuid.UUID(nil), o.IDs...)
		return nil
	})
	runSearch(t, g, SearchLibrarySpec{Player: me.ID, Pred: isBasicNamed("Swamp"), Dest: ZoneHand,
		ThenTail: KeyedTail{Ref: ref, Player: me.ID}})
	g = strictSearchRestore(t, g)
	c := openSearch(g, g.Seats[0].ID)
	if err := g.ResolveSearchLibrary(c.ID, c.Chooser, []uuid.UUID{ids[1]}); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, []uuid.UUID{ids[1]}) {
		t.Fatalf("the restored then was told %v, want the Swamp", got)
	}
}

// #4361 S4: a search has no closure form left, so every search prompt is
// written, whatever its predicate, set rule or tail: never counted, and the
// table is a restore point.
func TestEverySearchPromptIsWritten(t *testing.T) {
	for name, spec := range map[string]SearchLibrarySpec{
		"rule":      {Rule: testSearchRuleAtMost.WithN(2), Limit: 2},
		"predicate": {Pred: isBasicNamed("Island")},
		"tail":      {Pred: isBasicNamed("Island"), ThenTail: searchHook(func(*Game, []uuid.UUID) error { return nil })},
	} {
		g := newActiveGame(t)
		me := g.Seats[0]
		searchLibraryOf(me, "Island", "Swamp", "Island")
		spec.Player, spec.Dest = me.ID, ZoneHand
		runSearch(t, g, spec)
		snap := g.CaptureSnapshot()
		if !snap.Restorable() || snap.Continuations.ChoiceResumeFrames != 0 {
			t.Errorf("%s: a search prompt is not a restore point: %#v", name, snap.Continuations)
		}
		for _, c := range snap.PendingChoices {
			if c.SearchPrompt == nil || c.PickThen != SearchPromptKey || slices.Contains(c.ResumeFrames, "searchResume") {
				t.Errorf("%s: the prompt was not written (%v, %q) or was counted (%v)", name, c.SearchPrompt != nil, c.PickThen, c.ResumeFrames)
			}
		}
	}
}

// A spec with a rule this binary does not know is refused before any
// prompt opens.
func TestASearchWithTwoSetRulesOrAnUnknownRuleIsRefused(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	searchLibraryOf(me, "Island", "Swamp")
	for name, spec := range map[string]SearchLibrarySpec{
		"unknown": {Rule: SearchRule{key: "test/never-registered"}},
	} {
		spec.Player, spec.Dest = me.ID, ZoneHand
		var err error
		g.WithWriteLock(func() { err = g.SearchLibraryThenForEffect(spec) })
		if err == nil {
			t.Errorf("%s: the search was accepted", name)
		}
		if openSearch(g, me.ID) != nil {
			t.Fatalf("%s: a prompt opened", name)
		}
	}
}

// The marker without its frame, or the frame on another kind of prompt, is
// a shape this binary does not know: refused. An older binary refuses a
// written search prompt too, because it reads a search prompt's PickThen
// as a revealed-hand pick's continuation, a registry the marker is not in.
func TestAMalformedSearchPromptIsRefused(t *testing.T) {
	if KnownRevealedPickThen(SearchPromptKey) {
		t.Fatal("the search prompt's marker is a revealed-hand pick key, so an older binary would not refuse it")
	}
	build := func() *GameSnapshot {
		g := newActiveGame(t)
		me := g.Seats[0]
		searchLibraryOf(me, "Island", "Swamp")
		runSearch(t, g, SearchLibrarySpec{Player: me.ID, Pred: isBasicNamed("Island"), Dest: ZoneHand})
		return throughJSON(t, g.CaptureSnapshot())
	}
	snap := build()
	snap.PendingChoices[0].SearchPrompt = nil
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("the marker without its frame: %v, want ErrUnknownEffectKey", err)
	}
	snap = build()
	snap.PendingChoices[0].PickThen = ""
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("the frame without its marker: %v, want ErrUnknownEffectKey", err)
	}
	snap = build()
	snap.PendingChoices[0].Kind = PendingChoiceConfirm
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("the frame on a confirm: %v, want ErrUnknownEffectKey", err)
	}
	snap = build()
	snap.PendingChoices[0].SearchPrompt.Spec.ThenTail = &keyedTailSnapshot{Key: "test/never-registered"}
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("an unknown tail on the frame: %v, want ErrUnknownEffectKey", err)
	}
}
