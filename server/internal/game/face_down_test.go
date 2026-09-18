package game

import (
	"testing"

	"github.com/google/uuid"
)

// face_down_test.go — ADR 0069 / #656: the face-down OBJECT model.
//
// One section per decision:
//
//	1. the state, and why "face down with no kind" cannot be built
//	2. who may look, per kind and per zone
//	3. the CR 708.2 body reaches every reader the engine has
//	4. the catalog goes silent
//	5. the reveals — CR 708.9 on leaving, CR 702.143f on leaving the game
//	7. ManifestForEffect makes a real one
//	8. undo and the snapshot carry the kind
//
// The other half of decision 5 — the one CR 400.7 reset in MoveCard —
// is #697 and lives next door in face_down_lifecycle_test.go.

// faceDownGame is a two-seat active game with one named card on top of
// seat 0's library, returned by instance ID.
func faceDownGame(t *testing.T, name, typeLine, oracle string) (*Game, uuid.UUID) {
	t.Helper()
	g := newActiveGame(t)
	me := g.Seats[0]
	c := NewCard(name, me.ID)
	c.TypeLine = typeLine
	c.OracleID = oracle
	c.Power, c.Toughness = 7, 7
	c.ManaCost = "{4}{B}{B}"
	c.Colors = []string{"B"}
	c.Keywords = []string{"flying", "deathtouch"}
	id := c.InstanceID
	g.WithWriteLock(func() { me.Library.PushTop(c) })
	return g, id
}

// cardAnywhere finds a card by ID in any zone, for assertions that do
// not care where the engine put it.
func cardAnywhere(g *Game, id uuid.UUID) (Card, ZoneKind, bool) {
	var (
		out  Card
		kind ZoneKind
		ok   bool
	)
	g.ReadSnapshot(func() {
		if z := g.findCardZoneLocked(id); z != nil {
			if c, found := g.cardInZoneLocked(z, id); found {
				out, kind, ok = c, z.Kind, true
			}
		}
	})
	return out, kind, ok
}

// --- decision 2: who may look ---------------------------------------

// TestFaceDownExileViewersFollowTheKind: CR 406.3 gives a plain
// face-down exile no viewers at all — the player who exiled it
// included — and CR 702.143d gives a foretold card's owner a look. One
// table, one route, two answers.
func TestFaceDownExileViewersFollowTheKind(t *testing.T) {
	for _, tc := range []struct {
		name     string
		kind     FaceDownKind
		wantSelf bool
	}{
		{"CR 406.3 plain exile: nobody, not even the exiler", FaceDownExiled, false},
		{"CR 702.143d foretold: the owner may look", FaceDownForetold, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, id := faceDownGame(t, "Griselbrand", "Legendary Creature — Demon", "oracle-grissy")
			me, opp := g.Seats[0], g.Seats[1]
			g.WithWriteLock(func() {
				if _, err := g.routeCardToZoneLocked(zoneRoute{
					CardID: id, Dst: ZoneExile, Actor: me.ID, FaceDown: tc.kind,
				}); err != nil {
					t.Fatalf("route: %v", err)
				}
			})
			c, zone, ok := cardAnywhere(g, id)
			if !ok || zone != ZoneExile {
				t.Fatalf("card is in %q, want exile", zone)
			}
			if !c.FaceDown || c.FaceDownKind != tc.kind {
				t.Fatalf("face-down state = (%v, %q), want (true, %q)", c.FaceDown, c.FaceDownKind, tc.kind)
			}
			if got := c.IsKnownTo(me.ID); got != tc.wantSelf {
				t.Errorf("owner may look = %v, want %v", got, tc.wantSelf)
			}
			if c.IsKnownTo(opp.ID) {
				t.Error("an opponent may look at a face-down exiled card; no kind allows that")
			}
		})
	}
}

// TestManifestIsKnownToItsControllerAlone is CR 708.5: the controller
// of a face-down permanent may look at it. The battlefield is a PUBLIC
// zone, so the ordinary entry marks every seat a knower — this is the
// assertion that the face-down entry does not.
func TestManifestIsKnownToItsControllerAlone(t *testing.T) {
	g, id := faceDownGame(t, "Griselbrand", "Legendary Creature — Demon", "oracle-grissy")
	me, opp := g.Seats[0], g.Seats[1]
	var got uuid.UUID
	g.WithWriteLock(func() {
		var err error
		got, err = g.ManifestForEffect(me.ID)
		if err != nil {
			t.Fatalf("manifest: %v", err)
		}
	})
	if got != id {
		t.Fatalf("manifested %s, want the top of the library %s", got, id)
	}
	c, zone, ok := cardAnywhere(g, id)
	if !ok || zone != ZoneBattlefield {
		t.Fatalf("card is in %q, want the battlefield", zone)
	}
	if !c.FaceDown || c.FaceDownKind != FaceDownManifested {
		t.Fatalf("face-down state = (%v, %q), want (true, manifested)", c.FaceDown, c.FaceDownKind)
	}
	if !c.IsKnownTo(me.ID) {
		t.Error("the controller may not look at their own face-down permanent (CR 708.5)")
	}
	if c.IsKnownTo(opp.ID) {
		t.Error("an opponent knows a face-down permanent; the public zone marked them")
	}
}

// TestManifestAcceptsANonpermanentCard is CR 701.40a: manifest puts ANY
// card onto the battlefield face down. The object is a creature; the
// card underneath need not be a permanent card at all.
func TestManifestAcceptsANonpermanentCard(t *testing.T) {
	g, id := faceDownGame(t, "Ancestral Recall", "Instant", "oracle-recall")
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if _, err := g.ManifestForEffect(me.ID); err != nil {
			t.Fatalf("manifest an instant: %v", err)
		}
	})
	c, zone, ok := cardAnywhere(g, id)
	if !ok || zone != ZoneBattlefield {
		t.Fatalf("card is in %q, want the battlefield", zone)
	}
	if !c.IsCreature() {
		t.Error("a manifested instant is not a creature; CR 708.2 says the object is a 2/2 creature")
	}
}

// --- decision 3: the CR 708.2 body ----------------------------------

// TestFaceDownPermanentIsAColourlessVanillaTwoTwo walks every reader
// the engine has for "what is this object": the layered characteristic,
// the type predicates, the P/T the combat and SBA paths use, colour,
// and keywords.
func TestFaceDownPermanentIsAColourlessVanillaTwoTwo(t *testing.T) {
	g, id := faceDownGame(t, "Griselbrand", "Legendary Creature — Demon", "oracle-grissy")
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if _, err := g.ManifestForEffect(me.ID); err != nil {
			t.Fatalf("manifest: %v", err)
		}
		g.recomputeLayersLocked()
	})
	c, _, ok := cardAnywhere(g, id)
	if !ok {
		t.Fatal("manifested card vanished")
	}

	eff := c.Effective()
	if eff.Name != "" {
		t.Errorf("Effective().Name = %q, want empty (CR 708.2: no name)", eff.Name)
	}
	if eff.Power != 2 || eff.Toughness != 2 {
		t.Errorf("Effective() P/T = %d/%d, want 2/2", eff.Power, eff.Toughness)
	}
	if c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("combat P/T = %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
	if len(eff.Subtypes) != 0 || len(eff.Supertypes) != 0 {
		t.Errorf("Effective() types = super %v / sub %v, want neither", eff.Supertypes, eff.Subtypes)
	}
	if len(eff.Abilities) != 0 {
		t.Errorf("Effective().Abilities = %v, want none (CR 708.2a: no text)", eff.Abilities)
	}

	// The predicates targeting, combat and the catalog all run.
	if !c.IsCreature() {
		t.Error("a face-down permanent is not a creature")
	}
	if c.HasSubtype("Demon") {
		t.Error("a face-down permanent kept a printed subtype")
	}
	if c.HasSupertype("Legendary") {
		t.Error("a face-down permanent is still legendary; two morphs of one legend would fight")
	}
	if !c.IsColorless() {
		t.Errorf("EffectiveColors() = %v, want colourless", c.EffectiveColors())
	}
	if HasKeyword(&c, "flying") || HasKeyword(&c, "deathtouch") {
		t.Error("a face-down permanent kept its printed keywords")
	}
	// The printed surface is the copiable value (CR 707.2) and stays
	// the real card — the deliberate other half of the split.
	if !c.PrintedIsCreature() || c.TypeLine != "Legendary Creature — Demon" {
		t.Errorf("printed type line = %q, want the real card's", c.TypeLine)
	}
}

// TestFaceDownExiledCardKeepsItsCharacteristics is the other side of
// decision 3: CR 708.2 is about PERMANENTS. A card face down in exile
// is not one, gets no 2/2 body, and keeps the characteristics its
// owner may be entitled to read.
func TestFaceDownExiledCardKeepsItsCharacteristics(t *testing.T) {
	g, id := faceDownGame(t, "Griselbrand", "Legendary Creature — Demon", "oracle-grissy")
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if _, err := g.routeCardToZoneLocked(zoneRoute{
			CardID: id, Dst: ZoneExile, Actor: me.ID, FaceDown: FaceDownForetold,
		}); err != nil {
			t.Fatalf("route: %v", err)
		}
	})
	c, _, ok := cardAnywhere(g, id)
	if !ok {
		t.Fatal("exiled card vanished")
	}
	if c.FaceDownIsPermanent() {
		t.Fatal("a face-down EXILED card reads as a CR 708.2 permanent")
	}
	if eff := c.Effective(); eff.Name != "Griselbrand" || eff.Power != 7 {
		t.Errorf("Effective() = %q %d/%d, want the real card", eff.Name, eff.Power, eff.Toughness)
	}
	if CatalogKey(c) != "oracle-grissy" {
		t.Errorf("CatalogKey = %q, want the real key — the cast out of exile needs it (#658)", CatalogKey(c))
	}
}

// --- decision 4: catalog suppression --------------------------------

// TestFaceDownPermanentHasNoCatalogEntry is CR 708.2a through the one
// predicate: the key goes empty, so every Catalog* reader — triggers,
// statics, activated, mana, replacements, cost modifiers, printed
// keywords, the ETB hook — answers "no entry".
func TestFaceDownPermanentHasNoCatalogEntry(t *testing.T) {
	const oracle = "oracle-loud-card"
	prevLookup := CatalogLookup
	CatalogLookup = func(key string) *CardDef {
		if key != oracle {
			return nil
		}
		return &CardDef{
			PrintedKeywords: []string{"flying"},
			Triggered:       []TriggeredAbility{{}},
			Static:          []StaticAbility{{}},
			Activated:       []ActivatedAbilityShape{{}},
			Replacements:    []ReplacementEffect{{}},
			ManaAbilities:   []ManaAbilityShape{{}},
		}
	}
	t.Cleanup(func() { CatalogLookup = prevLookup })

	g, id := faceDownGame(t, "Loud Card", "Creature — Demon", oracle)
	me := g.Seats[0]

	// Face up in the library it is the loud card it prints.
	var up Card
	g.ReadSnapshot(func() { up, _ = g.cardInZoneLocked(me.Library, id) })
	if CatalogKey(up) != oracle || len(CatalogTriggers(CatalogKey(up))) == 0 {
		t.Fatal("fixture: the card has no catalog entry face up, so suppression proves nothing")
	}

	g.WithWriteLock(func() {
		if _, err := g.ManifestForEffect(me.ID); err != nil {
			t.Fatalf("manifest: %v", err)
		}
		g.recomputeLayersLocked()
	})
	down, _, _ := cardAnywhere(g, id)
	if key := CatalogKey(down); key != "" {
		t.Fatalf("CatalogKey on a face-down permanent = %q, want empty (CR 708.2a)", key)
	}
	if key := CatalogAbilityKey(down); key != "" {
		t.Errorf("CatalogAbilityKey = %q, want empty", key)
	}
	for name, n := range map[string]int{
		"triggered":     len(CatalogTriggers(CatalogKey(down))),
		"static":        len(CatalogStaticAbilities(CatalogKey(down))),
		"activated":     len(ActivatedAbilitiesForCard(down)),
		"replacements":  len(CatalogReplacements(CatalogAbilityKey(down))),
		"mana":          len(CatalogManaAbilities(CatalogAbilityKey(down))),
		"printed words": len(CatalogPrintedKeywords(CatalogKey(down))),
	} {
		if n != 0 {
			t.Errorf("a face-down permanent still has %d %s abilities", n, name)
		}
	}
	if Unimplemented(down) {
		t.Error("a face-down permanent is flagged unimplemented; it has no printed rules to fail at")
	}
	if HasKeyword(&down, "flying") {
		t.Error("a face-down permanent has flying from the catalog's printed-keyword slot")
	}
}

// --- decision 5: reveals (CR 708.9, CR 702.143f) --------------------

// TestFaceDownPermanentIsRevealedWhenItLeaves is CR 708.9. The card is
// revealed on the way out, so every seat knows it wherever it landed —
// a hidden zone included, which is what "its owner reveals it" means.
func TestFaceDownPermanentIsRevealedWhenItLeaves(t *testing.T) {
	g, id := faceDownGame(t, "Griselbrand", "Legendary Creature — Demon", "oracle-grissy")
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		if _, err := g.ManifestForEffect(me.ID); err != nil {
			t.Fatalf("manifest: %v", err)
		}
	})
	if c, _, _ := cardAnywhere(g, id); c.IsKnownTo(opp.ID) {
		t.Fatal("fixture: the opponent already knows the card")
	}
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(id); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	c, zone, ok := cardAnywhere(g, id)
	if !ok || zone != ZoneHand {
		t.Fatalf("card is in %q, want a hand", zone)
	}
	if !c.IsKnownTo(opp.ID) {
		t.Error("CR 708.9: a face-down permanent that left the battlefield was not revealed")
	}
	var revealed bool
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventRevealCards && ev.CardID == id {
				revealed = true
			}
		}
	})
	if !revealed {
		t.Error("no reveal event for the CR 708.9 reveal")
	}
}

// TestARevealedFaceDownPermanentIsStillLostInALibrary pins the ORDER
// of the two rules, which is the only place they disagree. CR 708.9
// reveals the card — every player SAW it — and then the destination
// decides what they still KNOW: a library is a hidden zone (CR 401.2),
// so nobody can find it afterwards. Reveal last would leave every seat
// able to read a library card by its position.
func TestARevealedFaceDownPermanentIsStillLostInALibrary(t *testing.T) {
	g, id := faceDownGame(t, "Griselbrand", "Legendary Creature — Demon", "oracle-grissy")
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		if _, err := g.ManifestForEffect(me.ID); err != nil {
			t.Fatalf("manifest: %v", err)
		}
	})
	if err := g.MoveCardByID(
		ZoneRef{Kind: ZoneBattlefield},
		ZoneRef{Kind: ZoneLibrary, Owner: me.ID}, id); err != nil {
		t.Fatalf("tuck: %v", err)
	}
	c, zone, ok := cardAnywhere(g, id)
	if !ok || zone != ZoneLibrary {
		t.Fatalf("card is in %q, want a library", zone)
	}
	var revealed bool
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventRevealCards && ev.CardID == id {
				revealed = true
			}
		}
	})
	if !revealed {
		t.Error("CR 708.9: the permanent was not revealed on its way into the library")
	}
	if c.IsKnownTo(me.ID) || c.IsKnownTo(opp.ID) {
		t.Error("CR 401.2: a seat can still read a card in a library because the reveal ran after the library cleared its knowers")
	}
}

// TestFaceDownExileIsNotRevealedWhenItLeaves is the boundary of the
// rule above: CR 708.9 is about PERMANENTS. A face-down exiled card
// that is moved to a hidden zone is subject to the ordinary knowledge
// rules and nothing more.
func TestFaceDownExileIsNotRevealedWhenItLeaves(t *testing.T) {
	g, id := faceDownGame(t, "Griselbrand", "Legendary Creature — Demon", "oracle-grissy")
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		if _, err := g.ExileTopFaceDownForEffect(me.ID, 1); err != nil {
			t.Fatalf("exile face down: %v", err)
		}
		if err := g.BounceToHandForEffect(id); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	c, zone, ok := cardAnywhere(g, id)
	if !ok || zone != ZoneHand {
		t.Fatalf("card is in %q, want a hand", zone)
	}
	if c.IsKnownTo(opp.ID) {
		t.Error("a face-down EXILED card moved to a hand was revealed to the table")
	}
}

// TestFaceDownCardsAreRevealedWhenTheirOwnerLeaves is CR 702.143f's
// half that can be built before foretell exists.
func TestFaceDownCardsAreRevealedWhenTheirOwnerLeaves(t *testing.T) {
	g, id := faceDownGame(t, "Griselbrand", "Legendary Creature — Demon", "oracle-grissy")
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if _, err := g.ExileTopFaceDownForEffect(me.ID, 1); err != nil {
			t.Fatalf("exile face down: %v", err)
		}
		g.revealFaceDownOwnedByLocked(me.ID)
	})
	var revealed bool
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventRevealCards && ev.CardID == id {
				revealed = true
			}
		}
	})
	if !revealed {
		t.Error("CR 702.143f: the leaver's face-down exiles were not revealed")
	}
}

// --- decisions 1 and 8: the state, undo and the snapshot ------------

// TestSetFaceDownKeepsTheFlagAndTheKindInStep pins decision 1's
// invariant. "Face down with no rule attached" has no viewers row and
// no CR 708.2 answer, so it must be unrepresentable through the
// writers.
func TestSetFaceDownKeepsTheFlagAndTheKindInStep(t *testing.T) {
	var c Card
	c.SetFaceDown(FaceDownMorphed)
	if !c.FaceDown || c.FaceDownKind != FaceDownMorphed || !c.FaceDownIsPermanent() {
		t.Fatalf("SetFaceDown = (%v, %q)", c.FaceDown, c.FaceDownKind)
	}
	c.SetFaceDown(FaceDownNone)
	if c.FaceDown || c.FaceDownKind != FaceDownNone {
		t.Errorf("SetFaceDown(none) = (%v, %q), want face up", c.FaceDown, c.FaceDownKind)
	}
	c.SetFaceDown(FaceDownExiled)
	c.ClearFaceDown()
	if c.FaceDown || c.FaceDownKind != FaceDownNone {
		t.Errorf("ClearFaceDown = (%v, %q), want face up", c.FaceDown, c.FaceDownKind)
	}
	if (Card{FaceDown: true}).FaceDownIsPermanent() {
		t.Error("a kindless face-down card reads as a CR 708.2 permanent")
	}
}

// TestUndoRestoresTheFaceDownState: an undo restores a clone, so the
// kind has to ride it. Without the field on the clone a manifested
// permanent comes back as a face-up Griselbrand.
func TestUndoRestoresTheFaceDownState(t *testing.T) {
	g, id := faceDownGame(t, "Griselbrand", "Legendary Creature — Demon", "oracle-grissy")
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if _, err := g.ManifestForEffect(me.ID); err != nil {
			t.Fatalf("manifest: %v", err)
		}
	})
	before := g.Clone()
	// Move it out, which clears the state (CR 400.7), then rewind.
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(id); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	if c, _, _ := cardAnywhere(g, id); c.FaceDown {
		t.Fatal("fixture: the bounce did not clear the state, so the undo proves nothing")
	}
	g.RestoreFrom(before)
	c, zone, ok := cardAnywhere(g, id)
	if !ok || zone != ZoneBattlefield {
		t.Fatalf("after undo the card is in %q, want the battlefield", zone)
	}
	if !c.FaceDown || c.FaceDownKind != FaceDownManifested {
		t.Errorf("after undo the state is (%v, %q), want (true, manifested)", c.FaceDown, c.FaceDownKind)
	}
	if !c.IsCreature() || c.CurrentPower() != 2 {
		t.Errorf("after undo the object is %s %d/%d, want a 2/2 creature",
			c.TypeLine, c.CurrentPower(), c.CurrentToughness())
	}
}

// TestOldSnapshotsReadAFaceDownCardAsANecropotenceExile pins the
// migration: a file written before the kind existed carries none, and
// a face-down card with no rule attached is not a state the engine
// should come back in.
func TestOldSnapshotsReadAFaceDownCardAsANecropotenceExile(t *testing.T) {
	got := restoreCard(&cardSnapshot{InstanceID: uuid.New(), FaceDown: true})
	if got.FaceDownKind != FaceDownExiled {
		t.Errorf("restored kind = %q, want %q", got.FaceDownKind, FaceDownExiled)
	}
	if up := restoreCard(&cardSnapshot{InstanceID: uuid.New()}); up.FaceDown || up.FaceDownKind != FaceDownNone {
		t.Errorf("a face-up card restored as (%v, %q)", up.FaceDown, up.FaceDownKind)
	}
}
