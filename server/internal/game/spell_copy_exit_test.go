package game

import (
	"testing"

	"github.com/google/uuid"
)

// spell_copy_exit_test.go — #1340, CR 707.10a / CR 704.5e as a shape.
// Every exit a spell can take off the stack without resolving goes
// through routeCardToZoneLocked (#1318, #1345); a COPY taking any of
// them ceases to exist instead of landing. The cards are pinned in
// cards/effects/spell_copy_exit_test.go.

// copyOfSpellForTest pushes a plain instant and copies it, returning
// the original and the copy.
func copyOfSpellForTest(t *testing.T, g *Game, controller uuid.UUID) (orig, cp uuid.UUID) {
	t.Helper()
	orig = pushLKISpell(t, g, controller, "Lightning Bolt")
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(orig, controller, false, nil); err != nil {
			t.Fatalf("CopySpellForEffect: %v", err)
		}
		for id, item := range g.StackMeta {
			if item != nil && item.IsCopy {
				cp = id
			}
		}
	})
	if cp == uuid.Nil {
		t.Fatal("no copy on the stack")
	}
	return orig, cp
}

// copyEvents reports what was announced about the copy: whether it was
// countered, and every zone it was announced moving INTO.
func copyEvents(g *Game, cp uuid.UUID) (countered bool, landed []ZoneKind) {
	for _, ev := range g.Events {
		if ev.CardID != cp {
			continue
		}
		switch ev.Kind {
		case EventCounterSpell:
			countered = true
		case EventZoneMove, EventMill, EventDiscardCard:
			if ev.NewZone != "" {
				landed = append(landed, ev.NewZone)
			}
		}
	}
	return countered, landed
}

// assertCopyGone is the whole postcondition: in no zone, no stack
// record, and never announced landing anywhere.
func assertCopyGone(t *testing.T, g *Game, cp uuid.UUID) {
	t.Helper()
	if z := zoneOfCard(g, cp); z != "" {
		t.Errorf("the copy is in %q — a copy of a spell in a zone other than the stack ceases to exist (CR 707.10a)", z)
	}
	if _, ok := g.StackMeta[cp]; ok {
		t.Error("the copy's stack record outlived it")
	}
	if _, landed := copyEvents(g, cp); len(landed) != 0 {
		t.Errorf("the copy was announced landing in %v; it goes nowhere", landed)
	}
}

// TestACounteredSpellCopyCeasesToExist — the issue's reproduction. The
// counter still happened (EventCounterSpell), the copy is nowhere, the
// original is untouched, and the last-known record of the copy was
// taken on the way out, so an effect that NAMES it can still copy it
// (CR 608.2h, #1255).
func TestACounteredSpellCopyCeasesToExist(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	orig, cp := copyOfSpellForTest(t, g, me.ID)

	counterForTest(t, g, cp)
	g.WithWriteLock(func() { g.runStateChecksLocked() })

	assertCopyGone(t, g, cp)
	if countered, _ := copyEvents(g, cp); !countered {
		t.Error("no EventCounterSpell for the copy — it was countered, and counter watchers must see it")
	}
	if me.Graveyard.Size() != 0 {
		t.Errorf("graveyard holds %d cards, want 0", me.Graveyard.Size())
	}
	if !g.Stack.Contains(orig) {
		t.Error("the original left the stack with its copy")
	}
	g.WithWriteLock(func() {
		card, item, ok := g.lastKnownSpellLocked(cp)
		if !ok || !item.IsCopy || card.Name != "Lightning Bolt" {
			t.Fatalf("last-known record of the countered copy = %v / %+v, want the copy as it stood", ok, item)
		}
		before := g.Stack.Size()
		if err := g.CopyLastKnownSpellForEffect(cp, me.ID, false, nil); err != nil {
			t.Fatalf("CopyLastKnownSpellForEffect on the countered copy: %v", err)
		}
		if g.Stack.Size() != before+1 {
			t.Error("copying the countered copy from last-known information made nothing")
		}
	})
}

// TestABouncedSpellCopyCeasesToExist — "return target spell to its
// owner's hand" (Remand, Venser) on a copy: not a card, so no hand.
func TestABouncedSpellCopyCeasesToExist(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	_, cp := copyOfSpellForTest(t, g, me.ID)
	hand := me.Hand.Size()

	g.WithWriteLock(func() {
		if err := g.ReturnSpellToHandForEffect(cp); err != nil {
			t.Fatalf("ReturnSpellToHandForEffect: %v", err)
		}
	})

	assertCopyGone(t, g, cp)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d: a copy cannot be returned to a hand", hand, me.Hand.Size())
	}
	if countered, _ := copyEvents(g, cp); countered {
		t.Error("a bounce is not a counter")
	}
	// Announced as ceasing to exist: a move out of the stack to nowhere.
	found := false
	for _, ev := range g.Events {
		if ev.Kind == EventZoneMove && ev.CardID == cp && ev.OldZone == ZoneStack && ev.NewZone == "" {
			found = true
		}
	}
	if !found {
		t.Error("the bounced copy was not announced leaving the stack")
	}
}

// TestAnExiledSpellCopyCeasesToExist — #1318's "exile target spell"
// (Aven Interrupter) on a copy. It is not exiled, and the caller's
// continuation still runs and hears so: "it becomes plotted" has
// nothing to plot.
func TestAnExiledSpellCopyCeasesToExist(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	_, cp := copyOfSpellForTest(t, g, me.ID)

	told, exiled := false, true
	g.WithWriteLock(func() {
		if err := g.ExileSpellThenForEffect(cp, func(_ *Game, ok bool) error {
			told, exiled = true, ok
			return nil
		}); err != nil {
			t.Fatalf("ExileSpellThenForEffect: %v", err)
		}
	})

	assertCopyGone(t, g, cp)
	if !told {
		t.Fatal("the exile's continuation never ran — a caller sequencing through it would wait forever")
	}
	if exiled {
		t.Error("the continuation heard the copy reached exile")
	}
}

// TestASpellCopyMovingItselfWhileResolvingCeasesToExist — the resolving
// half of stackCopyLocked. A copy whose own effect moves it ("shuffle
// this spell into its owner's library") has already had its StackMeta
// entry taken by the resolver, so only the #920 resolving slot still
// knows it is a copy. It must not reach the library.
func TestASpellCopyMovingItselfWhileResolvingCeasesToExist(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	_, cp := copyOfSpellForTest(t, g, me.ID)
	library := me.Library.Size()

	g.WithWriteLock(func() {
		item := g.StackMeta[cp]
		var card Card
		for _, c := range g.Stack.Cards {
			if c.InstanceID == cp {
				card = c
			}
		}
		delete(g.StackMeta, cp)
		g.beginResolvingSpellLocked(item, card)
		if _, err := g.routeCardToZoneLocked(zoneRoute{CardID: cp, Dst: ZoneLibrary}); err != nil {
			t.Fatalf("routeCardToZoneLocked: %v", err)
		}
	})

	assertCopyGone(t, g, cp)
	if me.Library.Size() != library {
		t.Errorf("library %d → %d: the resolving copy was shuffled in as a card", library, me.Library.Size())
	}
}

// TestACounteredCopyOfACommanderSpellAsksNothing — the route skips the
// replacement window for a copy, not just the landing, and it does so
// UNCONDITIONALLY on IsCopy rather than by reading IsCommander.
//
// Since #1363, createSpellCopyLocked no longer carries IsCommander
// onto the copy at all (CR 903.3 — the designation is not a copiable
// value), so an ordinary copy of a commander spell no longer even
// LOOKS like a commander to the CR 903.9 built-in. That would make
// this test pass for the wrong reason — commanderZoneReplacement's own
// AppliesTo would already say no — so the flag is set BY HAND on the
// copy after it is made, standing in for whatever future path might
// otherwise re-carry it (a card-level "except" clause, a fixture built
// directly rather than through createSpellCopyLocked). The skip in
// routeCardToZoneLocked (stackCopyLocked) has to hold regardless: no
// replacement has a card-shaped object to act on, and a copy must
// never pause on a prompt about a zone it will not reach.
func TestACounteredCopyOfACommanderSpellAsksNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	orig := pushLKISpell(t, g, me.ID, "Commander Spell")
	var cp uuid.UUID
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(orig, me.ID, false, nil); err != nil {
			t.Fatalf("CopySpellForEffect: %v", err)
		}
		for id, item := range g.StackMeta {
			if item != nil && item.IsCopy {
				cp = id
			}
		}
		// Set by hand: see the doc comment above for why the copy
		// does not carry this naturally any more.
		for i := range g.Stack.Cards {
			if g.Stack.Cards[i].InstanceID == cp {
				g.Stack.Cards[i].IsCommander = true
			}
		}
	})
	counterForTest(t, g, cp)

	if n := len(g.PendingChoices); n != 0 {
		t.Errorf("%d prompts queued by countering a copy, want none: %+v", n, g.PendingChoices)
	}
	assertCopyGone(t, g, cp)
}

// TestASpellCopyOfACommanderSpellIsNotACommander pins #1363 itself:
// CR 903.3 says the commander designation is an attribute of the
// physical card, not a copiable characteristic (CR 707.2), so a copy
// is never a commander — whatever the spell it was copied from was.
// Before the fix, createSpellCopyLocked's `copyCard := src` carried
// IsCommander across, and the wire's `is_commander` on the stack card
// (plus any future "commander spell" check) would see the copy as one.
func TestASpellCopyOfACommanderSpellIsNotACommander(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	orig := pushLKISpell(t, g, me.ID, "Commander Spell")
	var cp uuid.UUID
	g.WithWriteLock(func() {
		for i := range g.Stack.Cards {
			if g.Stack.Cards[i].InstanceID == orig {
				g.Stack.Cards[i].IsCommander = true
			}
		}
		if err := g.CopySpellForEffect(orig, me.ID, false, nil); err != nil {
			t.Fatalf("CopySpellForEffect: %v", err)
		}
		for id, item := range g.StackMeta {
			if item != nil && item.IsCopy {
				cp = id
			}
		}
	})
	if cp == uuid.Nil {
		t.Fatal("no copy on the stack")
	}
	for _, c := range g.Stack.Cards {
		if c.InstanceID == cp && c.IsCommander {
			t.Error("spell copy carries IsCommander from the original it was copied from")
		}
	}

	// No commander tax: a copy is never CAST (CR 707.10), so nothing
	// increments CommanderCasts for it, and no commander-damage
	// attribution: the original is still on the stack as an instant/
	// sorcery, not a permanent, so there is nothing on the battlefield
	// for combatDamageTailLocked to read IsCommander off in the first
	// place. Asserted here so a future change that routed a copy
	// through the ordinary cast path, or gave it a battlefield
	// presence, would have to update this test to keep it green.
	if n := me.CommanderCasts[orig]; n != 0 {
		t.Errorf("CommanderCasts[orig] = %d after copying, want 0 (a copy is not a cast)", n)
	}
	if n := me.CommanderDamage[cp]; n != 0 {
		t.Errorf("CommanderDamage[cp] = %d, want 0 (a spell copy never deals commander damage)", n)
	}
}

// TestACounteredOriginalStillLands — the branch is the COPY's, not the
// stack's: the spell the copy was made from, countered next, still goes
// to its owner's graveyard as the card it is.
func TestACounteredOriginalStillLands(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	orig, cp := copyOfSpellForTest(t, g, me.ID)

	counterForTest(t, g, cp)
	counterForTest(t, g, orig)

	if z := zoneOfCard(g, orig); z != ZoneGraveyard {
		t.Errorf("the countered original is in %q, want the graveyard", z)
	}
	assertCopyGone(t, g, cp)
}
