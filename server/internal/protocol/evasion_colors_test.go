package protocol

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestCardViewPreservesNegativePowerForComparisons(t *testing.T) {
	for _, power := range []int{-2, 0, 2} {
		c := game.Card{InstanceID: uuid.New(), Power: power + 1, Toughness: 4, Counters: map[string]int{"-1/-1": 1}}
		view := viewOfCard(c)
		if view.Power != max(0, power) || view.NegativePower != min(0, power) {
			t.Errorf("signed power %d projected as power=%d negative_power=%d", power, view.Power, view.NegativePower)
		}
		if got := redactCardForViewer(view, false).NegativePower; got != 0 {
			t.Errorf("unknown card leaked negative_power=%d", got)
		}
	}
}

func TestCardViewColorsFollowLayersIncludingBecomingColorless(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Red creature", TypeLine: "Creature — Test", Power: 2, Toughness: 2,
		ManaCost: "{R}", Colors: []string{"R"}, Owner: owner, Controller: owner,
	})
	if got := cardViewByID(t, g, id).Colors; !reflect.DeepEqual(got, []string{"R"}) {
		t.Fatalf("printed colors = %v", got)
	}
	old := game.CatalogStaticAbilities
	defer func() { game.CatalogStaticAbilities = old }()
	colors := []string{"B", "U"}
	game.CatalogStaticAbilities = func(key string) []game.StaticAbility {
		if key != "color-change" {
			return nil
		}
		return []game.StaticAbility{{
			Layer:     game.Layer5Color,
			AppliesTo: func(c *game.Card, _ *game.Game, _ *game.Card) bool { return c.InstanceID == id },
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Colors = append([]string(nil), colors...)
			},
		}}
	}
	g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), OracleID: "color-change", Name: "Color effect", TypeLine: "Enchantment", Owner: owner, Controller: owner})
	g.BumpLayerVersionForTest()
	before := cardViewByID(t, g, id)
	if !reflect.DeepEqual(before.Colors, colors) {
		t.Fatalf("effective colors = %v, want %v", before.Colors, colors)
	}
	colors = nil
	g.BumpLayerVersionForTest()
	if got := cardViewByID(t, g, id).Colors; len(got) != 0 {
		t.Fatalf("colorless permanent fell back to printed mana colors: %v", got)
	}
	if !reflect.DeepEqual(before.Colors, []string{"B", "U"}) {
		t.Fatal("later layer pass mutated an already captured view")
	}
}
