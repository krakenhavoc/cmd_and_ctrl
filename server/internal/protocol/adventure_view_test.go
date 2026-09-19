package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adventure_view_test.go — the wire half of CR 715.4 (#719).
//
// A card exiled by its own Adventure keeps exile's existing surface:
// the `exile_play` grant, which is what the client's exile button has
// keyed off since S21. What is new is that the grant NAMES A FACE and
// the face it names is 0, so the wire had to stop carrying a bare
// integer whose zero meant "no opinion".

// TestExiledAdventureShipsTheCreatureFace is the regression the field
// change exists for: `faces: [0]` on the wire, not an absent field.
// A client that saw nothing would open its face picker on a cast the
// server allows exactly one face of.
func TestExiledAdventureShipsTheCreatureFace(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]

	c := game.NewCard("Adventure Knight", me.ID)
	c.OracleID = "view-adventure-oracle"
	c.Layout = game.LayoutAdventure
	c.Faces = []game.Face{
		{Name: "Adventure Knight", TypeLine: "Creature — Knight", ManaCost: "{B}", Power: 1, Toughness: 1},
		{Name: "Adventure Insight", TypeLine: "Instant — Adventure", ManaCost: "{2}{B}"},
	}
	c.SetFace(0)
	c.KnownBy = map[uuid.UUID]bool{me.ID: true, g.Seats[1].ID: true}
	g.Exile.PushTop(c)
	g.GrantCastPermissionOverCardForEffect(c.InstanceID, game.CastPermission{
		Player:   me.ID,
		Duration: game.WhileInZoneDuration(),
		CastOnly: true,
		Faces:    []int{0},
		Label:    "Adventure — cast Adventure Knight from exile",
	})

	v := ViewOfGameFor(g, me.ID.String())
	var got *ExilePlayView
	for i := range v.Exile.Cards {
		if v.Exile.Cards[i].InstanceID == c.InstanceID.String() {
			got = v.Exile.Cards[i].ExilePlay
		}
	}
	if got == nil {
		t.Fatal("the exiled adventure card ships no grant")
	}
	if len(got.Faces) != 1 || got.Faces[0] != 0 {
		t.Errorf("faces = %v, want [0] — CR 715.4 opens the creature and nothing else", got.Faces)
	}
	if !got.CastOnly {
		t.Error("cast_only is not set on a CR 715.4 grant")
	}
}

// A grant that speaks about no faces still ships none, which is every
// impulse, airbend, warp and cascade grant and the overwhelming
// majority of what the client sees.
func TestAFacelessGrantShipsNoFaces(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]

	c := game.NewCard("Impulsed Sorcery", me.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{1}{R}"
	c.KnownBy = map[uuid.UUID]bool{me.ID: true}
	g.Exile.PushTop(c)
	g.GrantCastPermissionOverCardForEffect(c.InstanceID, game.CastPermission{
		Player: me.ID,
	})

	v := ViewOfGameFor(g, me.ID.String())
	for i := range v.Exile.Cards {
		if v.Exile.Cards[i].InstanceID != c.InstanceID.String() {
			continue
		}
		if got := v.Exile.Cards[i].ExilePlay; got == nil || len(got.Faces) != 0 {
			t.Errorf("faces = %+v, want none", got)
		}
	}
}
