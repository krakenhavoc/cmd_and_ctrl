package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prepare_view_test.go — the wire half of CR 722 (ADR 0090, #1328).
// Two facts reach the client and nothing new is needed for either
// beyond one field: the permanent says `prepared`, and the copy of its
// prepare spell sits in exile with the `exile_play` stamp the client's
// exile button has keyed off since S21 — opening the prepare-spell
// face, for the prepared permanent's controller.
func TestPreparedPermanentAndItsCopyOnTheWire(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	c := game.NewCard("Wire Conductor", me.ID)
	c.OracleID = "view-prepare-oracle"
	c.Layout = game.LayoutPrepare
	c.Faces = []game.Face{
		{Name: "Wire Conductor", TypeLine: "Creature — Bird Pilot", ManaCost: "{2}{U}", Power: 2, Toughness: 3},
		{Name: "Wire Aboard", TypeLine: "Instant", ManaCost: "{U}"},
	}
	c.SetFace(0)
	c.EnteredBattlefieldAt = 1
	c.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
	g.Battlefield.PushTop(c)
	g.WithWriteLock(func() {
		if ok, err := g.BecomePreparedForEffect(c.InstanceID); !ok || err != nil {
			t.Fatalf("BecomePreparedForEffect = %v, %v", ok, err)
		}
	})
	var copyID uuid.UUID
	for _, e := range g.Exile.Cards {
		if e.PrepareCopy {
			copyID = e.InstanceID
		}
	}

	for _, viewer := range []*game.Player{me, opp} {
		v := ViewOfGameFor(g, viewer.ID.String())
		var perm, cp *CardView
		for i := range v.Battlefield.Cards {
			if v.Battlefield.Cards[i].InstanceID == c.InstanceID.String() {
				perm = &v.Battlefield.Cards[i]
			}
		}
		for i := range v.Exile.Cards {
			if v.Exile.Cards[i].InstanceID == copyID.String() {
				cp = &v.Exile.Cards[i]
			}
		}
		if perm == nil || !perm.Prepared {
			t.Errorf("viewer %s: the permanent does not show `prepared`", viewer.Name)
		}
		if cp == nil {
			t.Fatalf("viewer %s: the copy is not in the exile view", viewer.Name)
		}
		if cp.Name != "Wire Aboard" {
			t.Errorf("viewer %s: the copy reads %q, want the prepare spell — it is public", viewer.Name, cp.Name)
		}
		if cp.ExilePlay == nil || cp.ExilePlay.Player != me.ID.String() {
			t.Fatalf("viewer %s: the copy ships no exile_play naming the controller: %+v", viewer.Name, cp.ExilePlay)
		}
		if f := cp.ExilePlay.Faces; len(f) != 1 || f[0] != 1 {
			t.Errorf("viewer %s: faces = %v, want [1] — the prepare spell", viewer.Name, f)
		}
	}
}
