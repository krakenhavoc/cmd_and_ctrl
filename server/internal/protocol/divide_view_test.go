package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// divide_view_test.go — #1563, CR 601.2d: a divided clause tells the
// picker the amount to divide, so it can ask for a share per target.
// Absent on every clause that divides nothing.

func TestLegalTargetsCarryTheDivision(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const divided, plain = "test-view-divided", "test-view-plain"

	prev := game.CatalogTargetSpec
	game.CatalogTargetSpec = func(id string) *game.TargetSpec {
		spec := &game.TargetSpec{
			Mode: "creature", Label: "up to two target creatures",
			Zones: []game.ZoneKind{game.ZoneBattlefield},
			CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
				return c.IsCreature()
			},
			Min: 0, Max: 2,
		}
		switch id {
		case divided:
			return spec.Dividing(game.DivideSpec{FromX: true, DoubleFromX: 6})
		case plain:
			return spec
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogTargetSpec = prev })

	bear := game.NewCard("Their Bear", opp.ID)
	bear.TypeLine = "Creature — Bear"
	g.Battlefield.PushTop(bear)
	ids := map[string]uuid.UUID{}
	for _, oracle := range []string{divided, plain} {
		spell := game.NewCard("Smash "+oracle, me.ID)
		spell.TypeLine = "Sorcery"
		spell.OracleID = oracle
		spell.KnownBy = map[uuid.UUID]bool{me.ID: true}
		me.Hand.PushTop(spell)
		ids[oracle] = spell.InstanceID
	}

	view := ViewOfGameFor(g, me.ID.String())
	lt := map[string]*LegalTargetsView{}
	for i := range view.Seats[0].Hand.Cards {
		c := &view.Seats[0].Hand.Cards[i]
		for oracle, id := range ids {
			if c.InstanceID == id.String() {
				lt[oracle] = c.LegalTargets
			}
		}
	}
	if lt[divided] == nil || lt[divided].Divide == nil {
		t.Fatalf("the divided clause ships its amount: %+v", lt[divided])
	}
	if d := lt[divided].Divide; !d.FromX || d.DoubleFromX != 6 || d.Total != 0 {
		t.Errorf("divide = %+v, want from_x doubling from 6", d)
	}
	if lt[plain] == nil || lt[plain].Divide != nil {
		t.Errorf("a clause that divides nothing ships no divide: %+v", lt[plain])
	}
	raw, err := json.Marshal(lt[plain])
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "divide") {
		t.Errorf("`divide` is omitted when absent: %s", raw)
	}
}
