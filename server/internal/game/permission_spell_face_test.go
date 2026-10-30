package game

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// permission_spell_face_test.go — #4509: what a cast permission's type
// filter is read off.
//
// A permission phrased about the SPELL ("cast an instant or sorcery
// spell from your graveyard", "creature spells from the top of your
// library", "play lands from your graveyard") judges the face being
// cast or played:
//
//	CR 601.3e  an alternative set of characteristics replaces the
//	           object's own for the cast-legality check
//	CR 709.3a  only the chosen half of a split card is evaluated
//	CR 712.11c only the face that will be up on the stack of a double-
//	           faced card is evaluated
//	CR 715.3a  only an Adventure's alternative characteristics are
//	           evaluated when it is cast as one
//
// A permission phrased about the CARD ("each instant and sorcery card in
// your graveyard has flashback") picks cards as they are in the zone
// (PermissionFilter.CardInZone):
//
//	CR 715.4   off the stack an adventurer card has only its normal
//	           characteristics
//	CR 712.8a  in a graveyard a double-faced card has only its front
//	           face's characteristics
//	CR 709.4c  a split card has every card type either half has
//
// Every fixture prints {0} so the tests are about faces, not mana.

// mdfcFixture is a modal double-faced card with the two faces given.
func mdfcFixture(owner uuid.UUID, front, back Face) Card {
	c := Card{
		InstanceID: uuid.New(),
		OracleID:   "cccccccc-dddd-eeee-ffff-000000004509",
		Owner:      owner,
		Controller: owner,
		Layout:     LayoutModalDFC,
		Faces:      []Face{front, back},
	}
	c.SetFace(0)
	return c
}

// creatureSorceryMDFC is the Selfless Glyphweaver // Deadly Vanity
// shape: a creature front and a sorcery back.
func creatureSorceryMDFC(owner uuid.UUID) Card {
	return mdfcFixture(owner,
		Face{Name: "Fixture Glyphweaver", TypeLine: "Creature — Human Cleric", ManaCost: "{0}",
			Colors: []string{"W"}, Power: 2, Toughness: 2},
		Face{Name: "Fixture Vanity", TypeLine: "Sorcery", ManaCost: "{0}", Colors: []string{"B"}},
	)
}

// sorceryLandMDFC is the Emeria's Call shape: a sorcery front and a land
// back.
func sorceryLandMDFC(owner uuid.UUID) Card {
	return mdfcFixture(owner,
		Face{Name: "Fixture Call", TypeLine: "Sorcery", ManaCost: "{0}", Colors: []string{"W"}},
		Face{Name: "Fixture Skyclave", TypeLine: "Land"},
	)
}

// freeSplit is an instant // sorcery split card that costs nothing.
func freeSplit(owner uuid.UUID) Card {
	return splitFixture(owner,
		Face{Name: "Fixture Commit", TypeLine: "Instant", ManaCost: "{0}", Colors: []string{"U"}},
		Face{Name: "Fixture Memory", TypeLine: "Sorcery", ManaCost: "{0}", Colors: []string{"U"}},
	)
}

// seedInGraveyard puts c in p's graveyard and returns its ID.
func seedInGraveyard(g *Game, p *Player, c Card) uuid.UUID {
	g.WithWriteLock(func() { p.Graveyard.PushTop(c) })
	return c.InstanceID
}

// grantStandingFilter stores a standing permission for p over `zone`,
// narrowed by f, for the rest of the turn.
func grantStandingFilter(g *Game, p *Player, zone ZoneKind, f PermissionFilter) {
	g.WithWriteLock(func() {
		g.GrantCastPermissionForEffect(CastPermission{
			Player: p.ID, Zone: zone, Scope: ScopeStanding, Filter: f, Label: "test permission",
		})
	})
}

// facesUnderGrant is CastableFacesUnder for the permission that covers
// the card, nil when none does.
func facesUnderGrant(g *Game, p *Player, id uuid.UUID, zone ZoneKind) []int {
	var out []int
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			return
		}
		perm := g.CastPermissionForLocked(p.ID, c, zone)
		if perm == nil {
			return
		}
		out = CastableFacesUnder(c, perm, p.ID)
	})
	return out
}

// Bilbo's and Kess's shape: an instant or sorcery SPELL from a graveyard.
// The adventurer card in the graveyard is a creature card, and its
// Adventure is an instant spell, which is what the permission opens.
func TestSpellFilterOpensAnAdventureFromTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	grantStandingFilter(g, me, ZoneGraveyard, PermissionFilter{InstantOrSorceryOnly: true})
	id := seedInGraveyard(g, me, adventureFixture(me.ID))

	if got := facesUnderGrant(g, me, id, ZoneGraveyard); !slices.Equal(got, []int{1}) {
		t.Fatalf("faces the permission opens = %v, want only the Adventure [1]", got)
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", Face: 0, Strict: true}); err == nil {
		t.Fatal("the creature half was cast through an instant-or-sorcery permission")
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", Face: 1, Strict: true}); err != nil {
		t.Fatalf("cast the Adventure from the graveyard: %v", err)
	}
	spell := findCardForTest(g.Stack, id)
	if spell == nil || spell.ActiveFace != 1 || !spell.IsInstant() || spell.IsCreature() || spell.Name != "Fixture Insight" {
		t.Fatalf("on the stack: %+v, want the Adventure's characteristics alone (CR 715.3b)", spell)
	}
	AssertFaceInvariant(t, g)
}

// The other direction, and the one the front-face reading got wrong in
// the player's favour: a CREATURE spell permission opens the creature
// half of an adventurer card and not its Adventure (the Giant Killer
// ruling of 2019-10-04 under Vivien, Champion of the Wilds).
func TestCreatureSpellFilterOpensOnlyTheCreatureHalf(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	grantStandingFilter(g, me, ZoneGraveyard, PermissionFilter{CreatureOnly: true})
	id := seedInGraveyard(g, me, adventureFixture(me.ID))

	if got := facesUnderGrant(g, me, id, ZoneGraveyard); !slices.Equal(got, []int{0}) {
		t.Fatalf("faces the permission opens = %v, want only the creature [0]", got)
	}
	err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", Face: 1, Strict: true})
	if err == nil {
		t.Fatal("the Adventure was cast through a creature-spell permission")
	}
	if !errors.Is(err, ErrCastZoneNotAllowed) {
		t.Errorf("err = %v, want ErrCastZoneNotAllowed: nothing opens the graveyard for that half", err)
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", Face: 0, Strict: true}); err != nil {
		t.Fatalf("cast the creature from the graveyard: %v", err)
	}
	if spell := findCardForTest(g.Stack, id); spell == nil || spell.ActiveFace != 0 || !spell.IsCreature() {
		t.Fatalf("on the stack: %+v, want the creature", spell)
	}
}

// CR 712.11b, 712.11c: a modal DFC is cast as the face chosen, and only
// that face is judged. The back face's sorcery is what the permission
// opens; the creature front is not.
func TestSpellFilterOpensAModalDFCsBackFace(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	grantStandingFilter(g, me, ZoneGraveyard, PermissionFilter{InstantOrSorceryOnly: true})
	id := seedInGraveyard(g, me, creatureSorceryMDFC(me.ID))

	if got := facesUnderGrant(g, me, id, ZoneGraveyard); !slices.Equal(got, []int{1}) {
		t.Fatalf("faces the permission opens = %v, want only the sorcery back [1]", got)
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", Face: 0, Strict: true}); err == nil {
		t.Fatal("the creature front was cast through an instant-or-sorcery permission")
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", Face: 1, Strict: true}); err != nil {
		t.Fatalf("cast the sorcery back face: %v", err)
	}
	spell := findCardForTest(g.Stack, id)
	if spell == nil || spell.ActiveFace != 1 || !spell.IsSorcery() || spell.IsCreature() || spell.Name != "Fixture Vanity" {
		t.Fatalf("on the stack: %+v, want the back face's characteristics alone (CR 712.8f)", spell)
	}
	passPriorityUntilResolvedForTest(t, g, id)
	if back := findCardForTest(me.Graveyard, id); back == nil || back.ActiveFace != 0 {
		t.Fatalf("after resolving: %+v, want the card in the graveyard front face up (CR 712.8a)", back)
	}
	AssertFaceInvariant(t, g)
}

// "Play lands from your graveyard" (Crucible of Worlds, Ramunap
// Excavator) is about the land played: a modal DFC whose back face is a
// land is played as that face (CR 712.12; the MDFC ruling of 2020-09-25,
// "you may play or cast a modal double-faced card with any face that
// fits the criteria of that effect"), and its sorcery front is not cast.
func TestLandsFilterPlaysAModalDFCsLandFace(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	grantStandingFilter(g, me, ZoneGraveyard, PermissionFilter{LandsOnly: true})
	id := seedInGraveyard(g, me, sorceryLandMDFC(me.ID))

	if got := facesUnderGrant(g, me, id, ZoneGraveyard); !slices.Equal(got, []int{1}) {
		t.Fatalf("faces the permission opens = %v, want only the land back [1]", got)
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", Face: 0, Strict: true}); err == nil {
		t.Fatal("the sorcery front was cast through a play-lands permission")
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", Face: 1}); err != nil {
		t.Fatalf("play the land face from the graveyard: %v", err)
	}
	land := findCardForTest(g.Battlefield, id)
	if land == nil || land.ActiveFace != 1 || !land.IsLand() || land.Name != "Fixture Skyclave" {
		t.Fatalf("on the battlefield: %+v, want the land face", land)
	}
	AssertFaceInvariant(t, g)
}

// The same filter on a Town whose Adventure is a sorcery (Final Fantasy):
// the card is a land card, so it is played, and its Adventure is a
// sorcery spell "play lands" does not open. Before #4509 the front-face
// reading covered the card and the face choice then let the Adventure be
// cast out of the graveyard.
func TestLandsFilterDoesNotCastATownsAdventure(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	grantStandingFilter(g, me, ZoneGraveyard, PermissionFilter{LandsOnly: true})
	id := seedInGraveyard(g, me, landAdventureFixture(me.ID))

	if got := facesUnderGrant(g, me, id, ZoneGraveyard); !slices.Equal(got, []int{0}) {
		t.Fatalf("faces the permission opens = %v, want only the land [0]", got)
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", Face: 1, Strict: true}); err == nil {
		t.Fatal("a Town's Adventure was cast through a play-lands permission")
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", Face: 0}); err != nil {
		t.Fatalf("play the Town from the graveyard: %v", err)
	}
	if land := findCardForTest(g.Battlefield, id); land == nil || !land.IsLand() {
		t.Fatalf("the Town is not on the battlefield: %+v", land)
	}
}

// CR 709.3a: only the half being cast is evaluated, so "sorcery spells"
// opens a split card's sorcery half and not its instant half. Phrased
// about the card instead, the card in the graveyard is an instant AND a
// sorcery card (CR 709.4c), and both halves are open.
func TestSplitCardHalvesUnderASpellFilterAndACardFilter(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	grantStandingFilter(g, me, ZoneGraveyard, PermissionFilter{SorceryOnly: true})
	id := seedInGraveyard(g, me, freeSplit(me.ID))

	if got := facesUnderGrant(g, me, id, ZoneGraveyard); !slices.Equal(got, []int{1}) {
		t.Fatalf("faces a sorcery-spell permission opens = %v, want the sorcery half [1]", got)
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", Face: 0, Strict: true}); err == nil {
		t.Fatal("the instant half was cast through a sorcery-spell permission")
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", Face: 1, Strict: true}); err != nil {
		t.Fatalf("cast the sorcery half: %v", err)
	}
	if spell := findCardForTest(g.Stack, id); spell == nil || !spell.IsSorcery() || spell.IsInstant() {
		t.Fatalf("on the stack: %+v, want the sorcery half alone (CR 709.3b)", spell)
	}

	g2 := newActiveGame(t)
	me2 := g2.Seats[0]
	grantStandingFilter(g2, me2, ZoneGraveyard, PermissionFilter{CardInZone: true, SorceryOnly: true})
	id2 := seedInGraveyard(g2, me2, freeSplit(me2.ID))
	if got := facesUnderGrant(g2, me2, id2, ZoneGraveyard); !slices.Equal(got, []int{0, 1}) {
		t.Fatalf("faces a sorcery-card permission opens = %v, want both halves [0 1]", got)
	}
}

// Lier's and Wrenn and Six's shape: "instant and sorcery CARDS in your
// graveyard". An adventurer card there is a creature card (CR 715.4) and
// a creature-front modal DFC is a creature card (CR 712.8a), so neither
// is covered, whatever their other halves are.
func TestCardFilterJudgesTheCardInTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	grantStandingFilter(g, me, ZoneGraveyard, PermissionFilter{CardInZone: true, InstantOrSorceryOnly: true})
	adventure := seedInGraveyard(g, me, adventureFixture(me.ID))
	mdfc := seedInGraveyard(g, me, creatureSorceryMDFC(me.ID))
	split := seedInGraveyard(g, me, freeSplit(me.ID))

	for name, id := range map[string]uuid.UUID{"the adventurer card": adventure, "the creature-front modal DFC": mdfc} {
		if perm := grantOn(g, me.ID, id, ZoneGraveyard); perm.Granted() {
			t.Errorf("%s is covered by an instant-and-sorcery-cards permission", name)
		}
		if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", Face: 1, Strict: true}); err == nil {
			t.Errorf("%s's instant or sorcery half was cast", name)
		}
	}
	if perm := grantOn(g, me.ID, split, ZoneGraveyard); !perm.Granted() {
		t.Error("an instant // sorcery split card is not covered")
	}
}

// Six's shape: "nonland permanent CARDS in your graveyard have retrace".
// An adventurer card is a creature card there, so it has the ability,
// and retrace names no type: the card may be cast as either face, its
// Adventure included (CR 715.3).
func TestCardFilterOpensEveryFaceOfACardItCovers(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	grantStandingFilter(g, me, ZoneGraveyard, PermissionFilter{CardInZone: true, NonLandPermanentOnly: true})
	id := seedInGraveyard(g, me, adventureFixture(me.ID))

	if got := facesUnderGrant(g, me, id, ZoneGraveyard); !slices.Equal(got, []int{0, 1}) {
		t.Fatalf("faces = %v, want both [0 1]", got)
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", Face: 1, Strict: true}); err != nil {
		t.Fatalf("cast the Adventure under a permission about the card: %v", err)
	}

	// Phrased about the spell, the same filter would open the creature
	// alone: "cast permanent spells from your graveyard".
	g2 := newActiveGame(t)
	me2 := g2.Seats[0]
	grantStandingFilter(g2, me2, ZoneGraveyard, PermissionFilter{NonLandPermanentOnly: true})
	id2 := seedInGraveyard(g2, me2, adventureFixture(me2.ID))
	if got := facesUnderGrant(g2, me2, id2, ZoneGraveyard); !slices.Equal(got, []int{0}) {
		t.Fatalf("faces a permanent-spell permission opens = %v, want the creature [0]", got)
	}
}

// Nivix, Aerie of the Firemind's shape: one exiled card, "you may cast it
// if it's an instant or sorcery spell". The gate on a ScopeCards
// permission is about the spell too.
func TestNamedCardSpellFilterJudgesTheFaceBeingCast(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	card := adventureFixture(me.ID)
	g.WithWriteLock(func() {
		g.Exile.PushTop(card)
		g.GrantCastPermissionToCardsForEffect(CastPermission{
			Player: me.ID, Zone: ZoneExile, Filter: PermissionFilter{InstantOrSorceryOnly: true},
			Label: "test exile permission",
		}, []Card{card})
	})
	id := card.InstanceID

	if got := facesUnderGrant(g, me, id, ZoneExile); !slices.Equal(got, []int{1}) {
		t.Fatalf("faces = %v, want the Adventure [1]", got)
	}
	err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "exile", Face: 0, Strict: true})
	if !errors.Is(err, ErrNoPlayPermission) {
		t.Fatalf("cast the creature out of exile: err = %v, want ErrNoPlayPermission", err)
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "exile", Face: 1, Strict: true}); err != nil {
		t.Fatalf("cast the Adventure out of exile: %v", err)
	}
}

// The view and the bot read the same gate (CastOffersForLocked): a face
// the permission does not open has no price at all, and a face it opens
// has its printed one.
func TestCastOffersFollowTheSpellFilterFaceByFace(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	grantStandingFilter(g, me, ZoneGraveyard, PermissionFilter{InstantOrSorceryOnly: true})
	id := seedInGraveyard(g, me, adventureFixture(me.ID))

	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(id)
		perm := g.CastPermissionForLocked(me.ID, c, ZoneGraveyard)
		creature, adventure := c, c
		creature.SetFace(0)
		adventure.SetFace(1)
		if offers := g.CastOffersForLocked(me.ID, creature, ZoneGraveyard, perm); len(offers) != 0 {
			t.Errorf("the creature half has %d offers, want none", len(offers))
		}
		if offers := g.CastOffersForLocked(me.ID, adventure, ZoneGraveyard, perm); len(offers) != 1 || offers[0] != nil {
			t.Errorf("the Adventure's offers = %v, want its printed cost alone", offers)
		}
	})
}

// A HAND permission prices one claim (ForClaim), so its filter and its
// cap narrow that claim and not the card's ordinary cast from hand
// (CR 601.2): "you may cast an instant or sorcery spell with mana value
// 0 or less from your hand without paying its mana cost".
func TestHandGrantFilterNarrowsOnlyItsOwnClaim(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	card := mdfcFixture(me.ID,
		Face{Name: "Fixture Bear", TypeLine: "Creature — Bear", ManaCost: "{0}", Colors: []string{"G"}, Power: 2, Toughness: 2},
		Face{Name: "Fixture Bolt", TypeLine: "Instant", ManaCost: "{3}", Colors: []string{"R"}},
	)
	cheap := mdfcFixture(me.ID,
		Face{Name: "Fixture Elk", TypeLine: "Creature — Elk", ManaCost: "{3}", Colors: []string{"G"}, Power: 3, Toughness: 3},
		Face{Name: "Fixture Zap", TypeLine: "Instant", ManaCost: "{0}", Colors: []string{"R"}},
	)
	zero := 0
	var grant CastPermission
	g.WithWriteLock(func() {
		me.Hand.PushTop(card)
		me.Hand.PushTop(cheap)
		grant = g.FreeCastFromHandPermission(me.ID, uuid.Nil, "Test")
		grant.Filter = PermissionFilter{InstantOrSorceryOnly: true}
		grant.MaxSpellManaValue = &zero
		grant.AltCostKey = AltCostKeyGrantedCost
	})

	if grant.OffersASpell(card) {
		t.Error("the {3} instant back is under a mana value 0 cap, and the front is a creature: no spell to offer")
	}
	if !grant.OffersASpell(cheap) {
		t.Error("the {0} instant back is a spell the clause admits")
	}

	g.ReadSnapshot(func() {
		bear, bolt, zap := card, card, cheap
		bear.SetFace(0)
		bolt.SetFace(1)
		zap.SetFace(1)
		// The ordinary casts from hand stand: the grant is not what they
		// are cast under.
		for name, c := range map[string]Card{"the creature front": bear, "the {3} instant back": bolt} {
			if err := g.validateCastPathLocked(c, ZoneHand, nil, &grant); err != nil {
				t.Errorf("%s's ordinary cast from hand: %v", name, err)
			}
		}
		// The free claim: refused on the creature (not an instant or
		// sorcery spell) and on the instant over the cap, open on the
		// {0} instant.
		if offer := grant.AlternativeCostFor(bear); offer != nil {
			if err := g.validateCastPathLocked(bear, ZoneHand, offer, &grant); err == nil {
				t.Error("the free claim was open on a creature spell")
			}
		}
		if offer := grant.AlternativeCostFor(bolt); offer != nil {
			if err := g.validateCastPathLocked(bolt, ZoneHand, offer, &grant); !errors.Is(err, ErrSpellManaValueTooHigh) {
				t.Errorf("the free claim on a mana value 3 spell: err = %v, want ErrSpellManaValueTooHigh", err)
			}
		}
		offer := grant.AlternativeCostFor(zap)
		if offer == nil {
			t.Fatal("no free offer on the {0} instant")
		}
		if err := g.validateCastPathLocked(zap, ZoneHand, offer, &grant); err != nil {
			t.Errorf("the free claim on the {0} instant: %v", err)
		}
	})
}

// OffersASpell is what a card's offer picks its cards by: a face that is
// not a land and that the filter admits.
func TestOffersASpellReadsEveryFace(t *testing.T) {
	owner := uuid.New()
	bilbo := CastPermission{Filter: PermissionFilter{ArtifactInstantOrSorceryOnly: true}, CastOnly: true}
	vault := NewCard("Fixture Vault", owner)
	vault.TypeLine = "Artifact Land"
	bear := NewCard("Fixture Bear", owner)
	bear.TypeLine = "Creature — Bear"
	bear.ManaCost = "{1}{G}"
	for _, tc := range []struct {
		name string
		card Card
		want bool
	}{
		{"an adventurer card with an instant Adventure", adventureFixture(owner), true},
		{"a creature // sorcery modal DFC", creatureSorceryMDFC(owner), true},
		{"a Town whose Adventure is a sorcery", landAdventureFixture(owner), true},
		{"an instant // sorcery split card", freeSplit(owner), true},
		{"an artifact land", vault, false},
		{"a creature card", bear, false},
	} {
		if got := bilbo.OffersASpell(tc.card); got != tc.want {
			t.Errorf("%s: OffersASpell = %v, want %v", tc.name, got, tc.want)
		}
	}
}
