package game

import (
	"testing"

	"github.com/google/uuid"
)

// exiled_with_face_down_test.go — #3787: a face-down exile that is also
// "exiled with" its source, and a cast permission's own "cost {N} less".

func exiledWithSetup(t *testing.T, g *Game, hidden Card) (source PermissionCardRef, cardID uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	src := pushTypedTestCard(g, Card{Name: "Test Box", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	hidden.InstanceID = uuid.New()
	hidden.Owner = me.ID
	me.Library.PushTop(hidden)
	g.WithWriteLock(func() {
		source = ObjectRefOf(*findBattlefieldCard(g, src))
		if paused, err := g.ExileHiddenLinkedForEffect(source, me.ID, hidden.InstanceID); paused || err != nil {
			t.Fatalf("ExileHiddenLinkedForEffect = %v, %v", paused, err)
		}
	})
	return source, hidden.InstanceID
}

// The card is hidden exactly as hideaway hides it AND linked as an
// ability's exile is: both the viewer rule and ExiledWithForEffect see it.
func TestHiddenLinkedExileIsFaceDownAndExiledWithItsSource(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	source, id := exiledWithSetup(t, g, Card{Name: "Boxed", TypeLine: "Instant", ManaCost: "{R}"})

	c, ok := hiddenZoneCard(g.Exile, id)
	if !ok {
		t.Fatal("the card is not in exile")
	}
	if c.FaceDownKind != FaceDownHidden || c.HiddenBy != source || c.ExiledWith != source {
		t.Errorf("kind %q hiddenBy %+v exiledWith %+v, want hidden and linked to %+v", c.FaceDownKind, c.HiddenBy, c.ExiledWith, source)
	}
	if !c.IsKnownTo(me.ID) || c.IsKnownTo(opp.ID) {
		t.Errorf("knowers: me=%v opp=%v, want only the source's controller", c.IsKnownTo(me.ID), c.IsKnownTo(opp.ID))
	}
	var linked []*Card
	g.ReadSnapshot(func() { linked = g.ExiledWithForEffect(source) })
	if len(linked) != 1 || linked[0].InstanceID != id {
		t.Errorf("ExiledWithForEffect = %d cards, want the one", len(linked))
	}
}

// Hideaway's own exile does NOT carry the "exiled with" link, so a
// standing permission over "cards exiled with" never reaches it.
func TestHideawayExileCarriesNoExiledWithLink(t *testing.T) {
	g := newActiveGame(t)
	_, id := hideawaySetup(t, g, Card{Name: "Hidden Bolt", TypeLine: "Instant", ManaCost: "{R}"})
	c, _ := hiddenZoneCard(g.Exile, id)
	if c.ExiledWith.ID != uuid.Nil {
		t.Errorf("hideaway's card carries ExiledWith %+v", c.ExiledWith)
	}
}

// The link and the permission's reduction survive a snapshot, and a
// restored table prices the cast the same way.
func TestCostLessPermissionSurvivesASnapshotAndPricesTheCast(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	source, id := exiledWithSetup(t, g, Card{Name: "Boxed Bear", TypeLine: "Creature — Bear", ManaCost: "{3}{G}"})
	g.WithWriteLock(func() {
		g.GrantCastPermissionForEffect(CastPermission{
			Player:   me.ID,
			Zone:     ZoneExile,
			Scope:    ScopeStanding,
			Filter:   PermissionFilter{ExiledWithSource: true, ExiledWith: source, OwnedByHolder: true},
			CostLess: 2,
		})
	})

	price := func(g *Game) string {
		var out string
		g.ReadSnapshot(func() {
			c, ok := hiddenZoneCard(g.Exile, id)
			if !ok {
				t.Fatal("card not in exile")
			}
			p, err := g.PriceCastForEffect(me.ID, c, CastSpellParams{FromZone: "exile"})
			if err != nil {
				t.Fatal(err)
			}
			out = p.Total.String()
		})
		return out
	}
	if got := price(g); got != "{1}{G}" {
		t.Fatalf("price = %s, want {1}{G}", got)
	}
	_, restored := roundTrip(t, g)
	if got := price(restored); got != "{1}{G}" {
		t.Errorf("restored price = %s, want {1}{G}", got)
	}
}
