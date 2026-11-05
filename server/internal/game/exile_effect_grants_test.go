package game

import (
	"testing"

	"github.com/google/uuid"
)

// exile_effect_grants_test.go — ADR 0219: an ability a resolved effect
// gives an exiled card is plain data on that object, survives a restore,
// and ends at its next zone change (CR 400.7).

func exileGrantFixture(t *testing.T) (*Game, *Player, uuid.UUID) {
	t.Helper()
	g := newActiveGame(t)
	me := g.Seats[0]
	id := uuid.New()
	g.Exile.PushTop(Card{InstanceID: id, Name: "Exiled", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	return g, me, id
}

func TestAnExileGrantLandsOnlyOnACardInExile(t *testing.T) {
	g, me, id := exileGrantFixture(t)
	hand := uuid.New()
	me.Hand.PushTop(Card{InstanceID: hand, Name: "In Hand", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})

	var onHand, onExile, empty bool
	g.WithWriteLock(func() {
		onHand = g.GrantAbilityToExiledCardForEffect(hand, uuid.Nil, me.ID, "test/bundle")
		empty = g.GrantAbilityToExiledCardForEffect(id, uuid.Nil, me.ID, "")
		onExile = g.GrantAbilityToExiledCardForEffect(id, uuid.Nil, me.ID, "test/bundle")
	})
	if onHand || empty || !onExile {
		t.Fatalf("hand %v, empty key %v, exile %v; want false, false, true", onHand, empty, onExile)
	}
	c := exiledCardByIDLocked(g, id)
	if len(c.ExileGrants) != 1 || c.ExileGrants[0].Key != "grant:test/bundle" || c.ExileGrants[0].Activator != me.ID {
		t.Fatalf("ExileGrants = %+v", c.ExileGrants)
	}
	if got := abilityGrantsOf(c); len(got) != 1 {
		t.Errorf("abilityGrantsOf = %+v, want the effect grant", got)
	}
}

func TestAnExileGrantSurvivesARestore(t *testing.T) {
	g, me, id := exileGrantFixture(t)
	src := uuid.New()
	g.WithWriteLock(func() { g.GrantAbilityToExiledCardForEffect(id, src, me.ID, "test/bundle") })

	_, restored := roundTrip(t, g)
	c := exiledCardByIDLocked(restored, id)
	if c == nil || len(c.ExileGrants) != 1 {
		t.Fatalf("restored card %+v lost its grant", c)
	}
	if gr := c.ExileGrants[0]; gr.Key != "grant:test/bundle" || gr.Source != src || gr.Activator != me.ID {
		t.Errorf("restored grant = %+v", gr)
	}
}

func TestAnExileGrantEndsWhenTheCardChangesZone(t *testing.T) {
	g, me, id := exileGrantFixture(t)
	g.WithWriteLock(func() { g.GrantAbilityToExiledCardForEffect(id, uuid.Nil, me.ID, "test/bundle") })

	moved, err := MoveCard(g.Exile, me.Graveyard, id)
	if err != nil {
		t.Fatalf("MoveCard: %v", err)
	}
	if len(moved.ExileGrants) != 0 {
		t.Fatalf("the card kept %+v across a zone change (CR 400.7)", moved.ExileGrants)
	}
}

func TestACloneDoesNotShareExileGrants(t *testing.T) {
	g, me, id := exileGrantFixture(t)
	g.WithWriteLock(func() { g.GrantAbilityToExiledCardForEffect(id, uuid.Nil, me.ID, "test/bundle") })
	c := *exiledCardByIDLocked(g, id)
	deepenCard(&c)
	c.ExileGrants[0].Key = "grant:changed"
	if exiledCardByIDLocked(g, id).ExileGrants[0].Key != "grant:test/bundle" {
		t.Error("a deepened copy shares its ExileGrants with the original")
	}
}
