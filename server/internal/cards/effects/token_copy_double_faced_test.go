package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// CR 707.8a: a token copy of a real transforming card is a transforming
// token with both faces, entering with the face the original shows (#3512).
func TestTokenCopyOfARealDoubleFacedPermanentIsDoubleFaced(t *testing.T) {
	for _, tc := range []struct {
		name     string
		backUp   bool
		wantName string
	}{
		{"front up", false, "Village Human"},
		{"back up", true, "Night Wolf"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := uuid.New()
			c := game.Card{
				InstanceID: id,
				OracleID:   "wolf-oracle",
				ScryfallID: "wolf-print",
				Owner:      g.Seats[0].ID,
				Controller: g.Seats[0].ID,
				Layout:     game.LayoutTransform,
				Faces: []game.Face{
					{Name: "Village Human", TypeLine: "Creature — Human", Power: 1, Toughness: 1},
					{Name: "Night Wolf", TypeLine: "Creature — Werewolf", Power: 3, Toughness: 3},
				},
			}
			c.SetFace(0)
			if tc.backUp {
				c.SetFace(1)
			}
			g.Battlefield.PushTop(c)

			var tmpl game.Card
			var ok bool
			g.ReadSnapshot(func() { tmpl, ok = TokenCopyTemplate(g, id) })
			if !ok {
				t.Fatal("no template")
			}
			if tmpl.Name != tc.wantName || tmpl.Layout != game.LayoutTransform ||
				len(tmpl.Faces) != 2 || (tmpl.ActiveFace == 1) != tc.backUp {
				t.Fatalf("template = %q layout %q faces %d active %d", tmpl.Name, tmpl.Layout, len(tmpl.Faces), tmpl.ActiveFace)
			}
			if !game.CanTransform(tmpl) {
				t.Error("the token copy can't transform")
			}
			// Turning it over keeps it a token.
			tmpl.SetFace(1 - tmpl.ActiveFace)
			if !tmpl.IsToken() {
				t.Errorf("the token is %q after transforming, want a token", tmpl.TypeLine)
			}
			// The template's faces are its own.
			tmpl.Faces[0].Name = "edited"
			var orig game.Card
			g.ReadSnapshot(func() { orig, _ = g.LookupCardForEffect(id) })
			if orig.Faces[0].Name != "Village Human" {
				t.Error("the token shares its Faces with the original")
			}
		})
	}
}
