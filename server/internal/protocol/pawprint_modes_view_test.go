package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pawprint_modes_view_test.go — CR 700.2i, #4786: a pawprint spell's
// `modes.pawprints` is its budget and each option's `pawprints` its
// printed {P}, so the picker can bound a selection by the {P} it spends.
// Both are printed text and survive on the public copy.

func TestModeSpecViewCarriesThePawprintBudgetAndEachBulletsPawprints(t *testing.T) {
	const oracle = "test-view-pawprints"
	g := buildActiveGame(t)
	me := g.Seats[0]
	prev := game.CatalogModeSpec
	spec := &game.ModeSpec{
		Prompt: "Choose up to five {P} worth of modes", Min: 0, Max: 5, Repeatable: true, Pawprints: 5,
		Options: []game.ModeOption{{Label: "A.", Pawprints: 1}, {Label: "B.", Pawprints: 2}, {Label: "C.", Pawprints: 3}},
	}
	game.CatalogModeSpec = func(id string) *game.ModeSpec {
		if id != oracle {
			return nil
		}
		return spec
	}
	t.Cleanup(func() { game.CatalogModeSpec = prev })

	spell := game.NewCard("Pawprint Test", me.ID)
	spell.TypeLine = "Sorcery"
	spell.OracleID = oracle
	spell.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Hand.PushTop(spell)

	m := handModes(t, g, me, 0, spell.InstanceID)
	if m == nil || m.Pawprints != 5 || m.Max != 5 || !m.Repeatable {
		t.Fatalf("modes = %+v, want a repeatable spec with a budget of 5 {P}", m)
	}
	for i, want := range []int{1, 2, 3} {
		if m.Options[i].Pawprints != want {
			t.Errorf("option %d pawprints = %d, want %d", i, m.Options[i].Pawprints, want)
		}
	}
	pub := publicModeSpec(m)
	if pub.Pawprints != 5 || pub.Options[2].Pawprints != 3 {
		t.Errorf("the printed {P} should survive on the public copy: %+v", pub)
	}
}
