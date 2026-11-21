package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// seriema_etb_test.go — #337 reported that The Seriema's enters
// trigger did not fire. At the time the card had no catalog entry at
// all; this pins the trigger through a real cast now that it has one.
// The card itself still carries its Station caveat (station_test.go).

// TestSeriemaEntersAndTutorsALegendaryCreature is #337's report as a
// test: The Seriema is CAST, resolves, and its enters trigger searches
// for a legendary creature card and puts it into the hand. Nonlegendary
// creatures and legendary noncreatures are not offered.
func TestSeriemaEntersAndTutorsALegendaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear"},
		game.Card{Name: "Sol Talisman", TypeLine: "Legendary Artifact"},
		game.Card{Name: "Aang, Swift Savior", TypeLine: "Legendary Creature — Human Avatar Ally"},
		game.Card{Name: "Katara, Bending Prodigy", TypeLine: "Legendary Creature — Human Warrior Ally"},
	)
	effectstest.CastCatalogSpell(t, g, "The Seriema", "Legendary Artifact — Spacecraft", effectstest.TheSeriemaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("The Seriema's enters trigger did not search (#337)")
	}
	for _, name := range []string{"Grizzly Bears", "Sol Talisman"} {
		if effectstest.SearchOptionNamed(g, c, name) != uuid.Nil {
			t.Errorf("%s is offered, but it is not a legendary creature card", name)
		}
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Katara, Bending Prodigy")
	if !effectstest.B02bHandHasNamed(me, "Katara, Bending Prodigy") {
		t.Error("the legendary creature did not reach the hand")
	}
}
