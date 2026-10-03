package game

import (
	"testing"

	"github.com/google/uuid"
)

// world_rule_test.go — ADR 0109 §8 (#1862): Card.EntryOrdinal and the
// world rule (CR 704.5k) that reads it. The catalog half (Caverns of
// Despair, Land's Edge and the rest) is in
// cards/effects/world_rule_cards_test.go.

func worldCard(owner uuid.UUID, name string) Card {
	return Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "World Enchantment",
		Owner: owner, Controller: owner,
	}
}

// worldPut puts the named hand cards onto the battlefield as ONE
// simultaneous entry and settles the state-based actions.
func worldPut(t *testing.T, g *Game, ids ...uuid.UUID) {
	t.Helper()
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, err := g.putOntoBattlefieldFromZoneLocked(ids, ZoneHand, ZoneEntryOptions{}); err != nil {
		t.Fatalf("put: %v", err)
	}
	g.runStateChecksLocked()
}

func wrPerm(g *Game, id uuid.UUID) *Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

// Two entries, one after the other: the second has a strictly larger
// ordinal. One simultaneous entry of two cards: they share one, and it
// is larger again.
func TestEntryOrdinalIsSharedBySimultaneousEntryAndGrowsAfter(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	a, b, c, d := ebLand(me.ID, "A"), ebLand(me.ID, "B"), ebLand(me.ID, "C"), ebLand(me.ID, "D")
	for _, card := range []Card{a, b, c, d} {
		me.Hand.PushTop(card)
	}
	worldPut(t, g, a.InstanceID)
	worldPut(t, g, b.InstanceID)
	worldPut(t, g, c.InstanceID, d.InstanceID)
	oa, ob := wrPerm(g, a.InstanceID).EntryOrdinal, wrPerm(g, b.InstanceID).EntryOrdinal
	oc, od := wrPerm(g, c.InstanceID).EntryOrdinal, wrPerm(g, d.InstanceID).EntryOrdinal
	if oa == 0 || ob <= oa {
		t.Fatalf("separate entries: ordinals %d then %d, want nonzero and increasing", oa, ob)
	}
	if oc != od {
		t.Errorf("one simultaneous entry: ordinals %d and %d, want one shared", oc, od)
	}
	if oc <= ob {
		t.Errorf("a later entry's ordinal %d is not above the earlier %d", oc, ob)
	}
	// A permanent that leaves carries no ordinal into its new zone.
	g.mu.Lock()
	err := g.sacrificePermanentLocked(a.InstanceID)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	for _, card := range me.Graveyard.Cards {
		if card.InstanceID == a.InstanceID && card.EntryOrdinal != 0 {
			t.Errorf("the card in the graveyard still has ordinal %d", card.EntryOrdinal)
		}
	}
}

// CR 704.5k: of two world permanents, the one that entered more
// recently stays and the other goes to its owner's graveyard.
func TestWorldRuleKeepsTheNewestWorldPermanent(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	old, newer := worldCard(me.ID, "Old World"), worldCard(opp.ID, "New World")
	me.Hand.PushTop(old)
	opp.Hand.PushTop(newer)
	worldPut(t, g, old.InstanceID)
	if wrPerm(g, old.InstanceID) == nil {
		t.Fatal("a lone world permanent was swept")
	}
	g.mu.Lock()
	if _, err := g.putOntoBattlefieldFromZoneLocked([]uuid.UUID{newer.InstanceID}, ZoneHand, ZoneEntryOptions{}); err != nil {
		t.Fatalf("put: %v", err)
	}
	g.runStateChecksLocked()
	g.mu.Unlock()
	if wrPerm(g, newer.InstanceID) == nil {
		t.Error("the newest world permanent left")
	}
	if wrPerm(g, old.InstanceID) != nil {
		t.Fatal("the older world permanent stayed")
	}
	if !me.Graveyard.Contains(old.InstanceID) {
		t.Error("the older world permanent is not in its OWNER's graveyard")
	}
}

// CR 704.5k: "In the event of a tie for the shortest amount of time,
// all are put into their owners' graveyards" — two world permanents one
// simultaneous entry put there.
func TestWorldRuleTieSweepsThemAll(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	a, b := worldCard(me.ID, "World A"), worldCard(me.ID, "World B")
	me.Hand.PushTop(a)
	me.Hand.PushTop(b)
	worldPut(t, g, a.InstanceID, b.InstanceID)
	if wrPerm(g, a.InstanceID) != nil || wrPerm(g, b.InstanceID) != nil {
		t.Fatal("a tie for the most recent entry left a world permanent on the battlefield")
	}
	if !me.Graveyard.Contains(a.InstanceID) || !me.Graveyard.Contains(b.InstanceID) {
		t.Error("both world permanents should be in the graveyard")
	}
}

// A tie at the newest ordinal sweeps them all, but an older third world
// permanent goes with them; and a world permanent beside non-world ones
// is untouched.
func TestWorldRuleOnlyCountsWorldPermanents(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	w := worldCard(me.ID, "Lone World")
	land := ebLand(me.ID, "Plain Land")
	me.Hand.PushTop(w)
	me.Hand.PushTop(land)
	worldPut(t, g, w.InstanceID)
	worldPut(t, g, land.InstanceID)
	if wrPerm(g, w.InstanceID) == nil || wrPerm(g, land.InstanceID) == nil {
		t.Fatal("one world permanent and a land: nothing should move")
	}
}

// A face-down permanent has no supertypes (CR 708.2), so it is never a
// world permanent.
func TestWorldRuleIgnoresAFaceDownPermanent(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	shown, hidden := worldCard(me.ID, "Shown World"), worldCard(me.ID, "Hidden World")
	me.Hand.PushTop(shown)
	me.Hand.PushTop(hidden)
	worldPut(t, g, hidden.InstanceID)
	g.mu.Lock()
	wrPerm(g, hidden.InstanceID).FaceDown = true
	wrPerm(g, hidden.InstanceID).FaceDownKind = FaceDownMorphed
	g.layerVersion.Add(1)
	g.mu.Unlock()
	worldPut(t, g, shown.InstanceID)
	g.mu.Lock()
	g.runStateChecksLocked()
	g.mu.Unlock()
	if wrPerm(g, hidden.InstanceID) == nil {
		t.Error("a face-down permanent was swept by the world rule")
	}
	if wrPerm(g, shown.InstanceID) == nil {
		t.Error("the face-up world permanent was swept beside a face-down one")
	}
}

// A file written before the field existed carries no ordinals: it
// round-trips exactly, the counter resumes past what a restored file
// does carry, and the world rule orders unstamped world permanents by
// their entry stamps — the newer one stays.
func TestAnOlderFileKeepsItsWorldPermanentsInEntryOrder(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	old, newer := worldCard(me.ID, "Old"), worldCard(me.ID, "Newer")
	old.EnteredBattlefieldAt, newer.EnteredBattlefieldAt = 100, 300
	g.Battlefield.PushTop(old)
	g.Battlefield.PushTop(newer)
	restored, err := throughJSON(t, g.CaptureSnapshot()).RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	if wrPerm(restored, old.InstanceID).EntryOrdinal != 0 {
		t.Fatal("restore rewrote an older file's permanents")
	}
	restored.mu.Lock()
	restored.runStateChecksLocked()
	restored.mu.Unlock()
	if wrPerm(restored, newer.InstanceID) == nil || wrPerm(restored, old.InstanceID) != nil {
		t.Error("the world rule did not keep the later entrant of an older file")
	}

	// A stamped ordinal round-trips, and the counter resumes past it.
	land := ebLand(me.ID, "Land")
	me.Hand.PushTop(land)
	worldPut(t, g, land.InstanceID)
	o := wrPerm(g, land.InstanceID).EntryOrdinal
	again, err := throughJSON(t, g.CaptureSnapshot()).RestoreStrict()
	if err != nil {
		t.Fatalf("second RestoreStrict: %v", err)
	}
	if got := wrPerm(again, land.InstanceID).EntryOrdinal; got != o || again.entryOrdinalSeq != o {
		t.Errorf("ordinal %d, counter %d after restore; want both %d", got, again.entryOrdinalSeq, o)
	}
}
