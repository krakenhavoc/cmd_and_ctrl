package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// variable_sacrifice_view_test.go — ADR 0100 §4: a variable sacrifice
// as a cast's additional cost ships through the existing
// `additional_cost.sacrifice_options`. "Any number" is min 0 / max 0
// (an open count from zero); "sacrifice X" sets `count_from_x`, so the
// client sends the number picked as the x_value. No new field.

const (
	anyNumberViewOracle = "test-any-number-view"
	sacXViewOracle      = "test-sacrifice-x-view"
)

func installVariableSacrificeForView(t *testing.T) {
	t.Helper()
	prev := game.CatalogAdditionalCost
	creatures := func() *game.TargetSpec {
		return &game.TargetSpec{
			Mode: "permanent", Label: "creatures", Zones: []game.ZoneKind{game.ZoneBattlefield},
			CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool { return c.IsCreature() },
		}
	}
	anyNumber := &game.AdditionalCost{Label: "Sacrifice any number of creatures", Sacrifice: creatures()}
	fromX := creatures()
	fromX.CountFromX = true
	sacX := &game.AdditionalCost{Label: "Sacrifice X creatures", Sacrifice: fromX}
	game.CatalogAdditionalCost = func(key string) *game.AdditionalCost {
		switch key {
		case anyNumberViewOracle:
			return anyNumber
		case sacXViewOracle:
			return sacX
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogAdditionalCost = prev })
}

func TestVariableSacrificeCostIsProjected(t *testing.T) {
	installVariableSacrificeForView(t)
	g := buildActiveGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	me := g.Seats[0]
	me.Hand.Cards = nil
	bear := game.NewCard("Bear", me.ID)
	bear.TypeLine = "Creature — Bear"
	bear.Controller = me.ID
	g.Battlefield.PushTop(bear)

	cards := map[string]game.Card{}
	for _, oracle := range []string{anyNumberViewOracle, sacXViewOracle} {
		spell := game.NewCard("Test "+oracle, me.ID)
		spell.TypeLine = "Sorcery"
		spell.ManaCost = "{3}{B}"
		spell.OracleID = oracle
		spell.KnownBy = map[uuid.UUID]bool{me.ID: true}
		me.Hand.PushTop(spell)
		cards[oracle] = spell
	}
	v := FilterViewFor(ViewOfGame(g), me.ID.String())

	open := handCardView(t, v, 0, cards[anyNumberViewOracle].InstanceID).AdditionalCost
	if open == nil || open.SacrificeOptions == nil {
		t.Fatalf("any-number additional_cost = %+v, want sacrifice options", open)
	}
	if o := open.SacrificeOptions; o.Min != 0 || o.Max != 0 || o.CountFromX || len(o.Cards) != 1 {
		t.Errorf("any-number options = %+v, want min 0 / max 0 (open), no count_from_x, the Bear", o)
	}

	x := handCardView(t, v, 0, cards[sacXViewOracle].InstanceID).AdditionalCost
	if x == nil || x.SacrificeOptions == nil || !x.SacrificeOptions.CountFromX || x.DemandsX {
		t.Fatalf("sacrifice-X additional_cost = %+v, want count_from_x options and no demands_x (the count is the X)", x)
	}
}
