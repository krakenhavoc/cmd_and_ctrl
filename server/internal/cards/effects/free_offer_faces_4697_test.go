package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// free_offer_faces_4697_test.go — #4697: Beseech the Mirror's "if that
// spell's mana value is 4 or less" and Ziatora's Envoy's and Rashmi's "a
// spell with mana value N or less" are about the spell, so the card is
// offered when some face a cast may choose is a spell that fits
// (CR 601.3e, 712.11c, 715.3a), and the cast is that face.

// cheapBackMDFC is a modal DFC whose front face (mana value 7) fits no cap
// and whose back face (a mana value 2 sorcery) fits every cap here.
func cheapBackMDFC(owner uuid.UUID) game.Card {
	return mdfcCard(owner, "", game.LayoutModalDFC,
		game.Face{Name: "Fixture Colossus", TypeLine: "Creature — Giant", ManaCost: "{5}{B}{B}",
			Colors: []string{"B"}, Power: 7, Toughness: 7},
		game.Face{Name: "Fixture Rite", TypeLine: "Sorcery", ManaCost: "{1}{B}", Colors: []string{"B"}},
	)
}

// beseechFaceTable is mcBeseechTable with a chosen card to find.
func beseechFaceTable(t *testing.T, wanted game.Card) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wanted.Owner, wanted.Controller = me.ID, me.ID
	g.WithWriteLock(func() {
		pushLibraryCardForTest(me, wanted)
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Sorcery", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	})
	return g, me, wanted.InstanceID
}

// bargainedBeseechFinds casts Beseech the Mirror bargained and finds `id`.
func bargainedBeseechFinds(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID) {
	t.Helper()
	token := pushTokens(g, me.ID, RedGoblinToken(), 1)[0]
	if _, err := castWithOptionalCosts(t, g, "Beseech the Mirror", "Sorcery", mcBeseechOracle, nil, []int{0}, []uuid.UUID{token}); err != nil {
		t.Fatalf("bargained cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, id)
}

func TestBeseechTheMirrorOffersAModalDFCWhoseBackFaceFits(t *testing.T) {
	g, me, id := beseechFaceTable(t, cheapBackMDFC(uuid.Nil))
	bargainedBeseechFinds(t, g, me, id)

	c := keywordCastPromptFor(g, me.ID)
	if c == nil || keywordCastCard(c) != id {
		t.Fatalf("the modal DFC went to the hand (%v) instead of being offered: %+v", me.Hand.Contains(id), c)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile", Face: 0}); err == nil {
		t.Fatal("the mana value 7 front face was cast through a cap of 4")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile", Face: 1}); err != nil {
		t.Fatalf("cast the back face free: %v", err)
	}
	if spell, ok := stackCardOf(g, id); !ok || spell.ActiveFace != 1 || !spell.IsSorcery() {
		t.Fatalf("on the stack: %+v, want the sorcery back face", spell)
	}
}

func TestBeseechTheMirrorOffersACreatureWhoseAdventureFits(t *testing.T) {
	g, me, id := beseechFaceTable(t, expensiveFrontCheapBack(uuid.Nil))
	bargainedBeseechFinds(t, g, me, id)

	if c := keywordCastPromptFor(g, me.ID); c == nil || keywordCastCard(c) != id {
		t.Fatalf("the adventurer card was not offered for its Adventure: %+v", c)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile", Face: 1}); err != nil {
		t.Fatalf("cast the Adventure free: %v", err)
	}
	if spell, ok := stackCardOf(g, id); !ok || spell.ActiveFace != 1 || spell.IsCreature() {
		t.Fatalf("on the stack: %+v, want the Adventure", spell)
	}
}

// A card no face of which fits still goes straight to the hand.
func TestBeseechTheMirrorStillSendsACardWithNoFittingFaceToHand(t *testing.T) {
	big := cheapBackMDFC(uuid.Nil)
	big.Faces[1].ManaCost = "{4}{B}"
	g, me, id := beseechFaceTable(t, big)
	bargainedBeseechFinds(t, g, me, id)
	if !me.Hand.Contains(id) || keywordCastPromptFor(g, me.ID) != nil {
		t.Fatal("a card with no face of mana value 4 or less was offered")
	}
}

// pushTop puts a built card on top of the Envoy controller's library.
func (tb *envoyTable) pushTop(c game.Card) uuid.UUID {
	c.Owner, c.Controller = tb.me.ID, tb.me.ID
	tb.g.WithWriteLock(func() { tb.me.Library.PushTop(c) })
	return c.InstanceID
}

func TestZiatorasEnvoyCastsAModalDFCsBackFaceFromTheTop(t *testing.T) {
	tb := newEnvoyTable(t)
	id := tb.pushTop(cheapBackMDFC(uuid.Nil))
	tb.attack(t)
	c := tb.mayCast()
	if c == nil || c.Kind != game.PendingChoiceResolutionCast || keywordCastCard(c) != id {
		t.Fatalf("the modal DFC whose back face fits was not offered: %+v (in hand %v)", c, tb.me.Hand.Contains(id))
	}
	if err := tb.g.CastSpell(tb.me.ID, id, game.CastSpellParams{FromZone: "library", Face: 0}); err == nil {
		t.Fatal("the mana value 7 front face was cast after 5 damage")
	}
	if err := tb.g.CastSpell(tb.me.ID, id, game.CastSpellParams{FromZone: "library", Face: 1}); err != nil {
		t.Fatalf("cast the back face from the library: %v", err)
	}
	if spell, ok := stackCardOf(tb.g, id); !ok || spell.ActiveFace != 1 || !spell.IsSorcery() {
		t.Fatalf("on the stack: %+v, want the sorcery back face", spell)
	}
}

func TestZiatorasEnvoyCastsAnAdventureThatFits(t *testing.T) {
	tb := newEnvoyTable(t)
	id := tb.pushTop(expensiveFrontCheapBack(uuid.Nil))
	tb.attack(t)
	if c := tb.mayCast(); c == nil || keywordCastCard(c) != id {
		t.Fatalf("the adventurer card was not offered for its Adventure: %+v", c)
	}
	if err := tb.g.CastSpell(tb.me.ID, id, game.CastSpellParams{FromZone: "library", Face: 1}); err != nil {
		t.Fatalf("cast the Adventure from the library: %v", err)
	}
	if spell, ok := stackCardOf(tb.g, id); !ok || spell.ActiveFace != 1 || spell.IsCreature() {
		t.Fatalf("on the stack: %+v, want the Adventure", spell)
	}
}

// A Town whose land cannot be played (the land drop is spent) is still a
// spell to cast: its sorcery Adventure.
func TestZiatorasEnvoyCastsATownsAdventureWhenTheLandCannotBePlayed(t *testing.T) {
	tb := newEnvoyTable(t)
	id := tb.pushTop(landAdventureCard(uuid.Nil, "", "Fixture Town", "Fixture Overture", "{1}{U}", []string{"U"}))
	tb.g.WithWriteLock(func() { tb.g.LandsPlayedThisTurn = map[uuid.UUID]int{tb.me.ID: 1} })
	tb.attack(t)
	c := tb.mayCast()
	if c == nil || c.Kind != game.PendingChoiceResolutionCast || keywordCastCard(c) != id {
		t.Fatalf("the Town's Adventure was not offered: %+v (in hand %v)", c, tb.me.Hand.Contains(id))
	}
	if err := tb.g.CastSpell(tb.me.ID, id, game.CastSpellParams{FromZone: "library", Face: 1}); err != nil {
		t.Fatalf("cast the Town's Adventure: %v", err)
	}
	if spell, ok := stackCardOf(tb.g, id); !ok || !spell.IsSorcery() || spell.IsLand() {
		t.Fatalf("on the stack: %+v, want the sorcery Adventure", spell)
	}
}

// With the land drop still there, a Town is the land play it was.
func TestZiatorasEnvoyStillOffersATownAsALandPlay(t *testing.T) {
	tb := newEnvoyTable(t)
	id := tb.pushTop(landAdventureCard(uuid.Nil, "", "Fixture Town", "Fixture Overture", "{1}{U}", []string{"U"}))
	tb.attack(t)
	c := tb.mayCast()
	if c == nil || c.MayCastKeyword != game.MayCastKeywordPlayOrHand || c.MayCastCard != id {
		t.Fatalf("offer %+v, want the land play", c)
	}
}

// Rashmi's clause is spell-only: a revealed Town is offered for its
// Adventure.
func TestRashmiOffersARevealedTownsAdventure(t *testing.T) {
	rt := newRashmiTable(t)
	town := landAdventureCard(rt.me.ID, "", "Fixture Town", "Fixture Overture", "{1}{U}", []string{"U"})
	rt.g.WithWriteLock(func() { rt.me.Library.PushTop(town) })
	rt.castFirst(t, "{3}")
	if c := rt.mayCast(); c == nil || keywordCastCard(c) != town.InstanceID {
		t.Fatalf("a Town whose mana value 2 Adventure is lesser than 3 was not offered: %+v", c)
	}
}
