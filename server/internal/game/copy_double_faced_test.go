package game

import (
	"testing"

	"github.com/google/uuid"
)

// copy_double_faced_test.go — CR 707.8: a permanent that becomes a copy of
// a double-faced permanent copies the face that is up and is single-faced
// (#3512).

func werewolfOnBattlefield(g *Game, owner uuid.UUID, layout string, backUp bool) uuid.UUID {
	c := Card{
		InstanceID: uuid.New(),
		OracleID:   "wolf-oracle",
		ScryfallID: "wolf-print",
		Owner:      owner,
		Controller: owner,
		Layout:     layout,
		Faces: []Face{
			{Name: "Village Human", TypeLine: "Creature — Human", Power: 1, Toughness: 1},
			{Name: "Night Wolf", TypeLine: "Creature — Werewolf", Power: 3, Toughness: 3},
		},
	}
	c.SetFace(0)
	if backUp {
		c.SetFace(1)
	}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func cloneOnBattlefield(g *Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id, Name: "Clone", TypeLine: "Creature — Shapeshifter",
		OracleID: "oracle-clone", Owner: owner, Controller: owner,
	})
	return id
}

func TestCopyOfAFrontUpDoubleFacedPermanentIsSingleFaced(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	srcID := werewolfOnBattlefield(g, me, LayoutTransform, false)
	cloneID := cloneOnBattlefield(g, me)

	g.mu.Lock()
	src, _ := g.battlefieldCardLocked(srcID)
	cl, _ := g.battlefieldCardLocked(cloneID)
	cl.applyCopy(SingleFacedCopiableValuesOf(*src), *src)
	cp := *cl
	g.mu.Unlock()

	if cp.Name != "Village Human" || cp.Layout != "" || len(cp.Faces) != 0 || cp.ActiveFace != 0 {
		t.Fatalf("copy = %q layout %q faces %d active %d, want a single-faced Village Human", cp.Name, cp.Layout, len(cp.Faces), cp.ActiveFace)
	}
	if CatalogKey(cp) != "wolf-oracle" {
		t.Errorf("catalog key = %q, want the front face's", CatalogKey(cp))
	}
	if CanTransform(cp) {
		t.Error("a Clone of a double-faced permanent can transform (CR 707.8)")
	}
	if err := g.TransformPermanentForEffect(cloneID); err != nil {
		t.Fatalf("transform: %v", err)
	}
	if got := battlefieldCard(t, g, cloneID).Name; got != "Village Human" {
		t.Errorf("after the transform attempt the copy is %q, want it unchanged", got)
	}
}

func TestCopyOfABackUpDoubleFacedPermanentCopiesTheBackFaceOnly(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	srcID := werewolfOnBattlefield(g, me, LayoutTransform, true)
	cloneID := cloneOnBattlefield(g, me)

	g.mu.Lock()
	src, _ := g.battlefieldCardLocked(srcID)
	cl, _ := g.battlefieldCardLocked(cloneID)
	cl.applyCopy(SingleFacedCopiableValuesOf(*src), *src)
	cp := *cl
	g.mu.Unlock()

	if cp.Name != "Night Wolf" || cp.Power != 3 || len(cp.Faces) != 0 || cp.ActiveFace != 0 {
		t.Fatalf("copy = %q %d/%d faces %d, want a single-faced Night Wolf 3/3", cp.Name, cp.Power, cp.Toughness, len(cp.Faces))
	}
	// The back face's abilities are found under its own catalog key.
	if CatalogKey(cp) != "wolf-oracle#1" {
		t.Errorf("catalog key = %q, want the back face's", CatalogKey(cp))
	}
	if CanTransform(cp) {
		t.Error("the copy of a back face can transform")
	}
	// The source is untouched.
	if s := battlefieldCard(t, g, srcID); s.Layout != LayoutTransform || len(s.Faces) != 2 || s.ActiveFace != 1 {
		t.Error("the copy source lost its faces")
	}
}

func TestCopyOfAModalDoubleFacedPermanentIsSingleFaced(t *testing.T) {
	g := newActiveGame(t)
	srcID := werewolfOnBattlefield(g, g.Seats[0].ID, LayoutModalDFC, false)
	g.mu.Lock()
	src, _ := g.battlefieldCardLocked(srcID)
	v := SingleFacedCopiableValuesOf(*src)
	g.mu.Unlock()
	if v.Layout != "" || len(v.Faces) != 0 || v.Name != "Village Human" {
		t.Errorf("copy values = %+v, want single-faced front", v)
	}
}

// A copy of a spell stays double-faced (CR 707.10g), so the plain
// CopiableValuesOf keeps its faces.
func TestCopiableValuesKeepFacesForNonPermanentCopies(t *testing.T) {
	g := newActiveGame(t)
	srcID := werewolfOnBattlefield(g, g.Seats[0].ID, LayoutTransform, false)
	g.mu.Lock()
	src, _ := g.battlefieldCardLocked(srcID)
	v := CopiableValuesOf(*src)
	g.mu.Unlock()
	if v.Layout != LayoutTransform || len(v.Faces) != 2 {
		t.Errorf("CopiableValuesOf dropped the faces: %+v", v)
	}
}
