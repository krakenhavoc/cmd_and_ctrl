package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mom_5649_cards_test.go — the "attacks a battle" seam (#5649) and the
// four March of the Machine cards on it. Each test plays the real
// catalog entry through the engine.

const (
	mom5649HistorianOracle  = "2676d0c3-0a24-4e93-ae80-8f3fe03da483"
	mom5649FrontlinerOracle = "e75f1834-10b8-4362-adbe-847ad9909dee"
	mom5649SlasherOracle    = "bc89face-92b1-47d3-8cdb-360b7334d3b8"
	mom5649GeodermOracle    = "e536e0d2-7117-4ca8-bdb4-e7dd380113c6"
)

// mom5649Battle puts a Siege on the battlefield that `protector`
// protects, so seat 0 may attack it.
func mom5649Battle(g *game.Game, owner, protector uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Siege", TypeLine: "Battle — Siege",
		Owner: owner, Controller: owner, ProtectorPlayerID: protector,
		Counters: map[string]int{game.CounterDefense: 9},
	})
}

// mom5649Has reads a keyword off the layered view, recomputed first.
func mom5649Has(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	return effectstest.HasEffectiveKeyword(t, g, id, kw)
}

// --- War Historian ----------------------------------------------------

func TestMom5649WarHistorianIsIndestructibleOnceItAttackedABattle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	historian := effectstest.B12Push(g, me.ID, "War Historian", "Creature — Human Monk", mom5649HistorianOracle, 3, 3)
	siege := mom5649Battle(g, me.ID, opp.ID)
	if !mom5649Has(t, g, historian, "reach") {
		t.Error("War Historian has no reach")
	}
	if mom5649Has(t, g, historian, "indestructible") {
		t.Fatal("indestructible before it attacked anything")
	}
	effectstest.DeclareAttack(t, g, siege, historian)
	if !mom5649Has(t, g, historian, "indestructible") {
		t.Fatal("no indestructible as it was declared attacking a battle")
	}

	// After combat, with the battle gone: it still attacked a battle
	// this turn, and a destroy effect does nothing to it.
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.KillCreature(t, g, me.ID, siege)
	if !mom5649Has(t, g, historian, "indestructible") {
		t.Fatal("indestructible ended with the combat or with the battle")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(historian); err != nil {
			t.Errorf("DestroyPermanentForEffect: %v", err)
		}
	})
	if !g.Battlefield.Contains(historian) {
		t.Fatal("War Historian was destroyed while indestructible")
	}

	// A restore mid-turn keeps the fact.
	restored := effectstest.RestoreRoundTrip(t, g, true)
	if !mom5649Has(t, restored, historian, "indestructible") {
		t.Error("the restored game forgot War Historian attacked a battle this turn")
	}

	// The next turn it has attacked nothing.
	effectstest.AdvanceToMainOf(t, g, 1)
	if mom5649Has(t, g, historian, "indestructible") {
		t.Error("indestructible carried into the next turn")
	}
}

func TestMom5649WarHistorianAttackingAPlayerOrAPlaneswalkerIsNotIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	historian := effectstest.B12Push(g, me.ID, "War Historian", "Creature — Human Monk", mom5649HistorianOracle, 3, 3)
	second := effectstest.B12Push(g, me.ID, "War Historian", "Creature — Human Monk", mom5649HistorianOracle, 3, 3)
	bystander := effectstest.B12Push(g, me.ID, "War Historian", "Creature — Human Monk", mom5649HistorianOracle, 3, 3)
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	siege := mom5649Battle(g, me.ID, opp.ID)
	walker := effectstest.P1g2Planeswalker(g, opp.ID, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, d := range []struct{ id, at uuid.UUID }{{historian, opp.ID}, {second, walker}, {other, siege}} {
		if err := g.DeclareAttacker(d.id, d.at); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	if mom5649Has(t, g, historian, "indestructible") {
		t.Error("attacking a player gave indestructible")
	}
	if mom5649Has(t, g, second, "indestructible") {
		t.Error("attacking a planeswalker gave indestructible")
	}
	if mom5649Has(t, g, bystander, "indestructible") {
		t.Error("another creature attacking a battle gave a War Historian that stayed home indestructible")
	}
}

// --- Thrashing Frontliner ---------------------------------------------

func TestMom5649ThrashingFrontlinerPumpsOnlyWhenItAttacksABattle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atBattle := effectstest.B12Push(g, me.ID, "Thrashing Frontliner", "Creature — Phyrexian Lizard", mom5649FrontlinerOracle, 2, 2)
	atPlayer := effectstest.B12Push(g, me.ID, "Thrashing Frontliner", "Creature — Phyrexian Lizard", mom5649FrontlinerOracle, 2, 2)
	siege := mom5649Battle(g, me.ID, opp.ID)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, d := range []struct{ id, at uuid.UUID }{{atBattle, siege}, {atPlayer, opp.ID}} {
		if err := g.DeclareAttacker(d.id, d.at); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	if effectstest.TriggerOnStack(g, atBattle) == nil {
		t.Fatal("no trigger for the Frontliner attacking a battle")
	}
	if effectstest.TriggerOnStack(g, atPlayer) != nil {
		t.Error("the Frontliner attacking a player triggered")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, atBattle); p != 3 || tt != 3 {
		t.Errorf("attacking a battle: %d/%d, want 3/3", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, atPlayer); p != 2 || tt != 2 {
		t.Errorf("attacking a player: %d/%d, want 2/2", p, tt)
	}
	if !mom5649Has(t, g, atBattle, "trample") {
		t.Error("Thrashing Frontliner has no trample")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if p, tt := effectstest.PtOf(t, g, atBattle); p != 2 || tt != 2 {
		t.Errorf("next turn: %d/%d, want 2/2 (until end of turn)", p, tt)
	}
}

// --- War-Trained Slasher ----------------------------------------------

func TestMom5649WarTrainedSlasherDoublesItsPowerWhenItAttacksABattle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atBattle := effectstest.B12Push(g, me.ID, "War-Trained Slasher", "Creature — Wolverine Dinosaur", mom5649SlasherOracle, 4, 3)
	atWalker := effectstest.B12Push(g, me.ID, "War-Trained Slasher", "Creature — Wolverine Dinosaur", mom5649SlasherOracle, 4, 3)
	siege := mom5649Battle(g, me.ID, opp.ID)
	walker := effectstest.P1g2Planeswalker(g, opp.ID, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, d := range []struct{ id, at uuid.UUID }{{atBattle, siege}, {atWalker, walker}} {
		if err := g.DeclareAttacker(d.id, d.at); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	if effectstest.TriggerOnStack(g, atWalker) != nil {
		t.Error("the Slasher attacking a planeswalker triggered")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, atBattle); p != 8 || tt != 3 {
		t.Errorf("attacking a battle: %d/%d, want 8/3", p, tt)
	}
	if p, _ := effectstest.PtOf(t, g, atWalker); p != 4 {
		t.Errorf("attacking a planeswalker: power %d, want 4", p)
	}
	if !mom5649Has(t, g, atBattle, "menace") {
		t.Error("War-Trained Slasher has no menace")
	}
}

// --- Rampaging Geoderm ------------------------------------------------

func TestMom5649RampagingGeodermPumpsAnAttackerAtAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	geoderm := effectstest.B12Push(g, me.ID, "Rampaging Geoderm", "Creature — Dinosaur Beast", mom5649GeodermOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	for _, kw := range []string{"trample", "haste"} {
		if !mom5649Has(t, g, geoderm, kw) {
			t.Errorf("Rampaging Geoderm has no %s", kw)
		}
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt: whenever you attack")
	}
	if effectstest.HasID(pick.PickTargetCards, idle) || effectstest.HasID(pick.PickTargetCards, geoderm) {
		t.Error("a creature that isn't attacking is not a legal target")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("the target is %d/%d, want 3/3", p, tt)
	}
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 0 {
		t.Errorf("%d +1/+1 counters on a creature attacking a player, want 0", n)
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if p, tt := effectstest.PtOf(t, g, bear); p != 2 || tt != 2 {
		t.Errorf("next turn: %d/%d, want 2/2 (until end of turn)", p, tt)
	}
}

func TestMom5649RampagingGeodermPutsACounterOnAnAttackerAtABattle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Rampaging Geoderm", "Creature — Dinosaur Beast", mom5649GeodermOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	siege := mom5649Battle(g, me.ID, opp.ID)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, d := range []struct{ id, at uuid.UUID }{{bear, siege}, {other, opp.ID}} {
		if err := g.DeclareAttacker(d.id, d.at); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PickCard(t, g, me.ID, bear)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a second target prompt: one attack declaration with two attackers is one trigger")
	}
	if n := len(g.StackMeta); n != 1 {
		t.Fatalf("%d triggers on the stack for one attack declaration with two attackers, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Fatalf("%d +1/+1 counters on a creature attacking a battle, want 1", n)
	}
	if p, tt := effectstest.PtOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("the target is %d/%d, want 3/3: the counter instead of the pump, not both", p, tt)
	}
	if p, _ := effectstest.PtOf(t, g, other); p != 2 {
		t.Errorf("the creature that was not targeted has power %d, want 2", p)
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if p, tt := effectstest.PtOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("next turn: %d/%d, want 3/3 (the counter stays)", p, tt)
	}
}

// "If it's attacking a battle" is read as the ability resolves: a
// creature whose battle left in response is attacking nothing (CR
// 506.4c's shape) and gets the pump.
func TestMom5649RampagingGeodermReadsTheAttackAsItResolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Rampaging Geoderm", "Creature — Dinosaur Beast", mom5649GeodermOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	siege := mom5649Battle(g, me.ID, opp.ID)
	effectstest.DeclareAttack(t, g, siege, bear)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.KillCreature(t, g, me.ID, siege)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 0 {
		t.Errorf("%d +1/+1 counters though the battle was gone as the ability resolved, want 0", n)
	}
	if p, tt := effectstest.PtOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("the target is %d/%d, want 3/3 from the pump", p, tt)
	}
}
