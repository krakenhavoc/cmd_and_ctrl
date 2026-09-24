package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// library_order_leftovers_test.go — #1298, ADR 0088's 2026-09-23
// amendment: the four library positions put_in_library did not cover.
// A look at ANOTHER player's library, a top lane at a depth, an exact
// top-lane count, and a counter whose chooser is not the card's owner.

// A look at another player's library marks the LOOKER, and only the
// looker; the library's owner learns nothing (CR 401.2, CR 701.20).
func TestLookAtAnotherPlayersLibraryMarksOnlyTheLooker(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	ids := loSeed(opp.Library, opp.ID, "Top", "Second", "Third")
	var looked []uuid.UUID
	g.WithWriteLock(func() { looked = g.LookAtTopOfPlayersLibraryForEffect(me.ID, opp.ID, 2) })
	if !loEqual(looked, ids[:2]) {
		t.Fatalf("looked at %v, want the top two of the OWNER's library %v", looked, ids[:2])
	}
	for _, id := range looked {
		c := g.findCardByIDLocked(id)
		if !c.IsKnownTo(me.ID) {
			t.Error("the looker knows what they looked at")
		}
		if c.IsKnownTo(opp.ID) || c.IsKnownTo(third.ID) {
			t.Error("a look is not a reveal: nobody else learns the cards, the owner included")
		}
	}
	if c := g.findCardByIDLocked(ids[2]); c.IsKnownTo(me.ID) {
		t.Error("only the top N are looked at")
	}
}

// The looker orders cards in the OTHER player's library: a reorder in
// place, and CR 401.4 leaves a lane of two known to the chooser alone.
func TestPutInLibraryReordersAnotherPlayersLibrary(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := loSeed(opp.Library, opp.ID, "A", "B", "C")
	for i := range opp.Library.Cards {
		opp.Library.Cards[i].AddKnower(opp.ID) // the owner had scryed them
	}
	var looked []uuid.UUID
	g.WithWriteLock(func() { looked = g.LookAtTopOfPlayersLibraryForEffect(me.ID, opp.ID, 2) })
	loQueue(t, g, PutInLibrarySpec{Chooser: me.ID, Cards: looked, From: ZoneLibrary, Placement: LibraryPlaceTop})
	c := loPrompt(t, g)
	if c.Chooser != me.ID {
		t.Fatal("the looker chooses")
	}
	if err := g.ResolvePutInLibrary(c.ID, opp.ID, nil, []uuid.UUID{ids[1], ids[0]}); !errors.Is(err, ErrNotTheChooser) {
		t.Errorf("the library's owner is not the chooser: %v", err)
	}
	if err := g.ResolvePutInLibrary(c.ID, me.ID, nil, []uuid.UUID{ids[1], ids[0]}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := loTop(opp.Library, 3); !loEqual(got, []uuid.UUID{ids[1], ids[0], ids[2]}) {
		t.Errorf("opponent's top three %v, want [B A C]", got)
	}
	if me.Library.Contains(ids[0]) || me.Library.Contains(ids[1]) {
		t.Error("each card stays in its OWNER's library")
	}
	for _, id := range ids[:2] {
		card := g.findCardByIDLocked(id)
		if card.IsKnownTo(opp.ID) {
			t.Error("CR 401.4: the owner does not learn the order someone else chose")
		}
		if !card.IsKnownTo(me.ID) {
			t.Error("the chooser knows the order")
		}
	}
}

// TopDepth: the top lane lands N from the top, first card at the depth
// — both as a reorder and through the tuck route.
func TestPutInLibraryTopDepthLandsSecondFromTheTop(t *testing.T) {
	t.Run("from the battlefield, through the tuck route", func(t *testing.T) {
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		lib := loSeed(opp.Library, opp.ID, "Top", "Next")
		perm := loSeed(g.Battlefield, opp.ID, "Permanent")[0]
		loQueue(t, g, PutInLibrarySpec{
			Chooser: opp.ID, Cards: []uuid.UUID{perm}, From: ZoneBattlefield,
			Placement: LibraryPlaceTopOrBottom, TopDepth: 2,
		})
		c := loPrompt(t, g)
		if c.LibraryTopDepth != 2 || c.Chooser != opp.ID {
			t.Fatalf("prompt %+v", c)
		}
		if err := g.ResolvePutInLibrary(c.ID, opp.ID, nil, []uuid.UUID{perm}); err != nil {
			t.Fatalf("ResolvePutInLibrary: %v", err)
		}
		if got := loTop(opp.Library, 3); !loEqual(got, []uuid.UUID{lib[0], perm, lib[1]}) {
			t.Errorf("top three %v, want [Top Permanent Next]", got)
		}
		card := g.findCardByIDLocked(perm)
		if !card.IsKnownTo(me.ID) || !card.IsKnownTo(opp.ID) {
			t.Error("a permanent alone in its lane was seen going there by the whole table")
		}
	})
	t.Run("a pile at a depth stays top-first", func(t *testing.T) {
		g := newActiveGame(t)
		me := g.Seats[0]
		lib := loSeed(me.Library, me.ID, "Top", "Next")
		hand := loSeed(me.Hand, me.ID, "X", "Y")
		loQueue(t, g, PutInLibrarySpec{Chooser: me.ID, Cards: hand, From: ZoneHand, Placement: LibraryPlaceTop, TopDepth: 2})
		c := loPrompt(t, g)
		if err := g.ResolvePutInLibrary(c.ID, me.ID, nil, []uuid.UUID{hand[1], hand[0]}); err != nil {
			t.Fatalf("ResolvePutInLibrary: %v", err)
		}
		if got := loTop(me.Library, 4); !loEqual(got, []uuid.UUID{lib[0], hand[1], hand[0], lib[1]}) {
			t.Errorf("top four %v, want [Top Y X Next]", got)
		}
	})
	t.Run("a reorder at a depth", func(t *testing.T) {
		g := newActiveGame(t)
		me := g.Seats[0]
		lib := loSeed(me.Library, me.ID, "Top", "Next", "Third")
		loQueue(t, g, PutInLibrarySpec{
			Chooser: me.ID, Cards: []uuid.UUID{lib[0]}, From: ZoneLibrary,
			Placement: LibraryPlaceTopOrBottom, TopDepth: 2,
		})
		c := loPrompt(t, g)
		if err := g.ResolvePutInLibrary(c.ID, me.ID, nil, []uuid.UUID{lib[0]}); err != nil {
			t.Fatalf("ResolvePutInLibrary: %v", err)
		}
		if got := loTop(me.Library, 3); !loEqual(got, []uuid.UUID{lib[1], lib[0], lib[2]}) {
			t.Errorf("top three %v, want [Next Top Third]", got)
		}
	})
}

// TopCount: exactly that many on top. Too few and too many are both
// refused; a pile that fits is a plain top placement.
func TestPutInLibraryTopCountIsExact(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := loSeed(me.Library, me.ID, "A", "B", "C", "Fourth")
	loQueue(t, g, PutInLibrarySpec{
		Chooser: me.ID, Cards: ids[:3], From: ZoneLibrary,
		Placement: LibraryPlaceTopOrBottom, TopCount: 1,
	})
	c := loPrompt(t, g)
	if c.LibraryTopCount != 1 || c.LibraryPlacement != LibraryPlaceTopOrBottom {
		t.Fatalf("prompt %+v", c)
	}
	for name, ans := range map[string][2][]uuid.UUID{
		"none on top": {{ids[0], ids[1], ids[2]}, nil},
		"two on top":  {{ids[2]}, {ids[0], ids[1]}},
	} {
		if err := g.ResolvePutInLibrary(c.ID, me.ID, ans[0], ans[1]); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("%s: err %v, want ErrInvalidParam", name, err)
		}
	}
	if err := g.ResolvePutInLibrary(c.ID, me.ID, []uuid.UUID{ids[2], ids[0]}, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := loTop(me.Library, 2); !loEqual(got, []uuid.UUID{ids[1], ids[3]}) {
		t.Errorf("top two %v, want [B Fourth]", got)
	}
	if got := loBottom(me.Library, 2); !loEqual(got, []uuid.UUID{ids[2], ids[0]}) {
		t.Errorf("bottom two %v, want [C A]", got)
	}
}

func TestPutInLibraryTopCountThatFitsIsATopPlacement(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := loSeed(me.Library, me.ID, "A", "B")
	loQueue(t, g, PutInLibrarySpec{
		Chooser: me.ID, Cards: ids, From: ZoneLibrary, Placement: LibraryPlaceTopOrBottom, TopCount: 2,
	})
	c := loPrompt(t, g)
	if c.LibraryPlacement != LibraryPlaceTop || c.LibraryTopCount != 0 {
		t.Errorf("two cards, two on top: only the order is left — got %q / %d", c.LibraryPlacement, c.LibraryTopCount)
	}
	if err := g.ResolvePutInLibrary(c.ID, me.ID, nil, []uuid.UUID{ids[1], ids[0]}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}

	one := loSeed(me.Library, me.ID, "Solo")[0]
	loQueue(t, g, PutInLibrarySpec{
		Chooser: me.ID, Cards: []uuid.UUID{one}, From: ZoneLibrary, Placement: LibraryPlaceTopOrBottom, TopCount: 1,
	})
	if len(g.PendingChoices) != 0 {
		t.Error("\"put one of those cards on top\" over one card asks nothing")
	}
	if top, _ := me.Library.Top(); top.InstanceID != one {
		t.Error("and leaves it on top")
	}
}

func TestPutInLibraryRefusesAMalformedSpec(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := loSeed(me.Library, me.ID, "A", "B")
	for name, spec := range map[string]PutInLibrarySpec{
		"a count on a one-lane placement": {Placement: LibraryPlaceTop, TopCount: 1},
		"a negative count":                {Placement: LibraryPlaceTopOrBottom, TopCount: -1},
		"a negative depth":                {Placement: LibraryPlaceTopOrBottom, TopDepth: -1},
	} {
		spec.Chooser, spec.Cards, spec.From = me.ID, ids, ZoneLibrary
		var err error
		g.WithWriteLock(func() { err = g.PutInLibraryInChosenOrderThenForEffect(spec) })
		if !errors.Is(err, ErrInvalidParam) {
			t.Errorf("%s: err %v, want ErrInvalidParam", name, err)
		}
	}
}

// Counter: the choice comes first with the spell still on the stack,
// then the leg is a COUNTER to the chosen end of its OWNER's library —
// the chooser is not the owner (Hinder).
func TestPutInLibraryCounterCountersToTheChosenEnd(t *testing.T) {
	for _, bottom := range []bool{false, true} {
		name := "top"
		if bottom {
			name = "bottom"
		}
		t.Run(name, func(t *testing.T) {
			g := newActiveGameWithSeats(t, 3)
			me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
			lib := loSeed(opp.Library, opp.ID, "Top", "Bottom")
			spell := uuid.New()
			pushStackSpell(t, g, Card{InstanceID: spell, Name: "Their Spell", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
			loQueue(t, g, PutInLibrarySpec{
				Chooser: me.ID, Cards: []uuid.UUID{spell}, Placement: LibraryPlaceTopOrBottom, Counter: true,
			})
			c := loPrompt(t, g)
			if !g.Stack.Contains(spell) {
				t.Fatal("nothing moves before the choice")
			}
			var top, under []uuid.UUID
			if bottom {
				under = []uuid.UUID{spell}
			} else {
				top = []uuid.UUID{spell}
			}
			if err := g.ResolvePutInLibrary(c.ID, me.ID, under, top); err != nil {
				t.Fatalf("ResolvePutInLibrary: %v", err)
			}
			if g.Stack.Contains(spell) || opp.Graveyard.Contains(spell) {
				t.Fatal("countered to the library, not the graveyard")
			}
			if _, ok := g.StackMeta[spell]; ok {
				t.Error("the stack record retires with the counter")
			}
			if !hasEventFor(g, EventCounterSpell, spell) {
				t.Error("it IS a counter (CR 701.6a)")
			}
			want := []uuid.UUID{spell, lib[0], lib[1]}
			got := loTop(opp.Library, 3)
			if bottom {
				want = []uuid.UUID{spell}
				got = loBottom(opp.Library, 1)
			}
			if !loEqual(got, want) {
				t.Errorf("owner's library %v, want %v", got, want)
			}
			if me.Library.Contains(spell) {
				t.Error("it goes to its OWNER's library, not the chooser's")
			}
			card := g.findCardByIDLocked(spell)
			for _, seat := range []uuid.UUID{me.ID, opp.ID, third.ID} {
				if !card.IsKnownTo(seat) {
					t.Error("the table watched a spell leave the stack for the library")
				}
			}
		})
	}
}

// A spell that can't be countered was not "countered this way": it is
// neither asked about nor moved.
func TestPutInLibraryCounterSkipsASpellThatCantBeCountered(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	old := CatalogCantBeCountered
	CatalogCantBeCountered = func(oracleID string) bool { return oracleID == "test-cant-be-countered-1298" }
	t.Cleanup(func() { CatalogCantBeCountered = old })
	spell := uuid.New()
	pushStackSpell(t, g, Card{
		InstanceID: spell, Name: "Uncounterable", OracleID: "test-cant-be-countered-1298",
		Owner: opp.ID, Controller: opp.ID,
	})
	ran := 0
	loQueue(t, g, PutInLibrarySpec{
		Chooser: me.ID, Cards: []uuid.UUID{spell}, Placement: LibraryPlaceTopOrBottom, Counter: true,
		Then: func(*Game) error { ran++; return nil },
	})
	if len(g.PendingChoices) != 0 {
		t.Error("nothing to ask about a spell the counter cannot counter")
	}
	if !g.Stack.Contains(spell) || ran != 1 {
		t.Errorf("the spell stays on the stack and the effect goes on (stack %v, then ran %d)", g.Stack.Contains(spell), ran)
	}
}

// A commander countered to the bottom is offered the command zone on
// the way, and a decline still lands it where the counter said.
func TestPutInLibraryCounterOffersACommanderTheCommandZone(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	loSeed(opp.Library, opp.ID, "Top")
	cmd := seatCommander(t, g.Stack, opp)
	top, _ := g.Stack.Top()
	_, _ = g.Stack.Remove(cmd)
	pushStackSpell(t, g, top)
	loQueue(t, g, PutInLibrarySpec{Chooser: me.ID, Cards: []uuid.UUID{cmd}, Placement: LibraryPlaceBottom, Counter: true})
	if !g.Stack.Contains(cmd) {
		t.Fatal("the commander waits on the stack for its owner's answer")
	}
	prompt := expectCommanderPrompt(t, g, opp)
	if err := g.ResolveOptionalReplacement(prompt.ID, opp.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if opp.Library.Cards[0].InstanceID != cmd {
		t.Error("declining the command zone leaves the commander on the bottom, where the counter put it")
	}
}
