package protocol

import (
	"slices"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestTargetRelationViewShipsCompatiblePairs(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a, b := game.NewCard("artifact a", me.ID), game.NewCard("artifact b", opp.ID)
	x, y := game.NewCard("grave a", me.ID), game.NewCard("grave b", opp.ID)
	a.TypeLine, b.TypeLine, x.TypeLine, y.TypeLine = "Artifact", "Artifact", "Artifact", "Artifact"
	g.Battlefield.PushTop(a)
	g.Battlefield.PushTop(b)
	me.Graveyard.PushTop(x)
	opp.Graveyard.PushTop(y)
	spec := (&game.TargetSpec{Zones: []game.ZoneKind{game.ZoneBattlefield}, Min: 1, Max: 1}).Then(&game.TargetSpec{
		Zones: []game.ZoneKind{game.ZoneGraveyard}, Min: 1, Max: 1,
		Relation: &game.TargetRelation{Slot: 0, Kind: game.TargetInControllersGraveyard},
	})
	var clauses []LegalTargetsView
	g.ReadSnapshot(func() { clauses = viewOfClauses(g, game.SourceChooser(me.ID), spec) })
	if len(clauses) != 2 || clauses[1].Relation == nil || clauses[1].Relation.Slot != 0 {
		t.Fatalf("clauses: %+v", clauses)
	}
	pairs := clauses[1].Relation.Pairs
	if !slices.Equal(pairs[a.InstanceID.String()], []string{x.InstanceID.String()}) || !slices.Equal(pairs[b.InstanceID.String()], []string{y.InstanceID.String()}) {
		t.Fatalf("pairs: %v", pairs)
	}
}
