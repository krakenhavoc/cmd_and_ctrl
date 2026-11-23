package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// devour_test.go — Devour N (CR 702.82a) over devour.go.

const (
	devourTyrannosaurusOracle = "9cfac390-4638-4654-8b50-65c0f0886b18"
	devourThunderThrashOracle = "aef6c5b3-824f-4b7d-9671-1ed61d547717"
	devourSkullmulcherOracle  = "66cdbd34-a864-4c39-acac-9fb26ef4adf9"
	devourMarrowChomperOracle = "96313465-6a97-4c0e-ae9a-91e95f247f5b"
	devourGorgerWurmOracle    = "f2126eb9-892c-4788-95cc-6f3e59f0375d"
	devourPredatorOracle      = "ef048ecc-581d-4a17-91fd-edf7d7733ef9"
)

func TestDevourZeroEntersWithNoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fodder := effectstest.PushDevourFodder(g, me.ID, "Goblin A")

	id := effectstest.CastDevourer(t, g, "Thunder-Thrash Elder", devourThunderThrashOracle)
	c := effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntrySacrifice, me.ID)
	if c == nil {
		t.Fatal("devour did not ask")
	}
	if c.ChooseMin != 0 || c.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1 (any number, zero included)", c.ChooseMin, c.ChooseMax)
	}
	if _, ok := effectstest.BattlefieldCard(g, id); ok {
		t.Fatal("the creature entered before the choice was made")
	}
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID)
	if got := effectstest.DevourPlusOnes(g, id); got != 0 {
		t.Errorf("counters = %d, want 0", got)
	}
	if _, ok := effectstest.BattlefieldCard(g, fodder); !ok {
		t.Error("declining devoured the fodder anyway")
	}
	if n := g.DevouredBy(id); n != 0 {
		t.Errorf("devoured = %d, want 0", n)
	}
}

func TestDevourWithNoOtherCreatureJustEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.CastDevourer(t, g, "Thunder-Thrash Elder", devourThunderThrashOracle)
	if effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntrySacrifice, me.ID) != nil {
		t.Error("asked a question whose only answer is none")
	}
	if got := effectstest.DevourPlusOnes(g, id); got != 0 {
		t.Errorf("did not enter with no counters (got %d)", got)
	}
}

func TestDevourOneScalesByN(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.PushDevourFodder(g, me.ID, "Goblin A")
	effectstest.PushDevourFodder(g, me.ID, "Goblin B")

	id := effectstest.CastDevourer(t, g, "Ravenous Tyrannosaurus", devourTyrannosaurusOracle)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, a)
	if got := effectstest.DevourPlusOnes(g, id); got != 3 {
		t.Errorf("counters = %d, want 3 (devour 3 x 1)", got)
	}
	if n := g.DevouredBy(id); n != 1 {
		t.Errorf("devoured = %d, want 1", n)
	}
	if !me.Graveyard.Contains(a) {
		t.Error("the devoured creature is not in the graveyard")
	}
}

func TestDevourThreeFiresSacrificeEvents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.PushDevourFodder(g, me.ID, "Goblin A")
	b := effectstest.PushDevourFodder(g, me.ID, "Goblin B")
	c3 := effectstest.PushDevourFodder(g, me.ID, "Goblin C")

	id := effectstest.CastDevourer(t, g, "Ravenous Tyrannosaurus", devourTyrannosaurusOracle)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, a, b, c3)
	if got := effectstest.DevourPlusOnes(g, id); got != 9 {
		t.Errorf("counters = %d, want 9 (devour 3 x 3)", got)
	}
	if n := g.DevouredBy(id); n != 3 {
		t.Errorf("devoured = %d, want 3", n)
	}
	for _, f := range []uuid.UUID{a, b, c3} {
		if n := effectstest.SacrificeEventsFor(g, f); n != 1 {
			t.Errorf("%d EventSacrifice for a devoured creature, want 1", n)
		}
		lt := 0
		for _, ev := range g.Events {
			if ev.Kind == game.EventLTB && ev.CardID == f {
				lt++
			}
		}
		if lt != 1 {
			t.Errorf("%d EventLTB for a devoured creature, want 1 (dies triggers need it)", lt)
		}
	}
	// The sacrifice came off the entry, before the creature landed.
	if _, ok := effectstest.BattlefieldCard(g, id); !ok {
		t.Fatal("the devourer did not enter")
	}
}

// CR 614.1c: the counters are part of the entry, so a doubler sees them.
func TestDevourCountersAreDoubledByDoublingSeason(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	_ = effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	a := effectstest.PushDevourFodder(g, me.ID, "Goblin A")
	b := effectstest.PushDevourFodder(g, me.ID, "Goblin B")

	id := effectstest.CastDevourer(t, g, "Gorger Wurm", devourGorgerWurmOracle)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, a, b)
	if got := effectstest.DevourPlusOnes(g, id); got != 4 {
		t.Errorf("counters = %d, want 4 (devour 1 x 2 creatures, doubled)", got)
	}
	if n := g.DevouredBy(id); n != 2 {
		t.Errorf("devoured = %d, want 2 (the number of creatures, not counters)", n)
	}
}

// CR 702.82b: "for each creature it devoured" reads the recorded count.
func TestSkullmulcherDrawsForEachCreatureDevoured(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 5; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Land"})
	}
	a := effectstest.PushDevourFodder(g, me.ID, "Goblin A")
	b := effectstest.PushDevourFodder(g, me.ID, "Goblin B")
	id := effectstest.CastDevourer(t, g, "Skullmulcher", devourSkullmulcherOracle)
	handBefore := me.Hand.Size()
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
	if got := effectstest.DevourPlusOnes(g, id); got != 2 {
		t.Errorf("counters = %d, want 2", got)
	}
}

func TestMarrowChomperGainsLifePerCreatureDevoured(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.PushDevourFodder(g, me.ID, "Goblin A")
	b := effectstest.PushDevourFodder(g, me.ID, "Goblin B")
	c3 := effectstest.PushDevourFodder(g, me.ID, "Goblin C")
	before := me.Life

	id := effectstest.CastDevourer(t, g, "Marrow Chomper", devourMarrowChomperOracle)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, a, b, c3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - before; got != 6 {
		t.Errorf("gained %d life, want 6", got)
	}
	if got := effectstest.DevourPlusOnes(g, id); got != 6 {
		t.Errorf("counters = %d, want 6", got)
	}
}

func TestPredatorDragonKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Predator Dragon", TypeLine: "Creature — Dragon",
		OracleID: devourPredatorOracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	abs := effectstest.EffectiveAbilities(t, g, id)
	has := map[string]bool{}
	for _, a := range abs {
		has[a] = true
	}
	if !has["flying"] || !has["haste"] {
		t.Errorf("abilities %v, want flying and haste", abs)
	}
}

// A token copy of a devour creature asks too (CR 707.2: the copy has
// the ability).
func TestDevourFromATokenCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	src := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Thunder-Thrash Elder", TypeLine: "Creature — Lizard Warrior",
		OracleID: devourThunderThrashOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	a := effectstest.PushDevourFodder(g, me.ID, "Goblin A")

	item := &game.StackItem{Controller: me.ID}
	err := effects.CreateTokenCopy{Controller: me.ID, Copy: src, N: 1}.Apply(effects.NewContext(g, item))
	if err != nil {
		t.Fatalf("CreateTokenCopy: %v", err)
	}
	c := effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntrySacrifice, me.ID)
	if c == nil {
		t.Fatal("a token copy of a devour creature did not ask")
	}
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, a)
	found := false
	for _, bc := range g.Battlefield.Cards {
		if bc.InstanceID != src && bc.Name == "Thunder-Thrash Elder" {
			found = true
			if bc.Counters[game.CounterPlusOne] != 3 {
				t.Errorf("token counters = %d, want 3", bc.Counters[game.CounterPlusOne])
			}
			if bc.Devoured != 1 {
				t.Errorf("token devoured = %d, want 1", bc.Devoured)
			}
		}
	}
	if !found {
		t.Error("the token copy did not enter")
	}
}

// Ravenous Tyrannosaurus's attack trigger: lethal to the creature, the
// excess to its controller.
func TestTyrannosaurusAttackTriggerSplitsExcessToController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rex := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ravenous Tyrannosaurus", TypeLine: "Creature — Dinosaur",
		OracleID: devourTyrannosaurusOracle, Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	victim := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	before := opp.Life
	item := &game.StackItem{
		Kind: game.StackItemTriggered, Controller: me.ID, SourceCardID: rex,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}
	spec, _ := effects.Lookup(devourTyrannosaurusOracle)
	if len(spec.Triggered) != 1 {
		t.Fatalf("triggers = %d, want 1", len(spec.Triggered))
	}
	if err := spec.Triggered[0].Effect(g, item); err != nil {
		t.Fatalf("effect: %v", err)
	}
	// State-based actions have not run here; the lethal damage is marked.
	if v, ok := effectstest.BattlefieldCard(g, victim); ok && v.DamageMarked != 2 {
		t.Errorf("victim has %d damage marked, want exactly its lethal 2", v.DamageMarked)
	}
	if got := before - opp.Life; got != 4 {
		t.Errorf("controller lost %d, want 4 (6 power less 2 lethal)", got)
	}
}
