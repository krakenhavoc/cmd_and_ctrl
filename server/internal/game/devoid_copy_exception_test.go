package game

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// TestColourExceptionDropsDevoidOnBothRoads — #2322 / CR 707.9d. A copy
// exception that provides a colour leaves the copy without the
// colour-defining ability: no devoid in its effective abilities,
// whether the keyword rode Card.Keywords (import road) or the copied
// oracle ID's catalog entry (catalog road). The mark is a copiable
// value, and a plain copy keeps devoid.
func TestColourExceptionDropsDevoidOnBothRoads(t *testing.T) {
	stubPrintedKeywords(t, map[string][]string{"devoid-catalog-probe": {KeywordDevoid, "haste"}})
	for _, tc := range []struct {
		name string
		mk   func() Card
	}{
		{"catalog road", func() Card {
			c := NewCard("Catalog Drone", uuid.New())
			c.OracleID = "devoid-catalog-probe"
			c.TypeLine = "Creature — Eldrazi Drone"
			c.ManaCost = "{1}{R}"
			return c
		}},
		{"import road", func() Card {
			c := NewCard("Import Drone", uuid.New())
			c.TypeLine = "Creature — Eldrazi Drone"
			c.ManaCost = "{1}{R}"
			c.Keywords = []string{KeywordDevoid, "haste"}
			return c
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			plain := tc.mk()
			if !containsKeyword(plain.Effective().Abilities, KeywordDevoid) {
				t.Fatalf("a plain copy must keep devoid: %v", plain.Effective().Abilities)
			}
			c := tc.mk()
			c.SetCopyExceptionColors("B")
			ab := c.Effective().Abilities
			if containsKeyword(ab, KeywordDevoid) || !containsKeyword(ab, "haste") {
				t.Errorf("abilities = %v, want haste and no devoid", ab)
			}
			if got := c.EffectiveColors(); !reflect.DeepEqual(got, []string{"B"}) {
				t.Errorf("colours = %v, want [B]", got)
			}
			// A copy of the copy inherits the drop (copiable value).
			var cc Card
			cc.applyCopy(CopiableValuesOf(c), Card{})
			if !cc.ColorCDADropped || containsKeyword(cc.Effective().Abilities, KeywordDevoid) {
				t.Errorf("a copy of the copy lost the drop: %v", cc.Effective().Abilities)
			}
		})
	}
}

// TestColourCDADroppedSurvivesSnapshotAndCloneAndClearsOnLeave is the
// per-instance lifecycle of the mark (#2322).
func TestColourCDADroppedSurvivesSnapshotAndCloneAndClearsOnLeave(t *testing.T) {
	stubPrintedKeywords(t, map[string][]string{"devoid-catalog-probe": {KeywordDevoid}})
	g := newRestorableGame(t)
	p := g.Seats[0]
	c := NewCard("Catalog Drone", p.ID)
	c.OracleID = "devoid-catalog-probe"
	c.TypeLine = "Creature — Eldrazi Drone"
	c.ManaCost = "{1}{R}"
	c.Power, c.Toughness = 2, 2
	c.SetCopyExceptionColors("B")
	id := pushTypedTestCard(g, c)

	_, restored := roundTrip(t, g)
	for name, gg := range map[string]*Game{"restored": restored, "clone": g.Clone()} {
		got := layeredBattlefieldCard(t, gg, id)
		if !got.ColorCDADropped || containsKeyword(got.Effective().Abilities, KeywordDevoid) {
			t.Errorf("%s: mark %v, abilities %v", name, got.ColorCDADropped, got.Effective().Abilities)
		}
	}

	if err := g.DestroyPermanentForEffect(id); err != nil {
		t.Fatalf("DestroyPermanentForEffect: %v", err)
	}
	gone, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatal("card vanished")
	}
	if gone.ColorCDADropped {
		t.Error("the mark must clear when the object leaves the battlefield")
	}
}
