package game

import (
	"testing"

	"github.com/google/uuid"
)

// face_down_listed_test.go — ADR 0082's second 2026-09-23 amendment:
// CR 708.2's LISTED characteristics (#1270) and CR 613.7f's
// face-change timestamp (#1271).
//
//	1. a listing REPLACES the CR 708.2a body — the Cyberman and the Forest
//	2. the three doors: turn face down, graveyard return, library put
//	3. what a listing does not survive: a zone change, turning face up,
//	   CR 708.2b
//	4. copiable values (CR 708.2's second sentence)
//	5. undo and the snapshot
//	6. CR 613.7f: the timestamp, both directions

// cybermanListing is the body Cyber Conversion, Cybership, Missy and
// The Cyber-Controller all list.
func cybermanListing() *FaceDownListing {
	return &FaceDownListing{
		Types:     []string{"Artifact", "Creature"},
		Subtypes:  []string{"Cyberman"},
		Power:     2,
		Toughness: 2,
	}
}

// forestListing is Yedora, Grave Gardener's.
func forestListing() *FaceDownListing {
	return &FaceDownListing{Types: []string{"Land"}, Subtypes: []string{"Forest"}}
}

// --- 1. the body -------------------------------------------------------

// TestAListedBodyReplacesTheDefaultTwoTwo is the headline: CR 708.2's
// "no characteristics other than those listed". A Cyber Conversion'd
// Sheoldred is an ARTIFACT creature and a CYBERMAN — Shatter can hit
// it and a Cyberman lord pumps it — and still has no name, no cost, no
// colour and no text.
func TestAListedBodyReplacesTheDefaultTwoTwo(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := turnableCreature(t, g, me, "Sheoldred")

	g.WithWriteLock(func() { g.TurnFaceDownListedForEffect(uuid.Nil, cybermanListing(), id) })

	cold := *findBattlefieldCard(g, id)
	// The cold-cache accessors first: they take a fast path off the
	// printed fields before the first recompute, and the face-down
	// guard on that path is where a listing is easiest to miss.
	if !cold.IsArtifact() || !cold.IsCreature() || !cold.HasSubtype("Cyberman") {
		t.Errorf("cold cache: artifact %v, creature %v, Cyberman %v — want all three",
			cold.IsArtifact(), cold.IsCreature(), cold.HasSubtype("Cyberman"))
	}
	if cold.HasSubtype("Demon") {
		t.Error("cold cache: the card underneath's subtype leaked through")
	}

	c := layeredBattlefieldCard(t, g, id)
	eff := c.Effective()
	if !typeListHas(eff.Types, "artifact") || !typeListHas(eff.Types, "creature") {
		t.Errorf("effective types = %v, want Artifact Creature", eff.Types)
	}
	if len(eff.Subtypes) != 1 || eff.Subtypes[0] != "Cyberman" {
		t.Errorf("effective subtypes = %v, want [Cyberman]", eff.Subtypes)
	}
	if c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("P/T = %d/%d, want the listed 2/2", c.CurrentPower(), c.CurrentToughness())
	}
	if eff.Name != "" || len(eff.Colors) != 0 || len(eff.Abilities) != 0 {
		t.Errorf("the listed body carries (%q, %v, %v) — CR 708.2 lists none of them",
			eff.Name, eff.Colors, eff.Abilities)
	}
	if CatalogKey(c) != "" {
		t.Error("a listed body has text — CatalogKey must stay silent (CR 708.2)")
	}
	if !c.HasSubtype("Cyberman") || c.HasSubtype("Demon") {
		t.Error("warm cache: HasSubtype does not read the listed subtype alone")
	}
}

// TestAListedForestIsALandAndNotACreature is Yedora's case, and the
// reason a listing is a REPLACEMENT: "It's a Forest land" is not a 2/2
// with a land type added. It is not a creature, it has no P/T, and —
// CR 305.6, not the listing — it taps for {G}.
func TestAListedForestIsALandAndNotACreature(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := turnableCreature(t, g, me, "Sheoldred")

	g.WithWriteLock(func() { g.TurnFaceDownListedForEffect(uuid.Nil, forestListing(), id) })

	cold := *findBattlefieldCard(g, id)
	if cold.IsCreature() || !cold.IsLand() {
		t.Errorf("cold cache: creature %v, land %v — want a land and not a creature", cold.IsCreature(), cold.IsLand())
	}
	if got := intrinsicLandManaAbilities(cold); len(got) != 1 || got[0].Produced != "{G}" {
		t.Errorf("cold cache: intrinsic mana = %+v, want one {G}", got)
	}

	c := layeredBattlefieldCard(t, g, id)
	if c.IsCreature() || !c.IsLand() || !c.HasSubtype("Forest") {
		t.Errorf("creature %v, land %v, Forest %v — want a Forest land", c.IsCreature(), c.IsLand(), c.HasSubtype("Forest"))
	}
	abs := ManaAbilitiesForCard(c)
	if len(abs) != 1 || abs[0].Produced != "{G}" {
		t.Errorf("ManaAbilitiesForCard = %+v, want exactly CR 305.6's {T}: Add {G}", abs)
	}
}

// TestAFaceDownTokenHasNoCarriedManaAbility closes the door the
// listing work found open: CatalogKey silences a face-down CARD, but a
// token carries its mana ability on the instance, and a face-down
// Treasure is not a Treasure (CR 708.2a).
func TestAFaceDownTokenHasNoCarriedManaAbility(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := turnableCreature(t, g, me, "Treasure")
	c := findBattlefieldCard(g, id)
	c.TypeLine = "Token Artifact — Treasure"
	c.ManaAbilities = []ManaAbilityShape{{TapCost: true, Produced: "{W|U|B|R|G}", Label: "Add one mana"}}
	if len(ManaAbilitiesForCard(*c)) != 1 {
		t.Fatal("fixture: the face-up Treasure has no mana ability, so the test proves nothing")
	}

	g.WithWriteLock(func() { g.TurnFaceDownForEffect(uuid.Nil, id) })

	if got := ManaAbilitiesForCard(layeredBattlefieldCard(t, g, id)); len(got) != 0 {
		t.Errorf("a face-down token still has %+v", got)
	}
}

// --- 2. the doors ------------------------------------------------------

// TestReturnFromGraveyardFaceDownLandsAListedForest walks Yedora's
// door: the reanimation pipeline, a face-down entry, the listed body,
// the controller as its only knower, and the kind an EFFECT gives.
func TestReturnFromGraveyardFaceDownLandsAListedForest(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c := NewCard("Sheoldred", me.ID)
	c.TypeLine = "Legendary Creature — Phyrexian Praetor"
	c.OracleID = "oracle-sheoldred"
	c.Power, c.Toughness = 4, 5
	id := c.InstanceID
	g.WithWriteLock(func() { me.Graveyard.PushTop(c) })

	var entered uuid.UUID
	var err error
	g.WithWriteLock(func() {
		entered, err = g.ReturnFromGraveyardFaceDownForEffect(id, uuid.Nil, false, forestListing())
	})
	if err != nil || entered != id {
		t.Fatalf("ReturnFromGraveyardFaceDownForEffect = (%s, %v), want (%s, nil)", entered, err, id)
	}
	got := layeredBattlefieldCard(t, g, id)
	if !got.FaceDown || got.FaceDownKind != FaceDownTurned {
		t.Fatalf("state = (%v, %q), want (true, %q)", got.FaceDown, got.FaceDownKind, FaceDownTurned)
	}
	if got.IsCreature() || !got.IsLand() || !got.HasSubtype("Forest") {
		t.Error("the returned permanent is not the listed Forest land")
	}
	if got.IsLegendary() {
		t.Error("the card underneath's supertype leaked through")
	}
	if got.Controller != me.ID || !got.IsKnownTo(me.ID) || got.IsKnownTo(opp.ID) {
		t.Errorf("controller %s, known to me %v, to opponent %v — CR 708.5 says the controller alone",
			got.Controller, got.IsKnownTo(me.ID), got.IsKnownTo(opp.ID))
	}
	// CR 708.7: an effect that did not give it a way back up, and a card
	// that prints no morph — it stays down.
	if TurnFaceUpOffer(got) != nil {
		t.Error("a face-down Forest with no morph was offered a way face up")
	}
}

// TestPutFromLibraryFaceDownLandsAListedCyberman is Cybership's door:
// the top of a library onto the battlefield face down — an INSTANT,
// which only the face-down entry may put there (CR 110.4 lifted, as
// for manifest) — as a 2/2 Cyberman artifact creature.
func TestPutFromLibraryFaceDownLandsAListedCyberman(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c := NewCard("Lightning Bolt", opp.ID)
	c.TypeLine = "Instant"
	c.OracleID = "oracle-bolt"
	id := c.InstanceID
	g.WithWriteLock(func() { opp.Library.PushTop(c) })

	var entered []uuid.UUID
	var err error
	g.WithWriteLock(func() {
		entered, err = g.PutCardsFromLibraryOntoBattlefieldForEffect([]uuid.UUID{id}, LibraryEntryOptions{
			Controller:     me.ID,
			FaceDown:       FaceDownTurned,
			FaceDownListed: cybermanListing(),
		})
	})
	if err != nil || len(entered) != 1 {
		t.Fatalf("put = (%v, %v), want one entry", entered, err)
	}
	got := layeredBattlefieldCard(t, g, id)
	if !got.IsArtifact() || !got.IsCreature() || !got.HasSubtype("Cyberman") || got.CurrentPower() != 2 {
		t.Errorf("entered as %v %v %d/%d, want a 2/2 Cyberman artifact creature",
			got.Effective().Types, got.Effective().Subtypes, got.CurrentPower(), got.CurrentToughness())
	}
	if got.Controller != me.ID || got.Owner != opp.ID {
		t.Errorf("controller/owner = %s/%s, want me/opponent", got.Controller, got.Owner)
	}
	if !got.IsKnownTo(me.ID) || got.IsKnownTo(opp.ID) {
		t.Error("CR 708.5: the controller, and only the controller, may look — even the owner may not")
	}
}

// TestAListedExileIsDropped: a listing names a CR 708.2 object's body,
// and a face-down card in exile has no characteristics at all
// (CR 406.3a). SetFaceDownListed refuses to carry one there.
func TestAListedExileIsDropped(t *testing.T) {
	var c Card
	c.SetFaceDownListed(FaceDownExiled, cybermanListing())
	if c.FaceDownListed != nil {
		t.Error("an exile kind kept a listed body")
	}
	listing := cybermanListing()
	c.SetFaceDownListed(FaceDownTurned, listing)
	listing.Subtypes[0] = "Dalek"
	if c.FaceDownListed.Subtypes[0] != "Cyberman" {
		t.Error("the card aliases the caller's listing")
	}
}

// --- 3. what a listing does not survive ---------------------------------

// TestAListingDiesWithTheObject: CR 400.7. The permanent that goes to
// the graveyard is a new object, and the card there is the card.
func TestAListingDiesWithTheObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := turnableCreature(t, g, me, "Sheoldred")
	g.WithWriteLock(func() { g.TurnFaceDownListedForEffect(uuid.Nil, cybermanListing(), id) })

	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, me.Graveyard, id); err != nil {
			t.Fatalf("move: %v", err)
		}
	})
	c, zone, _ := cardAnywhere(g, id)
	if zone != ZoneGraveyard || c.FaceDown || c.FaceDownListed != nil {
		t.Errorf("in %q: face down %v, listing %+v — want the plain card", zone, c.FaceDown, c.FaceDownListed)
	}
	if c.IsArtifact() || c.HasSubtype("Cyberman") {
		t.Error("the graveyard card is still a Cyberman")
	}
}

// TestTurningFaceUpDropsTheListing: the listed body is CR 708.2's, and
// CR 708.2 stops applying the moment the permanent is face up.
func TestTurningFaceUpDropsTheListing(t *testing.T) {
	withCatalogAlternativeCosts(t, altCostFor(morphOracle, morphOffer(FaceDownMorphed, morphCost, false)))
	g := newActiveGame(t)
	me := g.Seats[0]
	id := turnableCreature(t, g, me, "Willbender")
	findBattlefieldCard(g, id).OracleID = morphOracle
	g.WithWriteLock(func() { g.TurnFaceDownListedForEffect(uuid.Nil, cybermanListing(), id) })
	fillPool(g, me, 8)
	g.WithWriteLock(func() { g.Turn.Step = StepPrecombatMain })

	// CR 702.37e: the card's own morph opens the way up whatever put it
	// down, listed body or not.
	if err := g.PerformSpecialAction(me.ID, id, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up: %v", err)
	}
	c := layeredBattlefieldCard(t, g, id)
	if c.FaceDown || c.FaceDownListed != nil {
		t.Fatalf("still face down (%v) or still listed (%+v)", c.FaceDown, c.FaceDownListed)
	}
	if c.IsArtifact() || c.HasSubtype("Cyberman") || c.CurrentPower() != 7 {
		t.Error("the real card did not come back whole — the Cyberman body outlived the face-down state")
	}
}

// TestAListedTurnOnAFaceDownPermanentChangesNothing is CR 708.2b:
// "nothing happens and that effect doesn't change any of its
// characteristics". Cyber Conversion on a morph does not make it a
// Cyberman.
func TestAListedTurnOnAFaceDownPermanentChangesNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := turnableCreature(t, g, me, "Sheoldred")
	g.WithWriteLock(func() { g.TurnFaceDownForEffect(uuid.Nil, id) })

	var turned []uuid.UUID
	g.WithWriteLock(func() { turned = g.TurnFaceDownListedForEffect(uuid.Nil, cybermanListing(), id) })
	if len(turned) != 0 {
		t.Errorf("turned %v — CR 708.2b refuses a face-down permanent", turned)
	}
	c := layeredBattlefieldCard(t, g, id)
	if c.FaceDownListed != nil || c.IsArtifact() {
		t.Error("the refused turn still listed a body")
	}
}

// --- 4. copiable values ------------------------------------------------

// TestTheListingIsTheCopiableValues is CR 708.2's second sentence:
// "any listed characteristics are the copiable values of that object's
// characteristics" — and, with nothing listed, CR 708.2a's nameless
// 2/2 is. Never the card underneath.
func TestTheListingIsTheCopiableValues(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	plain := turnableCreature(t, g, me, "Sheoldred")
	listed := turnableCreature(t, g, me, "Griselbrand")
	g.WithWriteLock(func() {
		g.TurnFaceDownForEffect(uuid.Nil, plain)
		g.TurnFaceDownListedForEffect(uuid.Nil, cybermanListing(), listed)
	})

	v := CopiableValuesOf(*findBattlefieldCard(g, plain))
	if v.Name != "" || v.OracleID != "" || v.ManaCost != "" || v.TypeLine != "Creature" || v.Power != 2 || v.Toughness != 2 {
		t.Errorf("copy of a plain face-down = %+v, want a nameless 2/2 Creature", v)
	}
	v = CopiableValuesOf(*findBattlefieldCard(g, listed))
	if v.TypeLine != "Artifact Creature — Cyberman" || v.Power != 2 || v.Name != "" || v.OracleID != "" {
		t.Errorf("copy of a listed face-down = %+v, want a nameless 2/2 Artifact Creature — Cyberman", v)
	}
	if len(v.Keywords) != 0 || len(v.Colors) != 0 {
		t.Error("the copy carried the card underneath's keywords or colours")
	}
}

// --- 5. undo and the snapshot ------------------------------------------

func TestUndoAndSnapshotCarryTheListing(t *testing.T) {
	g := newRestorableGame(t)
	me := g.Seats[0]
	id := turnableCreature(t, g, me, "Sheoldred")
	before := g.Clone()
	g.WithWriteLock(func() { g.TurnFaceDownListedForEffect(uuid.Nil, forestListing(), id) })
	after := g.Clone()

	// The clone owns its listing.
	findBattlefieldCard(g, id).FaceDownListed.Subtypes[0] = "Island"
	if s := findBattlefieldCard(after, id).FaceDownListed.Subtypes[0]; s != "Forest" {
		t.Errorf("the undo snapshot's listing reads %q — it aliases the live card", s)
	}
	findBattlefieldCard(g, id).FaceDownListed.Subtypes[0] = "Forest"

	_, restored := roundTrip(t, g)
	c := findBattlefieldCard(restored, id)
	if c == nil || c.FaceDownListed == nil || !c.IsLand() || c.IsCreature() {
		t.Fatal("the snapshot lost the listed Forest — it restored as the default 2/2")
	}
	if c.FaceTurnedAt == 0 {
		t.Error("the snapshot lost the CR 613.7f face-change timestamp")
	}

	g.RestoreFrom(before)
	if c := findBattlefieldCard(g, id); c.FaceDown || c.FaceDownListed != nil {
		t.Error("undo did not rewind the listing")
	}
}

// --- 6. CR 613.7f -------------------------------------------------------

// tickingClock replaces the layer clock with a strictly increasing
// counter for one test, so "later" is never a same-nanosecond tie.
func tickingClock(t *testing.T) {
	t.Helper()
	prev := timeNowUnixNano
	var now int64 = 1_000
	timeNowUnixNano = func() int64 { now += 10; return now }
	t.Cleanup(func() { timeNowUnixNano = prev })
}

// setBaseStatic is a CR 613.4b layer 7b "creatures you control have
// base power and toughness N/N" — the effect whose ORDER decides the
// answer, which is the only way CR 613.7f can be observed.
func setBaseStatic(n int) StaticAbility {
	return StaticAbility{
		Layer:    Layer7PT,
		SubLayer: SubLayer7B_Set,
		AppliesTo: func(target *Card, _ *Game, source *Card) bool {
			return target.IsCreature() && target.Controller == source.Controller && target.InstanceID != source.InstanceID
		},
		Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
			c.Power, c.Toughness = n, n
		},
	}
}

// TestTurningFaceUpTakesANewTimestamp is #1271's two-effect test. The
// morph ("sets to 1/1") entered FIRST and the other lord ("sets to
// 4/4") second, so the bear is 4/4 — the later 7b effect wins
// (CR 613.7). Turned face down and back up, the morph receives a new
// timestamp (CR 613.7f) and is now the LATER of the two: the bear is
// 1/1. Without the re-stamp the morph keeps its entry timestamp and
// the bear stays 4/4.
func TestTurningFaceUpTakesANewTimestamp(t *testing.T) {
	tickingClock(t)
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		switch oracleID {
		case "oracle-Morph":
			return []StaticAbility{setBaseStatic(1)}
		case "oracle-Lord":
			return []StaticAbility{setBaseStatic(4)}
		}
		return nil
	})
	g := newActiveGame(t)
	me := g.Seats[0]
	morph := turnableCreature(t, g, me, "Morph")
	lord := turnableCreature(t, g, me, "Lord")
	bear := turnableCreature(t, g, me, "Bear")
	_ = lord
	// turnableCreature stamps the entries directly rather than through
	// the zone-move event, so tell the layer engine the board moved.
	g.layerVersion.Add(1)

	if p := layeredBattlefieldCard(t, g, bear).CurrentPower(); p != 4 {
		t.Fatalf("fixture: the bear is %d, want 4 — the later lord's set wins", p)
	}
	entered := findBattlefieldCard(g, morph).EnteredBattlefieldAt

	g.WithWriteLock(func() { g.TurnFaceDownForEffect(uuid.Nil, morph) })
	down := findBattlefieldCard(g, morph).FaceTurnedAt
	if down == 0 || down <= entered {
		t.Errorf("turning face down stamped %d (entered %d) — CR 613.7f gives it a new timestamp", down, entered)
	}
	if p := layeredBattlefieldCard(t, g, bear).CurrentPower(); p != 4 {
		t.Fatalf("with the morph face down the bear is %d, want 4 — the morph has no text", p)
	}

	g.WithWriteLock(func() {
		if err := g.turnFaceUpLocked(me, morph, SpecialAction{Kind: SpecialActionTurnFaceUp}); err != nil {
			t.Fatalf("turn face up: %v", err)
		}
	})
	up := findBattlefieldCard(g, morph)
	if up.FaceTurnedAt <= down {
		t.Errorf("turning face up stamped %d, not later than the face-down %d", up.FaceTurnedAt, down)
	}
	if up.EnteredBattlefieldAt != entered {
		t.Error("the entry timestamp moved — it is the object's identity pin and must ride through (CR 708.8)")
	}
	if p := layeredBattlefieldCard(t, g, bear).CurrentPower(); p != 1 {
		t.Errorf("after the morph turned face up the bear is %d, want 1 — "+
			"the morph's static must sort AFTER the lord's (CR 613.7f)", p)
	}
}

// TestTheFaceChangeTimestampGoesWithTheObject: CR 400.7. A permanent
// that leaves the battlefield takes its timestamps with it.
func TestTheFaceChangeTimestampGoesWithTheObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := turnableCreature(t, g, me, "Sheoldred")
	g.WithWriteLock(func() { g.TurnFaceDownForEffect(uuid.Nil, id) })
	if findBattlefieldCard(g, id).FaceTurnedAt == 0 {
		t.Fatal("fixture: no face-change timestamp to clear")
	}
	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, me.Graveyard, id); err != nil {
			t.Fatalf("move: %v", err)
		}
	})
	if c, _, _ := cardAnywhere(g, id); c.FaceTurnedAt != 0 {
		t.Errorf("FaceTurnedAt = %d in the graveyard, want 0", c.FaceTurnedAt)
	}
}
