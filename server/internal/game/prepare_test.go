package game

import (
	"testing"

	"github.com/google/uuid"
)

// prepare_test.go — CR 722, ADR 0090 (#1328). Engine-side, with an
// uncatalogued fixture: nothing here depends on what a prepare spell
// DOES, only on the designation, the copy, who may cast it, and when
// it stops existing. The catalog cards' own tests live in
// cards/effects (skycoach_conductor_test.go and its siblings).

// preparationFixture is a preparation card as the deck importer builds
// one: two faces, the creature first, the inset spell second, with the
// front face's printed keywords at card level (deck.printedKeywords).
// {0} on the spell so the mana gate is not the thing under test.
func preparationFixture(owner uuid.UUID) Card {
	c := Card{
		InstanceID: uuid.New(),
		OracleID:   "prepare-fixture-oracle",
		Layout:     LayoutPrepare,
		Owner:      owner,
		Controller: owner,
		Keywords:   []string{"flash", "flying"},
		Faces: []Face{
			{Name: "Fixture Conductor", TypeLine: "Creature — Bird Pilot", ManaCost: "{2}{U}", Colors: []string{"U"}, Power: 2, Toughness: 3},
			{Name: "Fixture Idea", TypeLine: "Sorcery", ManaCost: "{0}", Colors: []string{"U"}},
		},
	}
	c.SetFace(0)
	return c
}

// prepareCopiesIn lists the prepare copies sitting in a zone.
func prepareCopiesIn(z *Zone) []Card {
	var out []Card
	for _, c := range z.Cards {
		if c.PrepareCopy {
			out = append(out, c)
		}
	}
	return out
}

// zoneCard reads one card out of a zone by instance ID.
func prepZoneCard(z *Zone, id uuid.UUID) (Card, bool) {
	for _, c := range z.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return Card{}, false
}

// zoneOfCard names the zone a card is in, or "" when it is in none —
// the state of an object that has ceased to exist.
func zoneOfCard(g *Game, id uuid.UUID) ZoneKind {
	var out ZoneKind
	g.ReadSnapshot(func() {
		if z := g.findCardZoneLocked(id); z != nil {
			out = z.Kind
		}
	})
	return out
}

// preparedOnBattlefield seats the fixture on seat 0's battlefield and
// prepares it. Returns the permanent's ID and the copy's.
func preparedOnBattlefield(t *testing.T, g *Game) (permID, copyID uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	permID = pushTypedTestCard(g, preparationFixture(me.ID))
	g.WithWriteLock(func() {
		ok, err := g.BecomePreparedForEffect(permID)
		if err != nil || !ok {
			t.Fatalf("BecomePreparedForEffect = %v, %v; want true", ok, err)
		}
	})
	copies := prepareCopiesIn(g.Exile)
	if len(copies) != 1 {
		t.Fatalf("exile holds %d prepare copies, want 1", len(copies))
	}
	return permID, copies[0].InstanceID
}

// CR 722.3a / 722.3c: becoming prepared sets the designation and puts
// a copy of the prepare spell in exile — with ONLY the prepare spell's
// characteristics. The front face's printed keywords must not ride
// along: a flash creature's copy of its sorcery would otherwise be
// castable at instant speed.
func TestBecomingPreparedMakesACopyOfThePrepareSpellInExile(t *testing.T) {
	g := newActiveGame(t)
	permID, copyID := preparedOnBattlefield(t, g)

	if !g.IsPreparedForEffect(permID) {
		t.Fatal("the permanent is not prepared")
	}
	cp, _ := prepZoneCard(g.Exile, copyID)
	if cp.Name != "Fixture Idea" || cp.TypeLine != "Sorcery" || cp.ManaCost != "{0}" {
		t.Errorf("copy = %q / %q / %q, want the prepare spell Fixture Idea / Sorcery / {0}", cp.Name, cp.TypeLine, cp.ManaCost)
	}
	if cp.ActiveFace != 1 || CatalogKey(cp) != "prepare-fixture-oracle#1" {
		t.Errorf("copy face %d key %q, want face 1 and the prepare spell's catalog key", cp.ActiveFace, CatalogKey(cp))
	}
	if len(cp.Keywords) != 0 {
		t.Errorf("copy keywords = %v, want none — the creature's keywords are not the prepare spell's", cp.Keywords)
	}
	if cp.Owner != g.Seats[0].ID || cp.Controller != g.Seats[0].ID {
		t.Error("the copy is created by, and belongs to, the permanent's controller")
	}
}

// CR 722.3a's two refusals: a permanent with no prepare spell cannot
// become prepared, and a prepared permanent cannot become prepared
// again — so no second copy.
func TestBecomingPreparedRefusesWhatCR7223aRefuses(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushTypedTestCard(g, Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	permID, _ := preparedOnBattlefield(t, g)

	g.WithWriteLock(func() {
		if ok, err := g.BecomePreparedForEffect(bear); ok || err != nil {
			t.Errorf("a creature with no prepare spell became prepared (%v, %v)", ok, err)
		}
		if ok, err := g.BecomePreparedForEffect(permID); ok || err != nil {
			t.Errorf("an already-prepared permanent became prepared again (%v, %v)", ok, err)
		}
	})
	if n := len(prepareCopiesIn(g.Exile)); n != 1 {
		t.Errorf("exile holds %d prepare copies, want exactly 1", n)
	}
	if g.IsPreparedForEffect(bear) {
		t.Error("the bear is prepared")
	}
}

// CR 722.3c: "the prepared permanent's controller may cast the copy",
// and nobody else. The permission is derived, so every reader — the
// cast path, the enumerator's fast negative, the view — sees it.
func TestOnlyThePreparedPermanentsControllerMayCastTheCopy(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_, copyID := preparedOnBattlefield(t, g)
	cp, _ := prepZoneCard(g.Exile, copyID)

	g.ReadSnapshot(func() {
		perm := g.CastPermissionForLocked(me.ID, cp, ZoneExile)
		if perm == nil {
			t.Fatal("the controller has no permission over the copy")
		}
		if face, ok := perm.NamedFace(); !ok || face != 1 {
			t.Errorf("permission faces = %v, want just the prepare spell [1]", perm.Faces)
		}
		if perm.Cost != "" {
			t.Errorf("permission cost = %q, want the printed cost (CR 722.3c grants a cast, not a discount)", perm.Cost)
		}
		if g.CastPermissionForLocked(opp.ID, cp, ZoneExile) != nil {
			t.Error("an opponent may cast the copy")
		}
		if !g.AnyCastPermissionsForEffect() {
			t.Error("AnyCastPermissionsForEffect is false with a prepared permanent on the battlefield")
		}
		if on := g.CastPermissionOnCardForEffect(cp, ZoneExile); on == nil || on.Player != me.ID {
			t.Error("the view's permission lookup does not name the controller")
		}
	})
}

// CR 722.3c / 601.2i / 707.10: casting the copy puts a COPY on the
// stack, unprepares the permanent as the spell becomes cast, and the
// copy ceases to exist when it resolves rather than landing in a
// graveyard.
func TestCastingThePrepareCopyUnpreparesAndTheCopyCeasesToExist(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	permID, copyID := preparedOnBattlefield(t, g)

	if err := g.CastSpell(me.ID, copyID, CastSpellParams{FromZone: string(ZoneExile)}); err != nil {
		t.Fatalf("cast the prepare copy: %v", err)
	}
	if g.IsPreparedForEffect(permID) {
		t.Error("CR 601.2i: the permanent is still prepared after its copy was cast")
	}
	item := g.StackMeta[copyID]
	if item == nil || !item.IsCopy {
		t.Fatal("the cast copy is not a copy on the stack")
	}
	if item.CastFromZone != ZoneExile {
		t.Errorf("CastFromZone = %q, want exile", item.CastFromZone)
	}
	onStack, _ := prepZoneCard(g.Stack, copyID)
	if onStack.Name != "Fixture Idea" {
		t.Errorf("the spell on the stack is %q, want the prepare spell", onStack.Name)
	}
	for i := 0; i < 8 && g.Stack.Contains(copyID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if z := zoneOfCard(g, copyID); z != "" {
		t.Errorf("the resolved copy is in %q, want nowhere — a copy is not a card (CR 707.10)", z)
	}
	if n := len(prepareCopiesIn(me.Graveyard)); n != 0 {
		t.Error("a prepare copy reached the graveyard")
	}
}

// CR 722.3c's "for as long as" ends when the permanent leaves: the copy
// is uncastable at once (the permission is derived from a live
// permanent) and gone at the next state check (CR 704.5e).
func TestThePrepareCopyCeasesWhenThePermanentLeaves(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	permID, copyID := preparedOnBattlefield(t, g)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(permID); err != nil {
			t.Fatalf("destroy: %v", err)
		}
		cp, ok := prepZoneCard(g.Exile, copyID)
		if ok && g.CastPermissionForLocked(me.ID, cp, ZoneExile) != nil {
			t.Error("the copy is still castable with its permanent in the graveyard")
		}
		g.runStateChecksLocked()
	})
	if g.Exile.Contains(copyID) {
		t.Error("CR 704.5e: the copy is still in exile after its permanent died")
	}
	dead, ok := prepZoneCard(me.Graveyard, permID)
	if !ok || dead.Prepared {
		t.Error("CR 400.7: the card in the graveyard is still prepared")
	}
}

// CR 722.3b: an effect that unprepares the permanent takes the copy
// with it, immediately.
func TestUnpreparingRemovesTheCopy(t *testing.T) {
	g := newActiveGame(t)
	permID, copyID := preparedOnBattlefield(t, g)
	g.WithWriteLock(func() {
		if err := g.UnprepareForEffect(permID); err != nil {
			t.Fatalf("UnprepareForEffect: %v", err)
		}
	})
	if g.IsPreparedForEffect(permID) || g.Exile.Contains(copyID) {
		t.Error("the permanent is still prepared or its copy is still in exile")
	}
}

// A countered copy is still not a card: it goes to the graveyard for
// the moment CR 608.2b / 701.6a puts it there, wearing its prepare
// spell (not the creature face, CR 722.3c), and CR 704.5e removes it.
func TestACounteredPrepareCopyCeasesToExist(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	_, copyID := preparedOnBattlefield(t, g)
	if err := g.CastSpell(me.ID, copyID, CastSpellParams{FromZone: string(ZoneExile)}); err != nil {
		t.Fatalf("cast the prepare copy: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.counterSpellLocked(copyID, nil); err != nil {
			t.Fatalf("counter: %v", err)
		}
		if c, ok := prepZoneCard(me.Graveyard, copyID); ok && c.Name != "Fixture Idea" {
			t.Errorf("the countered copy landed as %q, want its prepare spell", c.Name)
		}
		g.runStateChecksLocked()
	})
	if z := zoneOfCard(g, copyID); z != "" {
		t.Errorf("the countered copy is in %q, want nowhere", z)
	}
}

// CR 722.3c: "…or phases in prepared". The designation survives
// phasing (CR 702.26d); the copy does not survive the time out, and a
// fresh one is made as the permanent phases back in.
func TestAPreparedPermanentPhasesInPreparedWithAFreshCopy(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	permID, oldCopy := preparedOnBattlefield(t, g)

	g.WithWriteLock(func() {
		g.phaseOutLocked(uuid.Nil, []uuid.UUID{permID}, phaseOutOptions{})
		g.runStateChecksLocked()
	})
	if g.Exile.Contains(oldCopy) {
		t.Fatal("the copy stayed in exile while its permanent was phased out")
	}
	g.WithWriteLock(func() { g.phaseInLocked([]uuid.UUID{permID}) })
	if !g.IsPreparedForEffect(permID) {
		t.Fatal("the permanent phased in unprepared")
	}
	copies := prepareCopiesIn(g.Exile)
	if len(copies) != 1 || copies[0].InstanceID == oldCopy {
		t.Fatalf("exile holds %d copies after phasing in, want one fresh copy", len(copies))
	}
	g.ReadSnapshot(func() {
		if g.CastPermissionForLocked(me.ID, copies[0], ZoneExile) == nil {
			t.Error("the fresh copy is not castable")
		}
	})
}

// The designation, the copy's marker and its link all survive a
// snapshot, and the restored copy is still castable by the controller.
func TestPreparedStateSurvivesASnapshot(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	permID, copyID := preparedOnBattlefield(t, g)

	_, restored := roundTrip(t, g)
	if !restored.IsPreparedForEffect(permID) {
		t.Fatal("the restored permanent is not prepared")
	}
	cp, ok := prepZoneCard(restored.Exile, copyID)
	if !ok || !cp.PrepareCopy || cp.ActiveFace != 1 {
		t.Fatal("the restored copy lost its marker or its face")
	}
	restored.ReadSnapshot(func() {
		if restored.CastPermissionForLocked(me.ID, cp, ZoneExile) == nil {
			t.Error("the restored copy is not castable")
		}
	})
}

// MoveCard drops a copy's link on EVERY move, so a copy that leaves
// exile and comes back by some route names nothing. The route that
// can reach it: the copy is cast, the permanent is prepared again in
// response (a new copy), and the first copy is countered INTO EXILE.
// With the link kept it would sit in exile naming a permanent that is
// once more that object and prepared — a second castable copy.
func TestACopyCounteredIntoExileIsNotCastableAgain(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	permID, first := preparedOnBattlefield(t, g)
	if err := g.CastSpell(me.ID, first, CastSpellParams{FromZone: string(ZoneExile)}); err != nil {
		t.Fatalf("cast the prepare copy: %v", err)
	}
	g.WithWriteLock(func() {
		if ok, err := g.BecomePreparedForEffect(permID); !ok || err != nil {
			t.Fatalf("re-prepare in response = %v, %v", ok, err)
		}
		if err := g.counterSpellLocked(first, &ZoneRef{Kind: ZoneExile}); err != nil {
			t.Fatalf("counter into exile: %v", err)
		}
		g.runStateChecksLocked()
	})
	if g.Exile.Contains(first) {
		t.Error("the countered copy sits in exile — it names nothing and must cease to exist")
	}
	if n := len(prepareCopiesIn(g.Exile)); n != 1 {
		t.Errorf("exile holds %d prepare copies, want exactly the fresh one", n)
	}
}
