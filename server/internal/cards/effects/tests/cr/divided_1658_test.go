package testscr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Rolling Thunder -------------------------------------------------

func TestRollingThunderDividesX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	effectstest.D1658Cast(t, g, "Rolling Thunder", "Sorcery", effectstest.RollingThunderOracle, game.CastSpellParams{
		XValue: 5, Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 2, b: 3},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if d := effectstest.E2Card(t, g, a).DamageMarked; d != 2 {
		t.Errorf("a: %d damage, want 2", d)
	}
	if d := effectstest.E2Card(t, g, b).DamageMarked; d != 3 {
		t.Errorf("b: %d damage, want 3", d)
	}
}

func TestRollingThunderRefusesBadSplit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	err := effectstest.D1658CastErr(t, g, "Rolling Thunder", "Sorcery", effectstest.RollingThunderOracle, game.CastSpellParams{
		XValue: 5, Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 0, b: 5},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a zero share: %v, want ErrInvalidParam", err)
	}
}

// --- Forked Bolt -------------------------------------------------

func TestForkedBoltDividesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	effectstest.D1658Cast(t, g, "Forked Bolt", "Sorcery", effectstest.ForkedBoltOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if d := effectstest.E2Card(t, g, a).DamageMarked; d != 1 {
		t.Errorf("a: %d damage, want 1", d)
	}
	if d := effectstest.E2Card(t, g, b).DamageMarked; d != 1 {
		t.Errorf("b: %d damage, want 1", d)
	}
}

func TestForkedBoltRefusesBadSplit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	err := effectstest.D1658CastErr(t, g, "Forked Bolt", "Sorcery", effectstest.ForkedBoltOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 2},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a sum of 3 for 2 points: %v, want ErrInvalidParam", err)
	}
}

// --- Twin Bolt -------------------------------------------------

func TestTwinBoltDividesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	effectstest.D1658Cast(t, g, "Twin Bolt", "Instant", effectstest.TwinBoltOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(a),
	})
	effectstest.PassPriorityAroundTable(t, g)
	if d := effectstest.E2Card(t, g, a).DamageMarked; d != 2 {
		t.Errorf("a single target gets the whole 2: %d, want 2", d)
	}
}

func TestTwinBoltRefusesBadSplit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	err := effectstest.D1658CastErr(t, g, "Twin Bolt", "Instant", effectstest.TwinBoltOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 0, b: 2},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a zero share: %v, want ErrInvalidParam", err)
	}
}

// --- Arc Lightning -------------------------------------------------

func TestArcLightningDividesThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Wall", 0, 30)
	effectstest.D1658Cast(t, g, "Arc Lightning", "Sorcery", effectstest.ArcLightningOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b, c), Distribution: map[uuid.UUID]int{a: 1, b: 1, c: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if d := effectstest.E2Card(t, g, id).DamageMarked; d != 1 {
			t.Errorf("target got %d, want 1", d)
		}
	}
}

func TestArcLightningRefusesBadSplit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Wall", 0, 30)
	err := effectstest.D1658CastErr(t, g, "Arc Lightning", "Sorcery", effectstest.ArcLightningOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b, c), Distribution: map[uuid.UUID]int{a: 1, b: 1, c: 0},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a zero share: %v, want ErrInvalidParam", err)
	}
}

// --- Electrolyze -------------------------------------------------

func TestElectrolyzeDividesAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	effectstest.D1658Cast(t, g, "Electrolyze", "Instant", effectstest.ElectrolyzeOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 1},
	})
	// Electrolyze itself just left the hand for the stack; the draw
	// should put the hand one card ABOVE this point once it resolves.
	handAfterCast := len(me.Hand.Cards)
	effectstest.PassPriorityAroundTable(t, g)
	if d := effectstest.E2Card(t, g, a).DamageMarked; d != 1 {
		t.Errorf("a: %d damage, want 1", d)
	}
	if d := effectstest.E2Card(t, g, b).DamageMarked; d != 1 {
		t.Errorf("b: %d damage, want 1", d)
	}
	if len(me.Hand.Cards) != handAfterCast+1 {
		t.Errorf("hand size %d → %d, want +1 (the draw)", handAfterCast, len(me.Hand.Cards))
	}
}

func TestElectrolyzeRefusesBadSplit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	err := effectstest.D1658CastErr(t, g, "Electrolyze", "Instant", effectstest.ElectrolyzeOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 2},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a sum of 3 for 2 points: %v, want ErrInvalidParam", err)
	}
}

// --- Magic Missile -------------------------------------------------

func TestMagicMissileCantBeCounteredAndDivides(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Wall", 0, 30)
	effectstest.D1658Cast(t, g, "Magic Missile", "Sorcery", effectstest.MagicMissileOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b, c), Distribution: map[uuid.UUID]int{a: 1, b: 1, c: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if d := effectstest.E2Card(t, g, id).DamageMarked; d != 1 {
			t.Errorf("target got %d, want 1", d)
		}
	}
	spec, ok := effects.Lookup(effectstest.MagicMissileOracle)
	if !ok || !spec.CantBeCountered {
		t.Error("Magic Missile should declare CantBeCountered")
	}
}

func TestMagicMissileRefusesBadSplit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	err := effectstest.D1658CastErr(t, g, "Magic Missile", "Sorcery", effectstest.MagicMissileOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 0, b: 3},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a zero share: %v, want ErrInvalidParam", err)
	}
}

// --- Flames of the Firebrand -------------------------------------------------

func TestFlamesOfTheFirebrandDividesThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	effectstest.D1658Cast(t, g, "Flames of the Firebrand", "Sorcery", effectstest.FlamesOfTheFirebrandOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 2, b: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if d := effectstest.E2Card(t, g, a).DamageMarked; d != 2 {
		t.Errorf("a: %d damage, want 2", d)
	}
	if d := effectstest.E2Card(t, g, b).DamageMarked; d != 1 {
		t.Errorf("b: %d damage, want 1", d)
	}
}

func TestFlamesOfTheFirebrandRefusesBadSplit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Wall", 0, 30)
	err := effectstest.D1658CastErr(t, g, "Flames of the Firebrand", "Sorcery", effectstest.FlamesOfTheFirebrandOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b, c), Distribution: map[uuid.UUID]int{a: 3, b: 0, c: 0},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a zero share: %v, want ErrInvalidParam", err)
	}
}

// --- Chandra's Pyrohelix -------------------------------------------------

func TestChandrasPyrohelixDividesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	effectstest.D1658Cast(t, g, "Chandra's Pyrohelix", "Instant", effectstest.ChandrasPyrohelixOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if d := effectstest.E2Card(t, g, a).DamageMarked; d != 1 {
		t.Errorf("a: %d damage, want 1", d)
	}
	if d := effectstest.E2Card(t, g, b).DamageMarked; d != 1 {
		t.Errorf("b: %d damage, want 1", d)
	}
}

func TestChandrasPyrohelixRefusesBadSplit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	err := effectstest.D1658CastErr(t, g, "Chandra's Pyrohelix", "Instant", effectstest.ChandrasPyrohelixOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 2, b: 2},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a sum of 4 for 2 points: %v, want ErrInvalidParam", err)
	}
}

// --- Violent Eruption -------------------------------------------------

func TestViolentEruptionDividesFourAmongAnyNumber(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	effectstest.D1658Cast(t, g, "Violent Eruption", "Instant", effectstest.ViolentEruptionOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 3},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if d := effectstest.E2Card(t, g, a).DamageMarked; d != 1 {
		t.Errorf("a: %d damage, want 1", d)
	}
	if d := effectstest.E2Card(t, g, b).DamageMarked; d != 3 {
		t.Errorf("b: %d damage, want 3", d)
	}
	spec, ok := effects.Lookup(effectstest.ViolentEruptionOracle)
	if !ok || spec.Madness != "{1}{R}{R}" {
		t.Errorf("Violent Eruption should declare Madness {1}{R}{R}, got %q", spec.Madness)
	}
}

func TestViolentEruptionRefusesBadSplit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	err := effectstest.D1658CastErr(t, g, "Violent Eruption", "Instant", effectstest.ViolentEruptionOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 0, b: 4},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a zero share: %v, want ErrInvalidParam", err)
	}
}

// --- Pyrokinesis -------------------------------------------------

func TestPyrokinesisPitchesAndDividesAmongCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	pitch := effectstest.HandCardFull(me, "Their Red Card", "Sorcery", "{R}", "", []string{"R"})

	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Pyrokinesis", TypeLine: "Instant", OracleID: effectstest.PyrokinesisOracle,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{pitch},
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 3},
	}); err != nil {
		t.Fatalf("cast Pyrokinesis: %v", err)
	}
	if me.Hand.Contains(pitch) {
		t.Error("the pitched card should have left the hand")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if d := effectstest.E2Card(t, g, a).DamageMarked; d != 1 {
		t.Errorf("a: %d damage, want 1", d)
	}
	if d := effectstest.E2Card(t, g, b).DamageMarked; d != 3 {
		t.Errorf("b: %d damage, want 3", d)
	}
}

func TestPyrokinesisRefusesBadSplit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	err := effectstest.D1658CastErr(t, g, "Pyrokinesis", "Instant", effectstest.PyrokinesisOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 0, b: 4},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a zero share: %v, want ErrInvalidParam", err)
	}
}

// --- Verdurous Gearhulk -------------------------------------------------

func TestVerdurousGearhulkDistributesFourAmongCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hulk := effectstest.CastCatalogSpell(t, g, "Verdurous Gearhulk", "Artifact Creature — Construct", effectstest.VerdurousGearhulkOracle1658, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the ETB trigger should ask for targets")
	}
	effectstest.B17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{hulk: 3, bear: 1}, hulk, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.E2Card(t, g, hulk).Counters[game.CounterPlusOne]; n != 3 {
		t.Errorf("hulk: %d +1/+1 counters, want 3", n)
	}
	if n := effectstest.E2Card(t, g, bear).Counters[game.CounterPlusOne]; n != 1 {
		t.Errorf("bear: %d +1/+1 counters, want 1", n)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, hulk), "trample") {
		t.Error("printed trample did not reach the effective abilities")
	}
}

func TestVerdurousGearhulkRefusesBadSplit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hulk := effectstest.CastCatalogSpell(t, g, "Verdurous Gearhulk", "Artifact Creature — Construct", effectstest.VerdurousGearhulkOracle1658, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the ETB trigger should ask for targets")
	}
	refs := effectstest.CardRefs(hulk, bear)
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, refs, map[uuid.UUID]int{hulk: 0, bear: 4}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a zero share: %v, want ErrInvalidParam", err)
	}
}

// --- Armament Dragon -------------------------------------------------

func TestArmamentDragonDistributesThreeAmongCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	dragon := effectstest.CastCatalogSpell(t, g, "Armament Dragon", "Creature — Dragon", effectstest.ArmamentDragonOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the ETB trigger should ask for targets")
	}
	effectstest.B17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{dragon: 2, bear: 1}, dragon, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.E2Card(t, g, dragon).Counters[game.CounterPlusOne]; n != 2 {
		t.Errorf("dragon: %d +1/+1 counters, want 2", n)
	}
	if n := effectstest.E2Card(t, g, bear).Counters[game.CounterPlusOne]; n != 1 {
		t.Errorf("bear: %d +1/+1 counters, want 1", n)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, dragon), "flying") {
		t.Error("printed flying did not reach the effective abilities")
	}
}

func TestArmamentDragonRefusesBadSplit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	dragon := effectstest.CastCatalogSpell(t, g, "Armament Dragon", "Creature — Dragon", effectstest.ArmamentDragonOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the ETB trigger should ask for targets")
	}
	refs := effectstest.CardRefs(dragon, bear)
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, refs, map[uuid.UUID]int{dragon: 1, bear: 1}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a sum of 2 for 3 points: %v, want ErrInvalidParam", err)
	}
}

// --- Glint Weaver -------------------------------------------------

func TestGlintWeaverDistributesThenGainsLifeEqualToGreatestToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// The target clause has no "you control" restriction (unlike
	// Verdurous Gearhulk / Armament Dragon), so an opponent's creature
	// is a legal target of the distribution — but the LIFE GAIN does
	// say "creatures you control", so this bigger opposing wall must
	// NOT be the one counted even though its toughness is higher.
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Wall", 0, 20)
	weaver := effectstest.CastCatalogSpell(t, g, "Glint Weaver", "Creature — Spider", effectstest.GlintWeaverOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the ETB trigger should ask for targets")
	}
	life := me.Life
	effectstest.B17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{weaver: 2, theirs: 1}, weaver, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.E2Card(t, g, weaver).Counters[game.CounterPlusOne]; n != 2 {
		t.Fatalf("weaver: %d +1/+1 counters, want 2", n)
	}
	if n := effectstest.E2Card(t, g, theirs).Counters[game.CounterPlusOne]; n != 1 {
		t.Fatalf("theirs: %d +1/+1 counters, want 1", n)
	}
	weaverToughness := effectstest.E2Card(t, g, weaver).CurrentToughness()
	if me.Life != life+weaverToughness {
		t.Errorf("life %d → %d, want +%d (weaver's own toughness — the opponent's bigger wall isn't a creature you control)", life, me.Life, weaverToughness)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, weaver), "reach") {
		t.Error("printed reach did not reach the effective abilities")
	}
}

func TestGlintWeaverRefusesBadSplit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	weaver := effectstest.CastCatalogSpell(t, g, "Glint Weaver", "Creature — Spider", effectstest.GlintWeaverOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the ETB trigger should ask for targets")
	}
	refs := effectstest.CardRefs(weaver)
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, refs, map[uuid.UUID]int{weaver: 2}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a share of 2 for a single target owed all 3: %v, want ErrInvalidParam", err)
	}
}

// --- Blessings of Nature -------------------------------------------------

func TestBlessingsOfNatureDistributesFourAmongAnyNumber(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	effectstest.D1658Cast(t, g, "Blessings of Nature", "Sorcery", effectstest.BlessingsOfNatureOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 3},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.E2Card(t, g, a).Counters[game.CounterPlusOne]; n != 1 {
		t.Errorf("a: %d +1/+1 counters, want 1", n)
	}
	if n := effectstest.E2Card(t, g, b).Counters[game.CounterPlusOne]; n != 3 {
		t.Errorf("b: %d +1/+1 counters, want 3", n)
	}
	spec, ok := effects.Lookup(effectstest.BlessingsOfNatureOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Error("Blessings of Nature should be CompletenessFull now that Miracle ships (#1665)")
	}
	if game.AlternativeCostByKey(effectstest.BlessingsOfNatureOracle, game.AltCostKeyMiracle) == nil {
		t.Error("Blessings of Nature should declare Miracle {G}")
	}
}

func TestBlessingsOfNatureRefusesBadSplit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	err := effectstest.D1658CastErr(t, g, "Blessings of Nature", "Sorcery", effectstest.BlessingsOfNatureOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 0, b: 4},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a zero share: %v, want ErrInvalidParam", err)
	}
}

// --- Splendid Agony -------------------------------------------------

func TestSplendidAgonyDistributesMinusOneOneCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 3, 3)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 3, 3)
	effectstest.D1658Cast(t, g, "Splendid Agony", "Instant", effectstest.SplendidAgonyOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.E2Card(t, g, a).Counters[game.CounterMinusOne]; n != 1 {
		t.Errorf("a: %d -1/-1 counters, want 1", n)
	}
	if n := effectstest.E2Card(t, g, b).Counters[game.CounterMinusOne]; n != 1 {
		t.Errorf("b: %d -1/-1 counters, want 1", n)
	}
	if p := effectstest.E2Card(t, g, a).CurrentPower(); p != 2 {
		t.Errorf("a's power after the counter: %d, want 2", p)
	}
}

func TestSplendidAgonyRefusesBadSplit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 3, 3)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 3, 3)
	err := effectstest.D1658CastErr(t, g, "Splendid Agony", "Instant", effectstest.SplendidAgonyOracle, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 2, b: 2},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a sum of 4 for 2 points: %v, want ErrInvalidParam", err)
	}
}
