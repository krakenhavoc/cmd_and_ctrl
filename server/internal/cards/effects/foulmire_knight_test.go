package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// foulmire_knight_test.go — the CARD half of CR 715 (#719). The
// engine half, against a fixture with no catalog entry, is
// server/internal/game/adventure_test.go; if this fails and that
// passes, the bug is in this file's Spec rather than in the rules.
//
// What a real adventure card adds over the fixture is the composite
// catalog key: the Adventure half's Spec is registered under
// "<oracle_id>#1", and it is only reached because CastSpell
// materialises face 1 before CatalogKey is read (ADR 0034 §5). A
// wiring mistake there shows up here as a card that draws nothing.

// foulmireKnight builds the printed card the way deck.toGameCard
// does: per-face data, then SetFace(0).
func foulmireKnight(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   foulmireKnightOracleID,
		Layout:     game.LayoutAdventure,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{
				Name:      "Foulmire Knight",
				TypeLine:  "Creature — Zombie Knight",
				ManaCost:  "{B}",
				Colors:    []string{"B"},
				Power:     1,
				Toughness: 1,
			},
			{
				Name:     "Profane Insight",
				TypeLine: "Instant — Adventure",
				ManaCost: "{2}{B}",
				Colors:   []string{"B"},
			},
		},
	}
	c.SetFace(0)
	return c
}

// handWithFoulmireKnight seats the card in the active player's hand
// at a main phase.
func handWithFoulmireKnight(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	c := foulmireKnight(me.ID)
	me.Hand.PushTop(c)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return g, me, c.InstanceID
}

// TestProfaneInsightResolvesAndExilesTheCard is the whole card in one
// pass: the Adventure half's own effect runs (a card drawn, a life
// lost), and then CR 715.3d puts the card in exile with CR 715.4's
// permission rather than in a graveyard.
func TestProfaneInsightResolvesAndExilesTheCard(t *testing.T) {
	g, me, id := handWithFoulmireKnight(t)
	handBefore, lifeBefore := me.Hand.Size(), me.Life

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Profane Insight: %v", err)
	}
	passPriorityAroundTable(t, g)

	// The card left hand for the stack, then exile, and one more was
	// drawn: net zero against the starting hand.
	if got := me.Hand.Size(); got != handBefore {
		t.Errorf("hand = %d, want %d — Profane Insight draws one and the "+
			"card itself leaves", got, handBefore)
	}
	if got := me.Life; got != lifeBefore-1 {
		t.Errorf("life = %d, want %d — 'you lose 1 life'", got, lifeBefore-1)
	}
	if !g.Exile.Contains(id) {
		t.Fatal("CR 715.3d: the resolved Adventure spell is not in exile")
	}
	perm := g.CastPermissionOnCardByIDForEffect(id)
	if !perm.Granted() || perm.Duration != game.WhileInZoneDuration() {
		t.Fatalf("CR 715.4 grant = %+v, want an unbounded exile permission", perm)
	}
	if face, ok := perm.NamedFace(); !ok || face != 0 {
		t.Errorf("grant faces = %v, want just the creature face [0]", perm.Faces)
	}
}

// TestFoulmireKnightIsCastFromExileAfterItsAdventure is CR 715.4 on a
// real card: the creature that comes back is the 1/1 deathtouch
// Zombie Knight, not the instant.
func TestFoulmireKnightIsCastFromExileAfterItsAdventure(t *testing.T) {
	g, me, id := handWithFoulmireKnight(t)

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Profane Insight: %v", err)
	}
	passPriorityAroundTable(t, g)

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast Foulmire Knight from exile: %v", err)
	}
	passPriorityAroundTable(t, g)

	var landed *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			landed = &g.Battlefield.Cards[i]
		}
	}
	if landed == nil {
		t.Fatal("Foulmire Knight never reached the battlefield")
	}
	if landed.Name != "Foulmire Knight" || !landed.IsCreature() {
		t.Fatalf("permanent = %q (%q), want the creature half", landed.Name, landed.TypeLine)
	}
	if !game.HasKeyword(landed, "deathtouch") {
		t.Error("the creature half lost its printed deathtouch")
	}
	if g.CastPermissionOnCardByIDForEffect(id).Granted() {
		t.Error("the grant survived the cast")
	}
}

// TestBothFacesOfFoulmireKnightAreRegistered pins the composite key.
// The creature half declares its coverage under the bare oracle ID so
// the card is not badged unimplemented in hand; the Adventure half is
// the "#1" spec that actually carries rules.
func TestBothFacesOfFoulmireKnightAreRegistered(t *testing.T) {
	for _, tc := range []struct{ key, name string }{
		{foulmireKnightOracleID, "Foulmire Knight"},
		{foulmireKnightOracleID + "#1", "Profane Insight"},
	} {
		spec, ok := Lookup(tc.key)
		if !ok {
			t.Fatalf("no spec registered for %q", tc.key)
		}
		if spec.Name != tc.name {
			t.Errorf("%q registers %q, want %q", tc.key, spec.Name, tc.name)
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%q declares %s, want full", tc.key, spec.Completeness)
		}
	}
}
