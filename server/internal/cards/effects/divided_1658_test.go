package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// divided_1658_test.go — #1658, the cards #1656's divided damage and
// counters (CR 601.2d) unblocked. One test per card for its main
// split, one refusal per card for a bad split (a zero share or a sum
// mismatch), and the trigger path for the three ETB creatures
// (Verdurous Gearhulk, Armament Dragon, Glint Weaver).

const (
	rollingThunderOracle        = "9c47888b-28a5-4c43-9ee4-a9059e3c367d"
	forkedBoltOracle            = "e0548f0d-4aec-4c39-8c34-52bfa2bf7c60"
	twinBoltOracle              = "970dc070-3668-4584-9904-2897b27bb806"
	arcLightningOracle          = "5acc8b39-3c3e-4012-8cfd-ac3c2c4ca982"
	electrolyzeOracle           = "07b222d7-24f2-4994-9004-ff6672ebe161"
	magicMissileOracle          = "e48b8d78-3f8b-4bf8-853a-76c435bc31f0"
	flamesOfTheFirebrandOracle  = "bc32c24f-2f2a-4125-917c-60d425166640"
	chandrasPyrohelixOracle     = "69a37af8-6bcd-42b3-b788-0a357e742cc0"
	violentEruptionOracle       = "842578ca-0a86-4e96-bfd5-45931488f7c1"
	pyrokinesisOracle           = "d001febc-d511-4e65-a631-91dc9415056b"
	verdurousGearhulkOracle1658 = "3765e8bb-e70d-4503-b8dc-1e684e434c18"
	armamentDragonOracle        = "23c75386-8011-4ee4-97e2-eaafeb20b788"
	glintWeaverOracle           = "eebe3fd6-afd0-4bb3-9e8f-c737497dae33"
	blessingsOfNatureOracle     = "1e77cf90-ac51-4ac7-b123-be04aabe1688"
	splendidAgonyOracle         = "e9958344-13d5-4f7b-acc8-904e5d99b90e"
)

// d1658Cast casts a hand card with the full CastSpellParams a divided
// spell needs (Distribution, XValue, AlternativeCost) — castCatalogSpell
// only carries Targets, and b12PlayFromHand's land-drop bookkeeping is
// noise for a spell.
func d1658Cast(t *testing.T, g *game.Game, name, typeLine, oracleID string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracleID,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, params); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// d1658CastErr is d1658Cast for a cast that is SUPPOSED to be refused:
// the announce error comes back instead of failing the test.
func d1658CastErr(t *testing.T, g *game.Game, name, typeLine, oracleID string, params game.CastSpellParams) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracleID,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return g.CastSpell(active.ID, id, params)
}

// --- Rolling Thunder -------------------------------------------------

func TestRollingThunderDividesX(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	d1658Cast(t, g, "Rolling Thunder", "Sorcery", rollingThunderOracle, game.CastSpellParams{
		XValue: 5, Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 2, b: 3},
	})
	passPriorityAroundTable(t, g)
	if d := e2Card(t, g, a).DamageMarked; d != 2 {
		t.Errorf("a: %d damage, want 2", d)
	}
	if d := e2Card(t, g, b).DamageMarked; d != 3 {
		t.Errorf("b: %d damage, want 3", d)
	}
}

func TestRollingThunderRefusesBadSplit(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	err := d1658CastErr(t, g, "Rolling Thunder", "Sorcery", rollingThunderOracle, game.CastSpellParams{
		XValue: 5, Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 0, b: 5},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a zero share: %v, want ErrInvalidParam", err)
	}
}

// --- Forked Bolt -------------------------------------------------

func TestForkedBoltDividesTwo(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	d1658Cast(t, g, "Forked Bolt", "Sorcery", forkedBoltOracle, game.CastSpellParams{
		Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 1},
	})
	passPriorityAroundTable(t, g)
	if d := e2Card(t, g, a).DamageMarked; d != 1 {
		t.Errorf("a: %d damage, want 1", d)
	}
	if d := e2Card(t, g, b).DamageMarked; d != 1 {
		t.Errorf("b: %d damage, want 1", d)
	}
}

func TestForkedBoltRefusesBadSplit(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	err := d1658CastErr(t, g, "Forked Bolt", "Sorcery", forkedBoltOracle, game.CastSpellParams{
		Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 2},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a sum of 3 for 2 points: %v, want ErrInvalidParam", err)
	}
}

// --- Twin Bolt -------------------------------------------------

func TestTwinBoltDividesTwo(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	d1658Cast(t, g, "Twin Bolt", "Instant", twinBoltOracle, game.CastSpellParams{
		Targets: cardRefs(a),
	})
	passPriorityAroundTable(t, g)
	if d := e2Card(t, g, a).DamageMarked; d != 2 {
		t.Errorf("a single target gets the whole 2: %d, want 2", d)
	}
}

func TestTwinBoltRefusesBadSplit(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	err := d1658CastErr(t, g, "Twin Bolt", "Instant", twinBoltOracle, game.CastSpellParams{
		Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 0, b: 2},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a zero share: %v, want ErrInvalidParam", err)
	}
}

// --- Arc Lightning -------------------------------------------------

func TestArcLightningDividesThree(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	c := b12Creature(g, opp.ID, "C", "Creature — Wall", 0, 30)
	d1658Cast(t, g, "Arc Lightning", "Sorcery", arcLightningOracle, game.CastSpellParams{
		Targets: cardRefs(a, b, c), Distribution: map[uuid.UUID]int{a: 1, b: 1, c: 1},
	})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if d := e2Card(t, g, id).DamageMarked; d != 1 {
			t.Errorf("target got %d, want 1", d)
		}
	}
}

func TestArcLightningRefusesBadSplit(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	c := b12Creature(g, opp.ID, "C", "Creature — Wall", 0, 30)
	err := d1658CastErr(t, g, "Arc Lightning", "Sorcery", arcLightningOracle, game.CastSpellParams{
		Targets: cardRefs(a, b, c), Distribution: map[uuid.UUID]int{a: 1, b: 1, c: 0},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a zero share: %v, want ErrInvalidParam", err)
	}
}

// --- Electrolyze -------------------------------------------------

func TestElectrolyzeDividesAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	d1658Cast(t, g, "Electrolyze", "Instant", electrolyzeOracle, game.CastSpellParams{
		Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 1},
	})
	// Electrolyze itself just left the hand for the stack; the draw
	// should put the hand one card ABOVE this point once it resolves.
	handAfterCast := len(me.Hand.Cards)
	passPriorityAroundTable(t, g)
	if d := e2Card(t, g, a).DamageMarked; d != 1 {
		t.Errorf("a: %d damage, want 1", d)
	}
	if d := e2Card(t, g, b).DamageMarked; d != 1 {
		t.Errorf("b: %d damage, want 1", d)
	}
	if len(me.Hand.Cards) != handAfterCast+1 {
		t.Errorf("hand size %d → %d, want +1 (the draw)", handAfterCast, len(me.Hand.Cards))
	}
}

func TestElectrolyzeRefusesBadSplit(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	err := d1658CastErr(t, g, "Electrolyze", "Instant", electrolyzeOracle, game.CastSpellParams{
		Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 2},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a sum of 3 for 2 points: %v, want ErrInvalidParam", err)
	}
}

// --- Magic Missile -------------------------------------------------

func TestMagicMissileCantBeCounteredAndDivides(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	c := b12Creature(g, opp.ID, "C", "Creature — Wall", 0, 30)
	d1658Cast(t, g, "Magic Missile", "Sorcery", magicMissileOracle, game.CastSpellParams{
		Targets: cardRefs(a, b, c), Distribution: map[uuid.UUID]int{a: 1, b: 1, c: 1},
	})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if d := e2Card(t, g, id).DamageMarked; d != 1 {
			t.Errorf("target got %d, want 1", d)
		}
	}
	spec, ok := Lookup(magicMissileOracle)
	if !ok || !spec.CantBeCountered {
		t.Error("Magic Missile should declare CantBeCountered")
	}
}

func TestMagicMissileRefusesBadSplit(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	err := d1658CastErr(t, g, "Magic Missile", "Sorcery", magicMissileOracle, game.CastSpellParams{
		Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 0, b: 3},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a zero share: %v, want ErrInvalidParam", err)
	}
}

// --- Flames of the Firebrand -------------------------------------------------

func TestFlamesOfTheFirebrandDividesThree(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	d1658Cast(t, g, "Flames of the Firebrand", "Sorcery", flamesOfTheFirebrandOracle, game.CastSpellParams{
		Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 2, b: 1},
	})
	passPriorityAroundTable(t, g)
	if d := e2Card(t, g, a).DamageMarked; d != 2 {
		t.Errorf("a: %d damage, want 2", d)
	}
	if d := e2Card(t, g, b).DamageMarked; d != 1 {
		t.Errorf("b: %d damage, want 1", d)
	}
}

func TestFlamesOfTheFirebrandRefusesBadSplit(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	c := b12Creature(g, opp.ID, "C", "Creature — Wall", 0, 30)
	err := d1658CastErr(t, g, "Flames of the Firebrand", "Sorcery", flamesOfTheFirebrandOracle, game.CastSpellParams{
		Targets: cardRefs(a, b, c), Distribution: map[uuid.UUID]int{a: 3, b: 0, c: 0},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a zero share: %v, want ErrInvalidParam", err)
	}
}

// --- Chandra's Pyrohelix -------------------------------------------------

func TestChandrasPyrohelixDividesTwo(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	d1658Cast(t, g, "Chandra's Pyrohelix", "Instant", chandrasPyrohelixOracle, game.CastSpellParams{
		Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 1},
	})
	passPriorityAroundTable(t, g)
	if d := e2Card(t, g, a).DamageMarked; d != 1 {
		t.Errorf("a: %d damage, want 1", d)
	}
	if d := e2Card(t, g, b).DamageMarked; d != 1 {
		t.Errorf("b: %d damage, want 1", d)
	}
}

func TestChandrasPyrohelixRefusesBadSplit(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	err := d1658CastErr(t, g, "Chandra's Pyrohelix", "Instant", chandrasPyrohelixOracle, game.CastSpellParams{
		Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 2, b: 2},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a sum of 4 for 2 points: %v, want ErrInvalidParam", err)
	}
}

// --- Violent Eruption -------------------------------------------------

func TestViolentEruptionDividesFourAmongAnyNumber(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	d1658Cast(t, g, "Violent Eruption", "Instant", violentEruptionOracle, game.CastSpellParams{
		Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 3},
	})
	passPriorityAroundTable(t, g)
	if d := e2Card(t, g, a).DamageMarked; d != 1 {
		t.Errorf("a: %d damage, want 1", d)
	}
	if d := e2Card(t, g, b).DamageMarked; d != 3 {
		t.Errorf("b: %d damage, want 3", d)
	}
	spec, ok := Lookup(violentEruptionOracle)
	if !ok || spec.Madness != "{1}{R}{R}" {
		t.Errorf("Violent Eruption should declare Madness {1}{R}{R}, got %q", spec.Madness)
	}
}

func TestViolentEruptionRefusesBadSplit(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	err := d1658CastErr(t, g, "Violent Eruption", "Instant", violentEruptionOracle, game.CastSpellParams{
		Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 0, b: 4},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a zero share: %v, want ErrInvalidParam", err)
	}
}

// --- Pyrokinesis -------------------------------------------------

func TestPyrokinesisPitchesAndDividesAmongCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	pitch := handCardFull(me, "Their Red Card", "Sorcery", "{R}", "", []string{"R"})

	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Pyrokinesis", TypeLine: "Instant", OracleID: pyrokinesisOracle,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{pitch},
		Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 3},
	}); err != nil {
		t.Fatalf("cast Pyrokinesis: %v", err)
	}
	if me.Hand.Contains(pitch) {
		t.Error("the pitched card should have left the hand")
	}
	passPriorityAroundTable(t, g)
	if d := e2Card(t, g, a).DamageMarked; d != 1 {
		t.Errorf("a: %d damage, want 1", d)
	}
	if d := e2Card(t, g, b).DamageMarked; d != 3 {
		t.Errorf("b: %d damage, want 3", d)
	}
}

func TestPyrokinesisRefusesBadSplit(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	err := d1658CastErr(t, g, "Pyrokinesis", "Instant", pyrokinesisOracle, game.CastSpellParams{
		Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 0, b: 4},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a zero share: %v, want ErrInvalidParam", err)
	}
}

// --- Verdurous Gearhulk -------------------------------------------------

func TestVerdurousGearhulkDistributesFourAmongCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hulk := castCatalogSpell(t, g, "Verdurous Gearhulk", "Artifact Creature — Construct", verdurousGearhulkOracle1658, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the ETB trigger should ask for targets")
	}
	b17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{hulk: 3, bear: 1}, hulk, bear)
	passPriorityAroundTable(t, g)
	if n := e2Card(t, g, hulk).Counters[game.CounterPlusOne]; n != 3 {
		t.Errorf("hulk: %d +1/+1 counters, want 3", n)
	}
	if n := e2Card(t, g, bear).Counters[game.CounterPlusOne]; n != 1 {
		t.Errorf("bear: %d +1/+1 counters, want 1", n)
	}
	if !containsString(effectiveAbilities(t, g, hulk), "trample") {
		t.Error("printed trample did not reach the effective abilities")
	}
}

func TestVerdurousGearhulkRefusesBadSplit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hulk := castCatalogSpell(t, g, "Verdurous Gearhulk", "Artifact Creature — Construct", verdurousGearhulkOracle1658, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the ETB trigger should ask for targets")
	}
	refs := cardRefs(hulk, bear)
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, refs, map[uuid.UUID]int{hulk: 0, bear: 4}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a zero share: %v, want ErrInvalidParam", err)
	}
}

// --- Armament Dragon -------------------------------------------------

func TestArmamentDragonDistributesThreeAmongCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	dragon := castCatalogSpell(t, g, "Armament Dragon", "Creature — Dragon", armamentDragonOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the ETB trigger should ask for targets")
	}
	b17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{dragon: 2, bear: 1}, dragon, bear)
	passPriorityAroundTable(t, g)
	if n := e2Card(t, g, dragon).Counters[game.CounterPlusOne]; n != 2 {
		t.Errorf("dragon: %d +1/+1 counters, want 2", n)
	}
	if n := e2Card(t, g, bear).Counters[game.CounterPlusOne]; n != 1 {
		t.Errorf("bear: %d +1/+1 counters, want 1", n)
	}
	if !containsString(effectiveAbilities(t, g, dragon), "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
}

func TestArmamentDragonRefusesBadSplit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	dragon := castCatalogSpell(t, g, "Armament Dragon", "Creature — Dragon", armamentDragonOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the ETB trigger should ask for targets")
	}
	refs := cardRefs(dragon, bear)
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, refs, map[uuid.UUID]int{dragon: 1, bear: 1}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a sum of 2 for 3 points: %v, want ErrInvalidParam", err)
	}
}

// --- Glint Weaver -------------------------------------------------

func TestGlintWeaverDistributesThenGainsLifeEqualToGreatestToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// The target clause has no "you control" restriction (unlike
	// Verdurous Gearhulk / Armament Dragon), so an opponent's creature
	// is a legal target of the distribution — but the LIFE GAIN does
	// say "creatures you control", so this bigger opposing wall must
	// NOT be the one counted even though its toughness is higher.
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Wall", 0, 20)
	weaver := castCatalogSpell(t, g, "Glint Weaver", "Creature — Spider", glintWeaverOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the ETB trigger should ask for targets")
	}
	life := me.Life
	b17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{weaver: 2, theirs: 1}, weaver, theirs)
	passPriorityAroundTable(t, g)
	if n := e2Card(t, g, weaver).Counters[game.CounterPlusOne]; n != 2 {
		t.Fatalf("weaver: %d +1/+1 counters, want 2", n)
	}
	if n := e2Card(t, g, theirs).Counters[game.CounterPlusOne]; n != 1 {
		t.Fatalf("theirs: %d +1/+1 counters, want 1", n)
	}
	weaverToughness := e2Card(t, g, weaver).CurrentToughness()
	if me.Life != life+weaverToughness {
		t.Errorf("life %d → %d, want +%d (weaver's own toughness — the opponent's bigger wall isn't a creature you control)", life, me.Life, weaverToughness)
	}
	if !containsString(effectiveAbilities(t, g, weaver), "reach") {
		t.Error("printed reach did not reach the effective abilities")
	}
}

func TestGlintWeaverRefusesBadSplit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	weaver := castCatalogSpell(t, g, "Glint Weaver", "Creature — Spider", glintWeaverOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the ETB trigger should ask for targets")
	}
	refs := cardRefs(weaver)
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, refs, map[uuid.UUID]int{weaver: 2}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a share of 2 for a single target owed all 3: %v, want ErrInvalidParam", err)
	}
}

// --- Blessings of Nature -------------------------------------------------

func TestBlessingsOfNatureDistributesFourAmongAnyNumber(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	d1658Cast(t, g, "Blessings of Nature", "Sorcery", blessingsOfNatureOracle, game.CastSpellParams{
		Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 3},
	})
	passPriorityAroundTable(t, g)
	if n := e2Card(t, g, a).Counters[game.CounterPlusOne]; n != 1 {
		t.Errorf("a: %d +1/+1 counters, want 1", n)
	}
	if n := e2Card(t, g, b).Counters[game.CounterPlusOne]; n != 3 {
		t.Errorf("b: %d +1/+1 counters, want 3", n)
	}
	spec, ok := Lookup(blessingsOfNatureOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Error("Blessings of Nature should be CompletenessFull now that Miracle ships (#1665)")
	}
	if game.AlternativeCostByKey(blessingsOfNatureOracle, game.AltCostKeyMiracle) == nil {
		t.Error("Blessings of Nature should declare Miracle {G}")
	}
}

func TestBlessingsOfNatureRefusesBadSplit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	err := d1658CastErr(t, g, "Blessings of Nature", "Sorcery", blessingsOfNatureOracle, game.CastSpellParams{
		Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 0, b: 4},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a zero share: %v, want ErrInvalidParam", err)
	}
}

// --- Splendid Agony -------------------------------------------------

func TestSplendidAgonyDistributesMinusOneOneCounters(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Bear", 3, 3)
	b := b12Creature(g, opp.ID, "B", "Creature — Bear", 3, 3)
	d1658Cast(t, g, "Splendid Agony", "Instant", splendidAgonyOracle, game.CastSpellParams{
		Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 1},
	})
	passPriorityAroundTable(t, g)
	if n := e2Card(t, g, a).Counters[game.CounterMinusOne]; n != 1 {
		t.Errorf("a: %d -1/-1 counters, want 1", n)
	}
	if n := e2Card(t, g, b).Counters[game.CounterMinusOne]; n != 1 {
		t.Errorf("b: %d -1/-1 counters, want 1", n)
	}
	if p := e2Card(t, g, a).CurrentPower(); p != 2 {
		t.Errorf("a's power after the counter: %d, want 2", p)
	}
}

func TestSplendidAgonyRefusesBadSplit(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Bear", 3, 3)
	b := b12Creature(g, opp.ID, "B", "Creature — Bear", 3, 3)
	err := d1658CastErr(t, g, "Splendid Agony", "Instant", splendidAgonyOracle, game.CastSpellParams{
		Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 2, b: 2},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a sum of 4 for 2 points: %v, want ErrInvalidParam", err)
	}
}
