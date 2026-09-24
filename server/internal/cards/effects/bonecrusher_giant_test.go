package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bonecrusher_giant_test.go — the first TARGETED adventure card
// (#992). Foulmire Knight proved the CR 715 lifecycle with a half that
// announces nothing; this one proves the half that announces a target,
// which is what the per-face announce data on the wire exists for.
//
// The wire half is server/internal/protocol/per_face_announce_view_test.go
// and the client half is client/src/lib/adventureCast.test.ts. If those
// pass and this fails, the bug is in this file's Spec rather than in
// the seam.

// bonecrusherGiant builds the printed card the way deck.toGameCard
// does: per-face data, then SetFace(0).
func bonecrusherGiant(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   bonecrusherGiantOracleID,
		Layout:     game.LayoutAdventure,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{
				Name:      "Bonecrusher Giant",
				TypeLine:  "Creature — Giant",
				ManaCost:  "{2}{R}",
				Colors:    []string{"R"},
				Power:     4,
				Toughness: 3,
			},
			{
				Name:     "Stomp",
				TypeLine: "Instant — Adventure",
				ManaCost: "{1}{R}",
				Colors:   []string{"R"},
			},
		},
	}
	c.SetFace(0)
	return c
}

// handWithBonecrusherGiant seats the card in the active player's hand
// at a main phase.
func handWithBonecrusherGiant(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	c := bonecrusherGiant(me.ID)
	me.Hand.PushTop(c)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return g, me, c.InstanceID
}

// Stomp at a player: two damage, then CR 715.3d exiles the card with
// CR 715.4's permission over the creature half.
func TestStompDealsTwoAndExilesTheCard(t *testing.T) {
	g, me, id := handWithBonecrusherGiant(t)
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	before := them.Life

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Face:    1,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	}); err != nil {
		t.Fatalf("cast Stomp: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := them.Life; got != before-2 {
		t.Errorf("life = %d, want %d — Stomp deals 2 to any target", got, before-2)
	}
	if !g.Exile.Contains(id) {
		t.Fatal("CR 715.3d: the resolved Adventure spell is not in exile")
	}
	perm := g.CastPermissionOnCardByIDForEffect(id)
	if face, ok := perm.NamedFace(); !ok || face != 0 {
		t.Errorf("grant faces = %v, want just the creature face [0]", perm.Faces)
	}
}

// The creature half's trigger: a spell that targets the Giant is
// answered before it resolves (CR 601.2c puts the trigger on the stack
// above it), and the two damage goes to THAT SPELL'S controller rather
// than to the Giant's controller or to the spell's own target.
func TestBonecrusherGiantAnswersASpellThatTargetsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	giant := bonecrusherGiant(me.ID)
	id := giant.InstanceID
	g.Battlefield.PushTop(giant)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}

	// An uncatalogued instant in the OPPONENT's hand, pointed at the
	// Giant. A fixture spell rather than a catalog card, so the
	// assertion is about this trigger rather than about what the other
	// card does on resolution.
	bolt := game.NewCard("Test Removal", them.ID)
	bolt.TypeLine = "Instant"
	bolt.ManaCost = "{R}"
	them.Hand.PushTop(bolt)

	before, mineBefore := them.Life, me.Life
	if err := g.CastSpell(them.ID, bolt.InstanceID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}},
	}); err != nil {
		t.Fatalf("cast the targeting spell: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := them.Life; got != before-2 {
		t.Errorf("the targeting spell's controller is at %d, want %d — "+
			"the Giant deals 2 to THAT SPELL'S controller", got, before-2)
	}
	// And not to the Giant's own controller, which is the other way
	// the sentence could have been read.
	if me.Life != mineBefore {
		t.Errorf("the Giant's controller took %d damage", mineBefore-me.Life)
	}
}

// An ABILITY that targets the Giant does not trigger it. "Becomes the
// target of A SPELL" is narrower than the "spell or ability" wording
// the family usually prints; EventBecomesTarget is emitted for both
// and carries no discriminator, so the spell half is read off the
// stack (SelfTargetedByASpell). Firing on abilities too would be a
// strictly better Bonecrusher Giant, which is the direction this
// catalog does not ship (#259).
//
// Asserted against the ability's own condition rather than through a
// second cast, because what is being pinned is the DISCRIMINATOR: an
// event whose Source names no stack item is an ability's.
func TestBonecrusherGiantIgnoresATargetingABILITY(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	giant := bonecrusherGiant(me.ID)
	id := giant.InstanceID
	g.Battlefield.PushTop(giant)
	src := &g.Battlefield.Cards[len(g.Battlefield.Cards)-1]

	spec, ok := Lookup(bonecrusherGiantOracleID)
	if !ok || len(spec.Triggered) != 1 {
		t.Fatalf("the creature half registers %d triggers, want exactly one", len(spec.Triggered))
	}
	byAnAbility := game.Event{
		Kind:   game.EventBecomesTarget,
		CardID: id,
		// A freshly minted ability ID: no stack item shares it, which
		// is exactly how an ability differs from a spell here.
		Source: uuid.New(),
	}
	if spec.Triggered[0].AppliesTo(byAnAbility, src, src.Effective(), g) {
		t.Error("an ABILITY that targets the Giant triggers it")
	}
}

// Both faces register, and each declares its own coverage: the
// creature is full, Stomp carries the one printed clause the engine
// does not model.
func TestBothFacesOfBonecrusherGiantAreRegistered(t *testing.T) {
	for _, tc := range []struct {
		key, name string
		want      Completeness
	}{
		{bonecrusherGiantOracleID, "Bonecrusher Giant", CompletenessFull},
		{bonecrusherGiantOracleID + "#1", "Stomp", CompletenessCaveats},
	} {
		spec, ok := Lookup(tc.key)
		if !ok {
			t.Fatalf("no spec registered for %q", tc.key)
		}
		if spec.Name != tc.name {
			t.Errorf("%q registers %q, want %q", tc.key, spec.Name, tc.name)
		}
		if spec.Completeness != tc.want {
			t.Errorf("%q declares %s, want %s", tc.key, spec.Completeness, tc.want)
		}
	}
	// The Adventure half is the one that targets, and the wire reads
	// its clause off this key — face 1's, never the creature's.
	if game.TargetModeFor(bonecrusherGiantOracleID) != "" {
		t.Error("the creature half declares a target mode")
	}
	if got := game.TargetModeFor(bonecrusherGiantOracleID + "#1"); got != "any" {
		t.Errorf("Stomp's target mode = %q, want \"any\"", got)
	}
}
