package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// #2176: an Adventure whose main half is a LAND. From exile, after the
// Adventure resolves, the bot is offered the land play (CR 715.3d
// "play") while the turn's land drop is open, and nothing once it is
// spent (CR 305.2).

// landAdvLandMoves is castMovesFor for the land-play kind.
func landAdvLandMoves(moves []legal.Move, source uuid.UUID) []legal.Move {
	var out []legal.Move
	for _, m := range moves {
		if m.Kind == legal.KindLand && m.Source == source {
			out = append(out, m)
		}
	}
	return out
}

func landAdventureInHand(p *game.Player) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   "land-adventure-enum-oracle",
		Layout:     game.LayoutAdventure,
		Owner:      p.ID,
		Controller: p.ID,
		Faces: []game.Face{
			{Name: "Enum Town", TypeLine: "Land — Town"},
			{Name: "Enum Overture", TypeLine: "Sorcery — Adventure", ManaCost: "{0}"},
		},
	}
	c.SetFace(0)
	p.Hand.PushTop(c)
	return c.InstanceID
}

func castLandAdventureToExile(t *testing.T, g *game.Game, seat *game.Player) uuid.UUID {
	t.Helper()
	id := landAdventureInHand(seat)
	if err := g.CastSpell(seat.ID, id, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast the Adventure half: %v", err)
	}
	for i := 0; i < 16 && g.Stack.Contains(id); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !g.Exile.Contains(id) {
		t.Fatal("the resolved Adventure spell is not in exile")
	}
	return id
}

func TestEnumeratorOffersTheLandFromTheLandAdventureExile(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	id := castLandAdventureToExile(t, g, seat)

	moves := legal.EnumerateFor(g, seat.ID)
	got := landAdvLandMoves(moves, id)
	if len(got) != 1 || faceOf(t, got[0]) != 0 {
		t.Fatalf("exiled land adventure offered %v, want the land play (face 0) only", labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}

func TestEnumeratorOffersNoLandFromExileOnceTheDropIsSpent(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	id := castLandAdventureToExile(t, g, seat)

	other := landAdventureInHand(seat)
	if err := g.CastSpell(seat.ID, other, game.CastSpellParams{}); err != nil {
		t.Fatalf("spend the land drop: %v", err)
	}
	if got := landAdvLandMoves(legal.EnumerateFor(g, seat.ID), id); len(got) != 0 {
		t.Fatalf("offered %d moves for the exiled land with no land drop left", len(got))
	}
}
