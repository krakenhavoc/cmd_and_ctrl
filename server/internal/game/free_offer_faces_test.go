package game

import (
	"testing"

	"github.com/google/uuid"
)

// free_offer_faces_test.go — #4697: Beseech the Mirror's "if that spell's
// mana value is N or less" (CastExiledFreeOrToHandForEffect) and Ziatora's
// Envoy's and Rashmi's "a spell with mana value N or less"
// (OfferFreePlayFromTopOrHandForEffect) are about the spell, so a card is
// offered when some face a cast may choose is a spell that fits (CR 601.3e,
// 712.11c, 715.3a), and only that face may be cast. Outside a resolution
// the permission alone is the offer, so these read it straight off the
// card. The card-level tests are cards/effects/free_offer_faces_4697_test.go.

// cheapBackMDFCFixture is a modal DFC whose front face is mana value 7 and
// whose back face is a mana value 2 sorcery.
func cheapBackMDFCFixture(owner uuid.UUID) Card {
	return mdfcFixture(owner,
		Face{Name: "Fixture Colossus", TypeLine: "Creature — Giant", ManaCost: "{5}{B}{B}", Colors: []string{"B"},
			Power: 7, Toughness: 7},
		Face{Name: "Fixture Rite", TypeLine: "Sorcery", ManaCost: "{1}{B}", Colors: []string{"B"}},
	)
}

// bigCreatureCheapAdventure is a mana value 7 creature whose Adventure is
// a mana value 3 sorcery (the Beanstalk Giant shape).
func bigCreatureCheapAdventure(owner uuid.UUID) Card {
	c := Card{
		InstanceID: uuid.New(), Layout: LayoutAdventure, Owner: owner, Controller: owner,
		Faces: []Face{
			{Name: "Fixture Beanstalk", TypeLine: "Creature — Giant", ManaCost: "{6}{G}", Colors: []string{"G"},
				Power: 7, Toughness: 7},
			{Name: "Fixture Footsteps", TypeLine: "Sorcery — Adventure", ManaCost: "{2}{G}", Colors: []string{"G"}},
		},
	}
	c.SetFace(0)
	return c
}

// freeOrHandOver exiles c and makes Beseech's offer over it, capped at 4.
func freeOrHandOver(t *testing.T, c Card) (*Game, *Player, uuid.UUID) {
	t.Helper()
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainPhase(t, g)
	c.Owner, c.Controller = me.ID, me.ID
	g.WithWriteLock(func() {
		g.Exile.PushTop(c)
		if err := g.CastExiledFreeOrToHandForEffect(me.ID, uuid.New(), c.InstanceID, 4); err != nil {
			t.Fatalf("CastExiledFreeOrToHandForEffect: %v", err)
		}
	})
	return g, me, c.InstanceID
}

// freePlayFromTopOver puts c on top of the active player's library and
// makes Ziatora's Envoy's offer (land: true) or Rashmi's (land: false).
func freePlayFromTopOver(t *testing.T, c Card, maxMV int, land bool) (*Game, *Player, uuid.UUID) {
	t.Helper()
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainPhase(t, g)
	c.Owner, c.Controller = me.ID, me.ID
	g.WithWriteLock(func() {
		me.Library.PushTop(c)
		if err := g.OfferFreePlayFromTopOrHandForEffect(FreePlayFromTop{
			Player: me.ID, Source: uuid.New(), SourceName: "Test Envoy", MaxManaValue: maxMV, Land: land,
		}); err != nil {
			t.Fatalf("OfferFreePlayFromTopOrHandForEffect: %v", err)
		}
	})
	return g, me, c.InstanceID
}

func freeCastGrantOn(g *Game, id uuid.UUID) *CastPermission {
	var perm *CastPermission
	g.ReadSnapshot(func() { perm = g.CastPermissionOnCardByIDForEffect(id) })
	return perm
}

func TestFreeOrHandOffersAModalDFCWhoseBackFaceFits(t *testing.T) {
	c := cheapBackMDFCFixture(uuid.Nil)
	g, me, id := freeOrHandOver(t, c)
	if me.Hand.Contains(id) || freeCastGrantOn(g, id) == nil {
		t.Fatal("a modal DFC whose back face fits the cap went to the hand")
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "exile", Face: 0}); err == nil {
		t.Fatal("the mana value 7 front face was cast through a cap of 4")
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "exile", Face: 1}); err != nil {
		t.Fatalf("cast the back face free: %v", err)
	}
	if spell := findCardForTest(g.Stack, id); spell == nil || spell.ActiveFace != 1 || !spell.IsSorcery() {
		t.Fatalf("on the stack: %+v, want the sorcery back face", spell)
	}
}

func TestFreeOrHandOffersACreatureWhoseAdventureFits(t *testing.T) {
	g, me, id := freeOrHandOver(t, bigCreatureCheapAdventure(uuid.Nil))
	if me.Hand.Contains(id) || freeCastGrantOn(g, id) == nil {
		t.Fatal("a creature card whose Adventure fits the cap went to the hand")
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "exile", Face: 1}); err != nil {
		t.Fatalf("cast the Adventure free: %v", err)
	}
	if spell := findCardForTest(g.Stack, id); spell == nil || spell.ActiveFace != 1 || spell.IsCreature() {
		t.Fatalf("on the stack: %+v, want the Adventure", spell)
	}
}

func TestFreeOrHandOffersATownForItsAdventure(t *testing.T) {
	g, me, id := freeOrHandOver(t, landAdventureFixture(uuid.Nil))
	if me.Hand.Contains(id) || freeCastGrantOn(g, id) == nil {
		t.Fatal("a Town whose Adventure fits the cap went to the hand")
	}
}

func TestFreeOrHandStillSendsACardWithNoFittingFaceToHand(t *testing.T) {
	c := cheapBackMDFCFixture(uuid.Nil)
	c.Faces[1].ManaCost = "{4}{B}"
	g, me, id := freeOrHandOver(t, c)
	if !me.Hand.Contains(id) || freeCastGrantOn(g, id) != nil {
		t.Fatal("a card with no face of mana value 4 or less was offered")
	}
}

func TestFreePlayFromTopOffersAModalDFCWhoseBackFaceFits(t *testing.T) {
	g, me, id := freePlayFromTopOver(t, cheapBackMDFCFixture(uuid.Nil), 5, true)
	if me.Hand.Contains(id) || freeCastGrantOn(g, id) == nil {
		t.Fatal("a modal DFC whose back face fits the cap went to the hand")
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "library", Face: 0}); err == nil {
		t.Fatal("the mana value 7 front face was cast through a cap of 5")
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "library", Face: 1}); err != nil {
		t.Fatalf("cast the back face from the library: %v", err)
	}
	if spell := findCardForTest(g.Stack, id); spell == nil || spell.ActiveFace != 1 || !spell.IsSorcery() {
		t.Fatalf("on the stack: %+v, want the sorcery back face", spell)
	}
}

func TestFreePlayFromTopOffersACreatureWhoseAdventureFits(t *testing.T) {
	g, me, id := freePlayFromTopOver(t, bigCreatureCheapAdventure(uuid.Nil), 5, true)
	if me.Hand.Contains(id) || freeCastGrantOn(g, id) == nil {
		t.Fatal("a creature card whose Adventure fits the cap went to the hand")
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "library", Face: 1}); err != nil {
		t.Fatalf("cast the Adventure from the library: %v", err)
	}
}

// Rashmi's clause is spell-only, so a Town is offered for its Adventure.
func TestFreePlayFromTopSpellOnlyOffersATownsAdventure(t *testing.T) {
	g, me, id := freePlayFromTopOver(t, landAdventureFixture(uuid.Nil), 2, false)
	if me.Hand.Contains(id) || freeCastGrantOn(g, id) == nil {
		t.Fatal("a Town whose Adventure fits went to the hand under a spell-only offer")
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "library", Face: 1}); err != nil {
		t.Fatalf("cast the Town's Adventure: %v", err)
	}
	if spell := findCardForTest(g.Stack, id); spell == nil || !spell.IsSorcery() || spell.IsLand() {
		t.Fatalf("on the stack: %+v, want the sorcery Adventure", spell)
	}
}

// A plain land under a spell-only offer, and a spell over the cap, still go
// to the hand.
func TestFreePlayFromTopStillSendsNoSpellToHand(t *testing.T) {
	land := Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest"}
	g, me, id := freePlayFromTopOver(t, land, 5, false)
	if !me.Hand.Contains(id) || freeCastGrantOn(g, id) != nil {
		t.Fatal("a land was offered as a spell")
	}
	c := cheapBackMDFCFixture(uuid.Nil)
	c.Faces[1].ManaCost = "{5}{B}"
	g, me, id = freePlayFromTopOver(t, c, 5, true)
	if !me.Hand.Contains(id) || freeCastGrantOn(g, id) != nil {
		t.Fatal("a card with no face of mana value 5 or less was offered")
	}
}

// spellFrontLandBackMDFC is the Zendikar Rising shape: a mana value 7
// sorcery front face and a land back face (CR 712.11c).
func spellFrontLandBackMDFC(owner uuid.UUID) Card {
	return mdfcFixture(owner,
		Face{Name: "Fixture Awakening", TypeLine: "Sorcery", ManaCost: "{5}{G}{G}", Colors: []string{"G"}},
		Face{Name: "Fixture Undercrypt", TypeLine: "Land"},
	)
}

// Ziatora's Envoy's "play a land from the top of your library": a modal
// DFC whose back face is a land, and no face of which is a spell that
// fits, is offered as a play of that land face, and accepting it puts the
// back face onto the battlefield as the turn's land play.
func TestFreePlayFromTopOffersAModalDFCsBackFaceLandPlay(t *testing.T) {
	g, me, id := freePlayFromTopOver(t, spellFrontLandBackMDFC(uuid.Nil), 3, true)
	c := mayCastChoiceFor(g, me.ID)
	if c == nil || c.MayCastKeyword != MayCastKeywordPlayOrHand || c.MayCastCard != id {
		t.Fatalf("offer %+v (in hand %v), want the back face's land play", c, me.Hand.Contains(id))
	}
	if freeCastGrantOn(g, id) != nil {
		t.Error("a card with no spell face that fits was also offered as a free cast")
	}
	if err := g.ResolveMayCast(c.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
	land := findCardForTest(g.Battlefield, id)
	if land == nil || land.ActiveFace != 1 || !land.IsLand() || land.Name != "Fixture Undercrypt" {
		t.Fatalf("on the battlefield: %+v, want the land back face", land)
	}
	if got := g.LandsPlayedThisTurnFor(me.ID); got != 1 {
		t.Errorf("lands played this turn = %d, want 1", got)
	}
}

// Declining puts the card into the hand as the card it is, front face up.
func TestFreePlayFromTopDecliningABackFaceLandPutsTheCardInHand(t *testing.T) {
	g, me, id := freePlayFromTopOver(t, spellFrontLandBackMDFC(uuid.Nil), 3, true)
	c := mayCastChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the back face's land play was not offered")
	}
	if err := g.ResolveMayCast(c.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
	card := findCardForTest(me.Hand, id)
	if card == nil || card.ActiveFace != 0 {
		t.Fatalf("in hand: %+v, want the card front face up", card)
	}
	if g.LandsPlayedThisTurnFor(me.ID) != 0 {
		t.Error("a declined land play spent the land drop")
	}
}

// The offer is a restore point: a restored table plays the back face.
func TestFreePlayFromTopBackFaceLandPlaySurvivesARestore(t *testing.T) {
	g, me, id := freePlayFromTopOver(t, spellFrontLandBackMDFC(uuid.Nil), 3, true)
	restored := restoreStrictForTest(t, g)
	c := mayCastChoiceFor(restored, me.ID)
	if c == nil {
		t.Fatal("the restored table lost the offer")
	}
	if err := restored.ResolveMayCast(c.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveMayCast on the restored game: %v", err)
	}
	if land := findCardForTest(restored.Battlefield, id); land == nil || land.ActiveFace != 1 || !land.IsLand() {
		t.Fatalf("on the restored battlefield: %+v, want the land back face", land)
	}
}

// No land drop left, or a spell-only clause (Rashmi): a back face land is
// not a play the player can make, so the card goes to the hand.
func TestFreePlayFromTopBackFaceLandWithNoPlayGoesToHand(t *testing.T) {
	g, me, id := freePlayFromTopOver(t, spellFrontLandBackMDFC(uuid.Nil), 3, false)
	if !me.Hand.Contains(id) || mayCastChoiceFor(g, me.ID) != nil {
		t.Fatal("a spell-only offer offered a land play")
	}

	g = newActiveGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	toMainPhase(t, g)
	c := spellFrontLandBackMDFC(me.ID)
	g.WithWriteLock(func() {
		g.LandsPlayedThisTurn = map[uuid.UUID]int{me.ID: 1}
		me.Library.PushTop(c)
		if err := g.OfferFreePlayFromTopOrHandForEffect(FreePlayFromTop{
			Player: me.ID, Source: uuid.New(), SourceName: "Test Envoy", MaxManaValue: 3, Land: true,
		}); err != nil {
			t.Fatalf("OfferFreePlayFromTopOrHandForEffect: %v", err)
		}
	})
	if !me.Hand.Contains(c.InstanceID) || mayCastChoiceFor(g, me.ID) != nil {
		t.Fatal("a back face land with no land drop left was offered")
	}
}
