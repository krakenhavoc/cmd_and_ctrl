package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ADR 0189 decision 4, wire half: a manifested card whose card also
// prints megamorph lists BOTH turn_face_up rows to its controller, each
// with its route, and the opponent sees neither.
func TestManifestedMegamorphListsBothRoutesToTheControllerOnly(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	prev := game.CatalogAlternativeCosts
	game.CatalogAlternativeCosts = func(oracleID string) []game.AlternativeCost {
		if oracleID != morphViewOracle {
			return nil
		}
		return []game.AlternativeCost{{
			Key: "megamorph", Label: "Megamorph — cast face down {3}", ManaCost: "{3}",
			FaceDown: &game.FaceDownCast{Kind: game.FaceDownMorphed, FaceUpCost: "{1}{G}", FaceUpCounter: true},
		}}
	}
	t.Cleanup(func() { game.CatalogAlternativeCosts = prev })

	c := game.NewCard("Den Protector", me.ID)
	c.OracleID = morphViewOracle
	c.TypeLine = "Creature — Human Warrior"
	c.ManaCost = "{1}{G}"
	id := c.InstanceID
	g.WithWriteLock(func() {
		me.Library.PushTop(c)
		if _, err := g.ManifestForEffect(me.ID); err != nil {
			t.Fatalf("manifest: %v", err)
		}
	})

	find := func(viewer uuid.UUID) CardView {
		t.Helper()
		for _, cv := range ViewOfGameFor(g, viewer.String()).Battlefield.Cards {
			if cv.InstanceID == id.String() {
				return cv
			}
		}
		t.Fatalf("face-down permanent missing from %s's view", viewer)
		return CardView{}
	}

	rows := find(me.ID).SpecialActions
	if len(rows) != 2 {
		t.Fatalf("controller: special_actions = %+v, want two turn_face_up rows", rows)
	}
	wantRoutes := []string{"morph", "mana_cost"}
	wantLabels := []string{
		"Turn face up {1}{G} (megamorph, +1/+1 counter)",
		"Turn face up {1}{G} (mana cost)",
	}
	for i, row := range rows {
		if row.Kind != "turn_face_up" || row.Route != wantRoutes[i] || row.Label != wantLabels[i] || row.Cost != "{1}{G}" {
			t.Errorf("row %d = %+v, want route %q label %q cost {1}{G}", i, row, wantRoutes[i], wantLabels[i])
		}
	}
	if got := find(opp.ID).SpecialActions; len(got) != 0 {
		t.Errorf("opponent: special_actions = %+v, want none", got)
	}
}
