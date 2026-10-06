package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #2176: the wire half of an Adventure whose main half is a land. The
// grant CR 715.3d writes is a PLAY permission (cast_only absent), the
// exiled card ships castable_here while the turn's land drop is open,
// and the strip stops offering it once the drop is spent (CR 305.2).

func TestExiledLandAdventureShipsAPlayableLand(t *testing.T) {
	g, me, opp := stripTable(t)

	c := game.NewCard("Adventure Town", me.ID)
	c.OracleID = "view-land-adventure-oracle"
	c.Layout = game.LayoutAdventure
	c.Faces = []game.Face{
		{Name: "Adventure Town", TypeLine: "Land — Town"},
		{Name: "Adventure Overture", TypeLine: "Sorcery — Adventure", ManaCost: "{4}{U}{U}"},
	}
	c.SetFace(0)
	c.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
	g.Exile.PushTop(c)
	g.GrantCastPermissionOverCardForEffect(c.InstanceID, game.CastPermission{
		Player:   me.ID,
		Duration: game.WhileInZoneDuration(),
		Faces:    []int{0},
		Label:    "Adventure — play Adventure Town from exile",
	})

	find := func() CardView {
		v := ViewOfGameFor(g, me.ID.String())
		for i := range v.Exile.Cards {
			if v.Exile.Cards[i].InstanceID == c.InstanceID.String() {
				return v.Exile.Cards[i]
			}
		}
		t.Fatal("the exiled land adventure is not in the exile view")
		return CardView{}
	}

	got := find()
	if got.ExilePlay == nil || got.ExilePlay.CastOnly {
		t.Fatalf("exile_play = %+v, want a play (not cast-only) grant", got.ExilePlay)
	}
	if !got.CastableHere {
		t.Error("castable_here is false with the land drop open")
	}

	// Spend the drop: the strip must stop offering the land.
	g.WithWriteLock(func() {
		if g.LandsPlayedThisTurn == nil {
			g.LandsPlayedThisTurn = map[uuid.UUID]int{}
		}
		g.LandsPlayedThisTurn[me.ID]++
	})
	if find().CastableHere {
		t.Error("castable_here stayed true with no land drop left")
	}
}
