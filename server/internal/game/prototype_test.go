package game

import "testing"

func TestParsePrototype(t *testing.T) {
	for _, tc := range []struct {
		text      string
		cost      string
		power     int
		toughness int
		ok        bool
	}{
		{"Prototype {2}{R} — 3/2 (You may cast this spell with different mana cost, color, and size. It keeps its abilities and types.)\nHaste", "{2}{R}", 3, 2, true},
		{"Flying\nPrototype {4}{U}{U} — 6/6", "{4}{U}{U}", 6, 6, true},
		{"Whenever you cast a spell with prototype, draw a card.", "", 0, 0, false},
		{"Haste", "", 0, 0, false},
	} {
		cost, p, tgh, ok := ParsePrototype(tc.text)
		if cost != tc.cost || p != tc.power || tgh != tc.toughness || ok != tc.ok {
			t.Errorf("ParsePrototype(%q) = %q %d/%d %v, want %q %d/%d %v", tc.text, cost, p, tgh, ok, tc.cost, tc.power, tc.toughness, tc.ok)
		}
	}
}

// The face rules a prototype card adds to ADR 0034's table: both faces
// are castable (CR 718.3), the permanent keeps the one that was cast
// (CR 718.3b), it is not a double-faced permanent, and a move anywhere
// but the stack or the battlefield turns it back (CR 718.4).
func TestPrototypeCardFaceRules(t *testing.T) {
	c := Card{
		OracleID: "proto-oracle",
		Layout:   LayoutPrototype,
		Faces: PrototypeFaces(Face{
			Name: "Test Engine", TypeLine: "Artifact Creature — Construct", ManaCost: "{6}", Power: 5, Toughness: 5,
			OracleText: "Prototype {1}{W}{U} — 2/2",
		}),
	}
	c.SetFace(0)
	if got := c.CastableFaces(); len(got) != 2 {
		t.Fatalf("CastableFaces = %v, want both", got)
	}
	if faceOnResolve(c.Layout, PrototypeFace) != PrototypeFace {
		t.Error("a prototyped spell does not resolve as the prototyped permanent")
	}
	c.SetFace(PrototypeFace)
	if eff := c.Effective(); !c.Prototyped() || c.ManaValue() != 3 || eff.Power != 2 || eff.Toughness != 2 || len(eff.Colors) != 2 {
		t.Errorf("prototyped = mv %d %d/%d %v, want a white-blue 2/2 with mana value 3", c.ManaValue(), eff.Power, eff.Toughness, eff.Colors)
	}
	if CatalogKey(c) != "proto-oracle" || CatalogKeyForLayoutFace(c.OracleID, c.Layout, PrototypeFace) != "proto-oracle" {
		t.Errorf("prototyped key = %q, want the bare oracle ID", CatalogKey(c))
	}
	if CanTransform(c) || isDoubleFacedPermanent(c) {
		t.Error("a prototype card is not double-faced")
	}
	v := CopiableValuesOf(c)
	if v.ActiveFace != PrototypeFace || v.ManaCost != "{1}{W}{U}" || v.Power != 2 || len(v.Faces) != 2 {
		t.Errorf("copiable values = face %d %s power %d, want the inset (CR 718.2a)", v.ActiveFace, v.ManaCost, v.Power)
	}

	src := &Zone{Kind: ZoneBattlefield}
	dst := &Zone{Kind: ZoneHand}
	src.PushTop(c)
	moved, err := MoveCard(src, dst, c.InstanceID)
	if err != nil {
		t.Fatalf("MoveCard: %v", err)
	}
	if moved.Prototyped() || moved.ManaCost != "{6}" || moved.Power != 5 {
		t.Errorf("bounced card = %s power %d prototyped=%v, want the normal card (CR 718.4)", moved.ManaCost, moved.Power, moved.Prototyped())
	}
}
