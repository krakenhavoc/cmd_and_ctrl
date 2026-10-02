package deck

import (
	"slices"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// face_keywords_test.go — ADR 0107 §4. Scryfall's card-level keyword
// list is the union over every face; the importer stamps each face of a
// double-faced card with its own share, and SetFace puts the face that
// is up onto the card (CR 712.8d, 712.8e). Before this the front face's
// keywords stayed on the card whichever face was up, so a disturbed
// back face kept the front face's keywords and never had its own.
func TestDoubleFacedCardKeywordsFollowTheFaceThatIsUp(t *testing.T) {
	c := ToGameCard(aangSwiftSavior(), false)
	if got := c.Faces[0].Keywords; !slices.Equal(got, []string{"flying", "flash"}) && !slices.Equal(got, []string{"flash", "flying"}) {
		t.Errorf("front face keywords = %v, want flash and flying", got)
	}
	if got := c.Faces[1].Keywords; !slices.Contains(got, "reach") || !slices.Contains(got, "trample") || slices.Contains(got, "flying") {
		t.Errorf("back face keywords = %v, want reach and trample only", got)
	}
	if !game.HasKeyword(&c, "flash") || game.HasKeyword(&c, "reach") {
		t.Errorf("front face up, card keywords = %v; want the front face's", c.Keywords)
	}

	c.SetFace(1)
	if game.HasKeyword(&c, "flash") || game.HasKeyword(&c, "flying") {
		t.Errorf("back face up, the front face's keywords stayed: %v", c.Keywords)
	}
	if !game.HasKeyword(&c, "reach") || !game.HasKeyword(&c, "trample") {
		t.Errorf("back face up, card keywords = %v; want reach and trample", c.Keywords)
	}

	c.SetFace(0)
	if !game.HasKeyword(&c, "flash") || game.HasKeyword(&c, "trample") {
		t.Errorf("front face up again, card keywords = %v", c.Keywords)
	}
}

// A card imported before per-face keywords existed carries none on its
// faces, and keeps its card-level list through a face change, as it
// always did — a restore point is never stripped of a keyword.
func TestFaceChangeKeepsKeywordsOfACardWithNoPerFaceList(t *testing.T) {
	c := ToGameCard(aangSwiftSavior(), false)
	for i := range c.Faces {
		c.Faces[i].Keywords = nil
	}
	before := slices.Clone(c.Keywords)
	c.SetFace(1)
	if !slices.Equal(c.Keywords, before) {
		t.Errorf("keywords changed from %v to %v on a card with no per-face list", before, c.Keywords)
	}
}
