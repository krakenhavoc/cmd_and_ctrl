package game

import (
	"testing"

	"github.com/google/uuid"
)

// face_test.go — the Faces[ActiveFace] invariant, and the setter
// that is its only enforcement.
//
// ADR 0034 named this as the model's weak point and asked for the
// test by name: "the invariant is enforced by convention plus one
// setter; a direct write to ActiveFace desynchronises a card and
// nothing will catch it." AssertFaceInvariant is the catch.

func twoFaced() Card {
	c := Card{
		InstanceID: uuid.New(),
		OracleID:   "11111111-2222-3333-4444-555555555555",
		Layout:     LayoutModalDFC,
		Faces: []Face{
			{
				Name: "Front", TypeLine: "Creature — Elephant",
				ManaCost: "{2}{G}", Colors: []string{"G"},
				Power: 3, Toughness: 3,
			},
			{
				Name: "Back", TypeLine: "Land",
				StartingLoyalty: 0,
			},
		},
	}
	c.SetFace(0)
	return c
}

// AssertFaceInvariant walks every card in every zone of a game and
// checks that its flat printed fields really are its active face's.
// Exported from the test file so other packages' integration tests
// can call it at the end of a played-out game, which is where a
// desynchronising write would actually show up.
func AssertFaceInvariant(t *testing.T, g *Game) {
	t.Helper()
	check := func(where string, c Card) {
		if len(c.Faces) == 0 {
			if c.ActiveFace != 0 {
				t.Errorf("%s: %q has no faces but ActiveFace = %d",
					where, c.Name, c.ActiveFace)
			}
			return
		}
		if c.ActiveFace < 0 || c.ActiveFace >= len(c.Faces) {
			t.Errorf("%s: %q ActiveFace = %d, out of range for %d faces",
				where, c.Name, c.ActiveFace, len(c.Faces))
			return
		}
		f := c.Faces[c.ActiveFace]
		if c.Name != f.Name {
			t.Errorf("%s: Name = %q but Faces[%d].Name = %q — something "+
				"wrote ActiveFace without going through SetFace",
				where, c.Name, c.ActiveFace, f.Name)
		}
		if c.TypeLine != f.TypeLine {
			t.Errorf("%s: %q TypeLine = %q but face has %q",
				where, c.Name, c.TypeLine, f.TypeLine)
		}
		if c.ManaCost != f.ManaCost {
			t.Errorf("%s: %q ManaCost = %q but face has %q",
				where, c.Name, c.ManaCost, f.ManaCost)
		}
		if c.Power != f.Power || c.Toughness != f.Toughness {
			t.Errorf("%s: %q P/T = %d/%d but face has %d/%d",
				where, c.Name, c.Power, c.Toughness, f.Power, f.Toughness)
		}
		if c.StartingLoyalty != f.StartingLoyalty {
			t.Errorf("%s: %q StartingLoyalty = %d but face has %d",
				where, c.Name, c.StartingLoyalty, f.StartingLoyalty)
		}
	}
	for _, z := range []*Zone{g.Battlefield, g.Stack, g.Exile} {
		if z == nil {
			continue
		}
		for _, c := range z.Cards {
			check(string(z.Kind), c)
		}
	}
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		for _, z := range []*Zone{p.Hand, p.Library, p.Graveyard, p.Command} {
			if z == nil {
				continue
			}
			for _, c := range z.Cards {
				check(string(z.Kind), c)
			}
		}
	}
}

func TestSetFaceMaterialisesEveryPrintedField(t *testing.T) {
	c := twoFaced()
	if c.Name != "Front" || c.TypeLine != "Creature — Elephant" ||
		c.ManaCost != "{2}{G}" || c.Power != 3 || c.Toughness != 3 {
		t.Fatalf("SetFace(0) did not materialise the front face: %+v", c)
	}

	c.SetFace(1)
	if c.ActiveFace != 1 {
		t.Errorf("ActiveFace = %d, want 1", c.ActiveFace)
	}
	if c.Name != "Back" || c.TypeLine != "Land" {
		t.Errorf("back face not materialised: name=%q type=%q", c.Name, c.TypeLine)
	}
	// The fields the back face does NOT print must be CLEARED, not
	// left over from the front. A land back that kept {2}{G} would
	// be charged for a land drop; one that kept 3/3 would be a
	// creature to every P/T reader in the engine.
	if c.ManaCost != "" {
		t.Errorf("ManaCost = %q after switching to a costless face, want empty", c.ManaCost)
	}
	if c.Power != 0 || c.Toughness != 0 {
		t.Errorf("P/T = %d/%d after switching to a non-creature face, want 0/0",
			c.Power, c.Toughness)
	}
	if len(c.Colors) != 0 {
		t.Errorf("Colors = %v after switching to a colourless face, want none", c.Colors)
	}

	c.SetFace(0)
	if c.Name != "Front" || c.ManaCost != "{2}{G}" || c.Power != 3 {
		t.Error("switching back did not restore the front face")
	}
}

// TestSetFaceIsANoOpForSingleFacedCards is the promise the whole
// design rests on: the ~33,000 ordinary oracle IDs keep today's
// behaviour to the byte.
func TestSetFaceIsANoOpForSingleFacedCards(t *testing.T) {
	c := Card{
		Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2,
	}
	before := c
	for _, i := range []int{0, 1, -1, 99} {
		c.SetFace(i)
		if c.Name != before.Name || c.TypeLine != before.TypeLine ||
			c.ManaCost != before.ManaCost || c.Power != before.Power ||
			c.Toughness != before.Toughness || c.ActiveFace != 0 {
			t.Fatalf("SetFace(%d) changed a single-faced card: %+v", i, c)
		}
	}
}

// TestSetFaceClampsRatherThanPanics: the index arrives from a
// client-supplied cast parameter. Refusing the cast is the action
// layer's job (CastSpell returns ErrInvalidFace); the setter's job
// is to never leave a card in an incoherent state.
func TestSetFaceClampsOutOfRange(t *testing.T) {
	for _, i := range []int{-1, 2, 1 << 20} {
		c := twoFaced()
		c.SetFace(i)
		if c.ActiveFace != 0 || c.Name != "Front" {
			t.Errorf("SetFace(%d): ActiveFace=%d name=%q, want the front face",
				i, c.ActiveFace, c.Name)
		}
	}
}

func TestCatalogKeyIsBareForFaceZero(t *testing.T) {
	c := twoFaced()
	if got := CatalogKey(c); got != c.OracleID {
		t.Errorf("face 0 key = %q, want the bare oracle ID %q", got, c.OracleID)
	}
	c.SetFace(1)
	if got, want := CatalogKey(c), c.OracleID+"#1"; got != want {
		t.Errorf("face 1 key = %q, want %q", got, want)
	}
	// A card with no oracle ID (tokens, fixtures, the demo seed)
	// keys on the empty string regardless of face, so a token never
	// invents a lookup key it could accidentally match on.
	c.OracleID = ""
	if got := CatalogKey(c); got != "" {
		t.Errorf("empty-oracle key = %q, want empty", got)
	}
}

func TestCastableFacesPerLayout(t *testing.T) {
	cases := []struct {
		layout string
		want   int
	}{
		{LayoutModalDFC, 2},
		{LayoutTransform, 1},
		{LayoutAdventure, 1},
		{LayoutSplit, 1},
		{LayoutPrepare, 1},
		{"flip", 1},
		{"", 1},
	}
	for _, tc := range cases {
		c := twoFaced()
		c.Layout = tc.layout
		if got := len(c.CastableFaces()); got != tc.want {
			t.Errorf("layout %q: %d castable faces, want %d",
				tc.layout, got, tc.want)
		}
	}
}

func TestFaceOnResolve(t *testing.T) {
	// Only a modal DFC keeps the face it was cast as. An adventure's
	// permanent is always the creature half; a transform card always
	// enters front-up (CR 712.4).
	if got := faceOnResolve(LayoutModalDFC, 1); got != 1 {
		t.Errorf("modal DFC cast as face 1 resolves as face %d, want 1", got)
	}
	for _, layout := range []string{LayoutTransform, LayoutAdventure, LayoutSplit, ""} {
		if got := faceOnResolve(layout, 1); got != 0 {
			t.Errorf("%q cast as face 1 resolves as face %d, want 0", layout, got)
		}
	}
}

// TestCloneDoesNotAliasFaces pins the undo path. A shared Faces
// slice would survive a snapshot restore and corrupt it silently —
// the failure mode clone.go exists to prevent.
func TestCloneDoesNotAliasFaces(t *testing.T) {
	c := twoFaced()
	out := cloneCard(c)
	if len(out.Faces) != len(c.Faces) {
		t.Fatalf("clone has %d faces, want %d", len(out.Faces), len(c.Faces))
	}
	out.Faces[0].Name = "MUTATED"
	if c.Faces[0].Name != "Front" {
		t.Error("writing the clone's face list changed the original")
	}
	out2 := cloneCard(c)
	if len(out2.Faces[0].Colors) == 0 {
		t.Fatal("clone dropped a face's colours")
	}
	out2.Faces[0].Colors[0] = "MUTATED"
	if c.Faces[0].Colors[0] != "G" {
		t.Error("writing the clone's face colours changed the original")
	}
}

// TestPrintedColorsPrefersStampedColors is the colour-indicator fix
// riding along with the spine: a face with no mana cost and a
// stamped colour is that colour, not colourless. Devoid is the same
// bug in the other direction.
func TestPrintedColorsPrefersStampedColors(t *testing.T) {
	// A transform back face: no cost, blue by colour indicator.
	back := Card{Name: "Jace, Telepath Unbound", TypeLine: "Legendary Planeswalker — Jace",
		Colors: []string{"U"}}
	if got := back.Effective().Colors; len(got) != 1 || got[0] != "U" {
		t.Errorf("colour-indicator face colours = %v, want [U]", got)
	}
	// Devoid: coloured pips in the cost, colourless card.
	devoid := Card{Name: "Eldrazi Skyspawner", TypeLine: "Creature — Eldrazi Drone",
		ManaCost: "{3}{U}", Colors: nil}
	if got := devoid.Effective().Colors; len(got) != 1 || got[0] != "U" {
		t.Errorf("unstamped card falls back to the cost: got %v, want [U]", got)
	}
	// Unstamped and costless is still colourless — the fallback
	// agrees with the stamp rather than contradicting it.
	blank := Card{Name: "Fixture", TypeLine: "Land"}
	if got := blank.Effective().Colors; len(got) != 0 {
		t.Errorf("costless unstamped card colours = %v, want none", got)
	}
}
