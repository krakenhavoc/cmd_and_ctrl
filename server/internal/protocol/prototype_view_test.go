package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prototype_view_test.go — the wire half of CR 718 (#5656). A prototype
// card needs no field of its own: it ships `layout: "prototype"` and the
// inset as `faces[1]`, and the flat fields follow the face that is up.

func prototypeViewCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		ScryfallID: "11111111-2222-3333-4444-555555555555",
		Layout:     game.LayoutPrototype,
		Owner:      owner,
		Controller: owner,
		KnownBy:    map[uuid.UUID]bool{owner: true},
		Faces: game.PrototypeFaces(game.Face{
			Name: "View Engine", TypeLine: "Artifact Creature — Construct", ManaCost: "{7}", Power: 6, Toughness: 4,
			OracleText: "Prototype {2}{R} — 3/2 (You may cast this spell with different mana cost, color, and size. It keeps its abilities and types.)\nHaste",
		}),
	}
	c.SetFace(0)
	return c
}

func TestPrototypeCardShipsItsInsetAsFaceOne(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	c := prototypeViewCard(me.ID)
	me.Hand.PushTop(c)

	hand := ViewOfGameFor(g, me.ID.String()).Seats[0].Hand
	card := cardInSeatZone(t, hand, c.InstanceID)
	if card.Layout != game.LayoutPrototype || card.ActiveFace != 0 || card.ManaCost != "{7}" || card.Power != 6 {
		t.Errorf("hand card = layout %q face %d %s power %d, want the normal {7} 6/4", card.Layout, card.ActiveFace, card.ManaCost, card.Power)
	}
	inset := faceOf(t, hand, c.InstanceID, game.PrototypeFace)
	if inset.ManaCost != "{2}{R}" || inset.Power != 3 || inset.Toughness != 2 || inset.Name != "View Engine" {
		t.Errorf("faces[1] = %q %s %d/%d, want View Engine {2}{R} 3/2", inset.Name, inset.ManaCost, inset.Power, inset.Toughness)
	}
	// CR 718.1: one printed image for both.
	if front := faceOf(t, hand, c.InstanceID, 0); inset.Image != front.Image {
		t.Errorf("faces[1].image = %q, want the card's one image %q", inset.Image, front.Image)
	}
}

// The prototyped permanent's flat fields are the inset's (CR 718.3b).
func TestPrototypedPermanentShowsTheInset(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	c := prototypeViewCard(me.ID)
	c.SetFace(game.PrototypeFace)
	g.Battlefield.PushTop(c)

	var got *CardView
	view := ViewOfGameFor(g, me.ID.String())
	for i := range view.Battlefield.Cards {
		if view.Battlefield.Cards[i].InstanceID == c.InstanceID.String() {
			got = &view.Battlefield.Cards[i]
		}
	}
	if got == nil {
		t.Fatal("the permanent is not in the view")
	}
	if got.ActiveFace != game.PrototypeFace || got.ManaCost != "{2}{R}" || got.Power != 3 || got.Toughness != 2 ||
		len(got.Colors) != 1 || got.Colors[0] != "R" {
		t.Errorf("permanent = face %d %s %d/%d %v, want face 1 {2}{R} 3/2 red", got.ActiveFace, got.ManaCost, got.Power, got.Toughness, got.Colors)
	}
}
