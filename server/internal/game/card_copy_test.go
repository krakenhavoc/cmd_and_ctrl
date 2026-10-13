package game

import (
	"testing"

	"github.com/google/uuid"
)

// card_copy_test.go — CR 707.12, ADR 0066's 2026-10-13 card-copy
// amendment (#3258). Engine-side, with uncatalogued fixtures: nothing
// here depends on what the copied card DOES, only on the copy, who may
// cast it, what it becomes and when it stops existing. The catalog
// cards' own tests are Mizzix's Mastery's and Arcane Bombardment's.

// copyFixtureInGraveyard puts a plain card into p's graveyard.
func copyFixtureInGraveyard(p *Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID,
	})
	return id
}

func freeOffer(p *Player) CardCopyOffer {
	return CardCopyOffer{
		Player: p.ID, SourceName: "Test", Cost: "{0}",
		Timing: TimingFlash, Lapse: LapseStaysInExile,
	}
}

func cardCopiesAnywhere(g *Game) []Card {
	var out []Card
	scan := func(z *Zone) {
		for _, c := range z.Cards {
			if c.CardCopy {
				out = append(out, c)
			}
		}
	}
	scan(g.Exile)
	for _, p := range g.Seats {
		scan(p.Hand)
		scan(p.Graveyard)
	}
	return out
}

// A copy of a creature card is a NEW object with the original's
// copiable values, cast through the ordinary path; it resolves as a
// token (CR 707.10b), the original stays, and nothing is left behind.
func TestCardCopyOfAPermanentCardResolvesAsAToken(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	orig := copyFixtureInGraveyard(me, "Fixture Bear", "Creature — Bear", "{1}{G}")

	var copies []uuid.UUID
	g.WithWriteLock(func() {
		copies = g.CopyCardsAndOfferCastsForEffect(freeOffer(me), []uuid.UUID{orig})
	})
	if len(copies) != 1 || copies[0] == orig {
		t.Fatalf("copies = %v, want one new object", copies)
	}
	if !me.Graveyard.Contains(orig) || !me.Graveyard.Contains(copies[0]) {
		t.Fatal("a copy of a graveyard card is made in that graveyard, beside the original")
	}
	if err := g.CastSpell(me.ID, copies[0], CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("free cast of the copy: %v", err)
	}
	if item := g.StackMeta[copies[0]]; item == nil || !item.IsCopy {
		t.Fatal("the cast copy is a spell with IsCopy set")
	}
	resolveTop(t, g)

	var token *Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Name == "Fixture Bear" {
			token = c
		}
	}
	if token == nil || !token.IsToken() {
		t.Fatalf("battlefield = %+v, want a Fixture Bear TOKEN", g.Battlefield.Cards)
	}
	if !me.Graveyard.Contains(orig) {
		t.Error("the original card stays in the graveyard")
	}
	if n := len(cardCopiesAnywhere(g)); n != 0 {
		t.Errorf("%d copies left over", n)
	}
}

// CR 704.5e: a copy nobody is offered ceases to exist at the next state
// check — whatever closed the offer.
func TestCardCopyWithNoOfferCeasesAtTheNextStateCheck(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	orig := copyFixtureInGraveyard(me, "Fixture Spell", "Instant", "{R}")
	g.WithWriteLock(func() {
		g.CopyCardsAndOfferCastsForEffect(freeOffer(me), []uuid.UUID{orig})
	})
	if n := len(cardCopiesAnywhere(g)); n != 1 {
		t.Fatalf("%d copies, want 1", n)
	}
	g.WithWriteLock(func() {
		g.runStateChecksLocked()
	})
	if n := len(cardCopiesAnywhere(g)); n != 1 {
		t.Fatalf("%d copies after a state check with the offer live, want 1", n)
	}
	// The offer is withdrawn (the end-of-turn sweep, a pass): the copy goes.
	g.WithWriteLock(func() {
		me.CastPermissions = nil
		g.runStateChecksLocked()
	})
	if n := len(cardCopiesAnywhere(g)); n != 0 {
		t.Errorf("%d copies after the offer closed, want 0", n)
	}
	if !me.Graveyard.Contains(orig) {
		t.Error("the original stays")
	}
}

// The window closes at the end of the turn by the permission's own
// duration, and the copy goes with it.
func TestCardCopyDoesNotOutliveTheTurn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	orig := copyFixtureInGraveyard(me, "Fixture Spell", "Instant", "{R}")
	g.WithWriteLock(func() {
		g.CopyCardsAndOfferCastsForEffect(CardCopyOffer{Player: me.ID, Cost: "{0}", Timing: TimingFlash}, []uuid.UUID{orig})
		g.sweepCastPermissionsLocked(true)
		g.runStateChecksLocked()
	})
	if n := len(cardCopiesAnywhere(g)); n != 0 {
		t.Errorf("%d copies after the end of the turn, want 0", n)
	}
}

// Nothing to copy, nobody to copy for: no copy, no offer.
func TestCardCopyRefusesWhatIsNotThere(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	inHand := uuid.New()
	me.Hand.PushTop(Card{InstanceID: inHand, Name: "In Hand", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	gone := uuid.New()
	g.WithWriteLock(func() {
		if got := g.CopyCardsAndOfferCastsForEffect(freeOffer(me), []uuid.UUID{inHand, gone}); len(got) != 0 {
			t.Errorf("copies = %v, want none: a hand card and a missing card are not copied", got)
		}
		off := freeOffer(me)
		off.Player = uuid.Nil
		orig := copyFixtureInGraveyard(me, "Fixture Spell", "Instant", "{R}")
		if got := g.CopyCardsAndOfferCastsForEffect(off, []uuid.UUID{orig}); len(got) != 0 {
			t.Errorf("copies = %v, want none for no player", got)
		}
	})
}

// Several copies are offered at once with no total cap, each an
// independent grant: casting one leaves the others castable.
func TestCardCopiesAreOfferedTogetherAndIndependently(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var ids []uuid.UUID
	for _, n := range []string{"A", "B", "C"} {
		ids = append(ids, copyFixtureInGraveyard(me, n, "Instant", "{R}"))
	}
	var copies []uuid.UUID
	g.WithWriteLock(func() { copies = g.CopyCardsAndOfferCastsForEffect(freeOffer(me), ids) })
	if len(copies) != 3 {
		t.Fatalf("%d copies, want 3", len(copies))
	}
	if err := g.CastSpell(me.ID, copies[2], CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("cast the third copy first: %v", err)
	}
	if err := g.CastSpell(me.ID, copies[0], CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("cast the first copy second: %v", err)
	}
	if top := g.Stack.Cards[len(g.Stack.Cards)-1]; top.InstanceID != copies[0] {
		t.Errorf("the copy cast last is on top, got %q", top.Name)
	}
	if n := len(cardCopiesAnywhere(g)); n != 1 {
		t.Errorf("%d copies still waiting, want the one never cast", n)
	}
}

// A copy survives Clone (undo) and a snapshot round trip with its marker
// and its offer.
func TestCardCopySurvivesCloneAndSnapshot(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	orig := copyFixtureInGraveyard(me, "Fixture Spell", "Instant", "{R}")
	g.WithWriteLock(func() { g.CopyCardsAndOfferCastsForEffect(freeOffer(me), []uuid.UUID{orig}) })

	clone := g.Clone()
	if n := len(cardCopiesAnywhere(clone)); n != 1 {
		t.Errorf("clone holds %d copies, want 1", n)
	}
	_, restored := roundTrip(t, g)
	copies := cardCopiesAnywhere(restored)
	if len(copies) != 1 || !copies[0].CardCopy {
		t.Fatalf("restored copies = %+v, want 1 still marked", copies)
	}
	rme := restored.Seats[0]
	restored.ReadSnapshot(func() {
		if perm := restored.CastPermissionForLocked(rme.ID, copies[0], ZoneGraveyard); perm == nil || perm.Cost != "{0}" {
			t.Errorf("restored permission = %+v, want the free offer", perm)
		}
	})
	// Casting the restored copy and the original game's copy is independent.
	if err := restored.CastSpell(rme.ID, copies[0].InstanceID, CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("cast the restored copy: %v", err)
	}
}
