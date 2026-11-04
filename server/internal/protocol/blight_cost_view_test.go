package protocol

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blight_cost_view_test.go — #4684: the wire half of "Blight N" as part of
// an activated ability's cost (Gristle Glutton). The row carries `blight`
// (N) and `blight_options`, the engine's own candidate walk: the creatures
// the activator controls, the source included, and no one else's.

const gristleGluttonOracleForView = "b7163800-41fc-410a-9650-50ae90842170"

func TestAbilityViewCarriesBlightAndItsOptions(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	glutton := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: glutton, Name: "Gristle Glutton", TypeLine: "Creature — Goblin Scout",
		OracleID: gristleGluttonOracleForView, Power: 1, Toughness: 3, Owner: me, Controller: me,
	})
	bear := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me, Controller: me})
	g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp, Controller: opp})
	g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact",
		Owner: me, Controller: me})
	g.BumpLayerVersionForTest()

	ab := vehicleView(t, g, glutton).ActivatedAbilities[0]
	if ab.Blight != 1 {
		t.Errorf("blight = %d, want 1", ab.Blight)
	}
	if ab.BlightOptions == nil {
		t.Fatal("blight_options absent on an ability that prints the clause")
	}
	if ab.BlightOptions.Min != 1 || ab.BlightOptions.Max != 1 {
		t.Errorf("bounds %d/%d, want 1/1", ab.BlightOptions.Min, ab.BlightOptions.Max)
	}
	got := append([]string(nil), ab.BlightOptions.Cards...)
	sort.Strings(got)
	want := []string{glutton.String(), bear.String()}
	sort.Strings(want)
	if !sameStrings(got, want) {
		t.Errorf("blight_options = %v, want the two creatures I control", ab.BlightOptions.Cards)
	}
	raw, err := json.Marshal(ab)
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]any
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"blight", "blight_options"} {
		if _, ok := wire[k]; !ok {
			t.Errorf("wire row has no %q", k)
		}
	}
}

// An ability without the clause ships neither field.
func TestAbilityViewWithoutBlightShipsNoBlight(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0].ID
	cap := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: cap, Name: "Captain America, First Avenger", TypeLine: "Legendary Creature — Human Soldier Hero",
		OracleID: capFirstAvengerOracleForView, Power: 4, Toughness: 4, Owner: me, Controller: me,
	})
	g.BumpLayerVersionForTest()
	ab := vehicleView(t, g, cap).ActivatedAbilities[0]
	if ab.Blight != 0 || ab.BlightOptions != nil {
		t.Errorf("blight %d / options %+v on a row without the clause", ab.Blight, ab.BlightOptions)
	}
}
