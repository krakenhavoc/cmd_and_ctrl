package game

import "testing"

// exile_until_cast_test.go — #3213. The caller's hit test decides where
// the walk stops, the grant has no mana value cap, and the chooser is
// whoever the caller names.

func nonlandNotNamed(name string) func(Card) bool {
	return func(c Card) bool { return !c.IsLand() && c.Name != name }
}

// The walk skips lands and a card with the excluded name, stops on the
// first card the test accepts, and leaves the cards below it alone.
func TestExileUntilCastStopsOnTheCallersHit(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[1]
	hit := libCard(me.ID, "Divination", "Sorcery", "{2}{U}")
	same := libCard(me.ID, "Same Name", "Instant", "{U}")
	land := libCard(me.ID, "Mountain", "Basic Land — Mountain", "")
	deep := libCard(me.ID, "Never Reached", "Instant", "{U}")
	libraryOf(t, me, deep, hit, same, land)

	g.WithWriteLock(func() {
		if err := g.ExileUntilCastFreeForEffect(me.ID, g.Seats[0].ID, nonlandNotNamed("Same Name")); err != nil {
			t.Fatalf("ExileUntilCastFreeForEffect: %v", err)
		}
	})
	c := discoverPrompt(t, g)
	if c.MayCastCard != hit.InstanceID || c.Chooser != me.ID {
		t.Fatalf("prompt = card %s chooser %s, want the Divination for the walked player", c.MayCastCard, c.Chooser)
	}
	if c.MayCastKeyword != MayCastKeywordExileUntil {
		t.Errorf("keyword = %q", c.MayCastKeyword)
	}
	if !me.Library.Contains(deep.InstanceID) {
		t.Errorf("the walk went past its hit")
	}
}

// Accepting grants an uncapped, free, flash-timed cast to the chooser
// and bottoms the pile; the hit stays castable in exile.
func TestExileUntilCastAcceptGrantsAnUncappedFreeCast(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	hit := libCard(me.ID, "Emrakul", "Legendary Creature — Eldrazi", "{15}")
	land := libCard(me.ID, "Mountain", "Basic Land — Mountain", "")
	libraryOf(t, me, hit, land)

	g.WithWriteLock(func() {
		_ = g.ExileUntilCastFreeForEffect(me.ID, hit.InstanceID, nonlandNotNamed("x"))
	})
	answerMayCast(t, g, me.ID, true)
	perm := g.CastPermissionOnCardByIDForEffect(hit.InstanceID)
	if perm == nil || perm.Cost != "{0}" || perm.Timing != TimingFlash || !perm.CastOnly ||
		perm.MaxSpellManaValue != nil || perm.LapseOnPass != LapseToLibraryBottom {
		t.Fatalf("grant = %+v", perm)
	}
	if !me.Library.Contains(land.InstanceID) || me.Library.Contains(hit.InstanceID) {
		t.Errorf("the passed-over land should be on the bottom and the hit still in exile")
	}
}

// Declining puts the hit on the bottom with the rest. No hit at all
// puts everything back and asks nothing.
func TestExileUntilCastDeclineAndNoHitBottomEverything(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	hit := libCard(me.ID, "Divination", "Sorcery", "{2}{U}")
	land := libCard(me.ID, "Mountain", "Basic Land — Mountain", "")
	libraryOf(t, me, hit, land)
	g.WithWriteLock(func() {
		_ = g.ExileUntilCastFreeForEffect(me.ID, hit.InstanceID, nonlandNotNamed("x"))
	})
	answerMayCast(t, g, me.ID, false)
	if len(g.Exile.Cards) != 0 || me.Library.Size() != 2 {
		t.Errorf("declined: exile %d, library %d, want 0 and 2", len(g.Exile.Cards), me.Library.Size())
	}

	libraryOf(t, me, land)
	g.WithWriteLock(func() {
		if err := g.ExileUntilCastFreeForEffect(me.ID, hit.InstanceID, nonlandNotNamed("x")); err != nil {
			t.Fatalf("no-hit walk: %v", err)
		}
	})
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceMayCast {
			t.Fatalf("prompted with no hit")
		}
	}
	if len(g.Exile.Cards) != 0 || me.Library.Size() != 1 {
		t.Errorf("no hit: exile %d, library %d, want 0 and 1", len(g.Exile.Cards), me.Library.Size())
	}
}

// #4205: the keeping form leaves a declined hit in exile and bottoms the
// rest of the pile; the base form (cascade's) still bottoms the hit.
func TestExileUntilCastKeepingLeavesADeclinedHitInExile(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	hit := libCard(me.ID, "Divination", "Sorcery", "{2}{U}")
	land := libCard(me.ID, "Mountain", "Basic Land — Mountain", "")
	libraryOf(t, me, hit, land)
	g.WithWriteLock(func() {
		if err := g.ExileUntilCastFreeKeepingForEffect(me.ID, hit.InstanceID, nonlandNotNamed("x")); err != nil {
			t.Fatalf("ExileUntilCastFreeKeepingForEffect: %v", err)
		}
	})
	c := discoverPrompt(t, g)
	if c.DeclineLabel != "Leave it in exile" {
		t.Errorf("decline label = %q", c.DeclineLabel)
	}
	answerMayCast(t, g, me.ID, false)
	if !g.Exile.Contains(hit.InstanceID) || me.Library.Contains(hit.InstanceID) {
		t.Errorf("the declined hit should stay in exile")
	}
	if !me.Library.Contains(land.InstanceID) || g.Exile.Contains(land.InstanceID) {
		t.Errorf("the passed-over land should be on the bottom of the library")
	}
}

// Accepting and not casting before the pass also leaves the hit in
// exile (LapseStaysInExile); the grant is gone.
func TestExileUntilCastKeepingUncastHitStaysInExileOnThePass(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	hit := libCard(me.ID, "Counterspell", "Instant", "{U}{U}")
	land := libCard(me.ID, "Mountain", "Basic Land — Mountain", "")
	libraryOf(t, me, hit, land)
	g.WithWriteLock(func() {
		_ = g.ExileUntilCastFreeKeepingForEffect(me.ID, hit.InstanceID, nonlandNotNamed("x"))
	})
	answerMayCast(t, g, me.ID, true)
	perm := g.CastPermissionOnCardByIDForEffect(hit.InstanceID)
	if perm == nil || perm.LapseOnPass != LapseStaysInExile {
		t.Fatalf("grant = %+v", perm)
	}
	if g.Seats[g.Turn.PriorityHolder].ID != me.ID {
		t.Fatalf("test needs the caster to hold priority")
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if !g.Exile.Contains(hit.InstanceID) || me.Library.Contains(hit.InstanceID) {
		t.Errorf("an uncast hit should stay in exile after the pass")
	}
	if !me.Library.Contains(land.InstanceID) {
		t.Errorf("the passed-over land should be on the bottom")
	}
	if g.CastPermissionOnCardByIDForEffect(hit.InstanceID) != nil {
		t.Errorf("the grant should have closed on the pass")
	}
}
