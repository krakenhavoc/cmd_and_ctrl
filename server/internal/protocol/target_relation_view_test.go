package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

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

// #1943: the same-controller relation ships the same pairs shape, so the
// client's second picker narrows to the creatures the first pick's
// controller has, and never repeats the first pick.
func TestSameControllerRelationViewShipsCompatiblePairs(t *testing.T) {
	g := buildActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	a1, a2, b1 := game.NewCard("creature a1", a.ID), game.NewCard("creature a2", a.ID), game.NewCard("creature b1", b.ID)
	for _, c := range []*game.Card{&a1, &a2, &b1} {
		c.TypeLine, c.Power, c.Toughness = "Creature — Test", 1, 1
		g.Battlefield.PushTop(*c)
	}
	creature := func() *game.TargetSpec {
		return &game.TargetSpec{Zones: []game.ZoneKind{game.ZoneBattlefield}, Min: 1, Max: 1,
			CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool { return c.IsCreature() }}
	}
	second := creature()
	second.Distinct = true
	second.Relation = &game.TargetRelation{Slot: 0, Kind: game.TargetControlledBySameController}
	spec := creature().Then(second)
	var clauses []LegalTargetsView
	g.ReadSnapshot(func() { clauses = viewOfClauses(g, game.SourceChooser(a.ID), spec) })
	if len(clauses) != 2 || clauses[1].Relation == nil {
		t.Fatalf("clauses: %+v", clauses)
	}
	pairs := clauses[1].Relation.Pairs
	if !slices.Equal(pairs[a1.InstanceID.String()], []string{a2.InstanceID.String()}) ||
		!slices.Equal(pairs[a2.InstanceID.String()], []string{a1.InstanceID.String()}) ||
		len(pairs[b1.InstanceID.String()]) != 0 {
		t.Fatalf("pairs: %v", pairs)
	}
}
