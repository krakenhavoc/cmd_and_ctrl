package effects

import (
	"testing"

	"github.com/google/uuid"

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

func pushDevourFodder(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Goblin",
		Power: 1, Toughness: 1, Owner: owner, Controller: owner,
	})
}

// castDevourer casts the creature and passes priority until its entry
// asks (or it lands).
func castDevourer(t *testing.T, g *game.Game, name, oracle string) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, name, "Creature", oracle, nil)
	passPriorityAroundTable(t, g)
	return id
}

func devourPlusOnes(g *game.Game, id uuid.UUID) int {
	c, ok := battlefieldCard(g, id)
	if !ok {
		return -1
	}
	return c.Counters[game.CounterPlusOne]
}

func TestDevourZeroEntersWithNoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fodder := pushDevourFodder(g, me.ID, "Goblin A")

	id := castDevourer(t, g, "Thunder-Thrash Elder", devourThunderThrashOracle)
	c := entryChoiceOfKind(g, game.PendingChoiceEntrySacrifice, me.ID)
	if c == nil {
		t.Fatal("devour did not ask")
	}
	if c.ChooseMin != 0 || c.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1 (any number, zero included)", c.ChooseMin, c.ChooseMax)
	}
	if _, ok := battlefieldCard(g, id); ok {
		t.Fatal("the creature entered before the choice was made")
	}
	answerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID)
	if got := devourPlusOnes(g, id); got != 0 {
		t.Errorf("counters = %d, want 0", got)
	}
	if _, ok := battlefieldCard(g, fodder); !ok {
		t.Error("declining devoured the fodder anyway")
	}
	if n := g.DevouredBy(id); n != 0 {
		t.Errorf("devoured = %d, want 0", n)
	}
}

func TestDevourWithNoOtherCreatureJustEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := castDevourer(t, g, "Thunder-Thrash Elder", devourThunderThrashOracle)
	if entryChoiceOfKind(g, game.PendingChoiceEntrySacrifice, me.ID) != nil {
		t.Error("asked a question whose only answer is none")
	}
	if got := devourPlusOnes(g, id); got != 0 {
		t.Errorf("did not enter with no counters (got %d)", got)
	}
}

func TestDevourOneScalesByN(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushDevourFodder(g, me.ID, "Goblin A")
	pushDevourFodder(g, me.ID, "Goblin B")

	id := castDevourer(t, g, "Ravenous Tyrannosaurus", devourTyrannosaurusOracle)
	answerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, a)
	if got := devourPlusOnes(g, id); got != 3 {
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushDevourFodder(g, me.ID, "Goblin A")
	b := pushDevourFodder(g, me.ID, "Goblin B")
	c3 := pushDevourFodder(g, me.ID, "Goblin C")

	id := castDevourer(t, g, "Ravenous Tyrannosaurus", devourTyrannosaurusOracle)
	answerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, a, b, c3)
	if got := devourPlusOnes(g, id); got != 9 {
		t.Errorf("counters = %d, want 9 (devour 3 x 3)", got)
	}
	if n := g.DevouredBy(id); n != 3 {
		t.Errorf("devoured = %d, want 3", n)
	}
	for _, f := range []uuid.UUID{a, b, c3} {
		if n := sacrificeEventsFor(g, f); n != 1 {
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
	if _, ok := battlefieldCard(g, id); !ok {
		t.Fatal("the devourer did not enter")
	}
}

// CR 614.1c: the counters are part of the entry, so a doubler sees them.
func TestDevourCountersAreDoubledByDoublingSeason(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	_ = seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
	a := pushDevourFodder(g, me.ID, "Goblin A")
	b := pushDevourFodder(g, me.ID, "Goblin B")

	id := castDevourer(t, g, "Gorger Wurm", devourGorgerWurmOracle)
	answerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, a, b)
	if got := devourPlusOnes(g, id); got != 4 {
		t.Errorf("counters = %d, want 4 (devour 1 x 2 creatures, doubled)", got)
	}
	if n := g.DevouredBy(id); n != 2 {
		t.Errorf("devoured = %d, want 2 (the number of creatures, not counters)", n)
	}
}

// CR 702.82b: "for each creature it devoured" reads the recorded count.
func TestSkullmulcherDrawsForEachCreatureDevoured(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 5; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Land"})
	}
	a := pushDevourFodder(g, me.ID, "Goblin A")
	b := pushDevourFodder(g, me.ID, "Goblin B")
	id := castDevourer(t, g, "Skullmulcher", devourSkullmulcherOracle)
	handBefore := me.Hand.Size()
	answerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, a, b)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
	if got := devourPlusOnes(g, id); got != 2 {
		t.Errorf("counters = %d, want 2", got)
	}
}

func TestMarrowChomperGainsLifePerCreatureDevoured(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushDevourFodder(g, me.ID, "Goblin A")
	b := pushDevourFodder(g, me.ID, "Goblin B")
	c3 := pushDevourFodder(g, me.ID, "Goblin C")
	before := me.Life

	id := castDevourer(t, g, "Marrow Chomper", devourMarrowChomperOracle)
	answerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, a, b, c3)
	passPriorityAroundTable(t, g)
	if got := me.Life - before; got != 6 {
		t.Errorf("gained %d life, want 6", got)
	}
	if got := devourPlusOnes(g, id); got != 6 {
		t.Errorf("counters = %d, want 6", got)
	}
}

func TestPredatorDragonKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Predator Dragon", TypeLine: "Creature — Dragon",
		OracleID: devourPredatorOracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	abs := effectiveAbilities(t, g, id)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	src := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Thunder-Thrash Elder", TypeLine: "Creature — Lizard Warrior",
		OracleID: devourThunderThrashOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	a := pushDevourFodder(g, me.ID, "Goblin A")

	item := &game.StackItem{Controller: me.ID}
	err := CreateTokenCopy{Controller: me.ID, Copy: src, N: 1}.Apply(NewContext(g, item))
	if err != nil {
		t.Fatalf("CreateTokenCopy: %v", err)
	}
	c := entryChoiceOfKind(g, game.PendingChoiceEntrySacrifice, me.ID)
	if c == nil {
		t.Fatal("a token copy of a devour creature did not ask")
	}
	answerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, a)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rex := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ravenous Tyrannosaurus", TypeLine: "Creature — Dinosaur",
		OracleID: devourTyrannosaurusOracle, Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	victim := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	before := opp.Life
	item := &game.StackItem{
		Kind: game.StackItemTriggered, Controller: me.ID, SourceCardID: rex,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}
	spec, _ := Lookup(devourTyrannosaurusOracle)
	if len(spec.Triggered) != 1 {
		t.Fatalf("triggers = %d, want 1", len(spec.Triggered))
	}
	if err := spec.Triggered[0].Effect(g, item); err != nil {
		t.Fatalf("effect: %v", err)
	}
	// State-based actions have not run here; the lethal damage is marked.
	if v, ok := battlefieldCard(g, victim); ok && v.DamageMarked != 2 {
		t.Errorf("victim has %d damage marked, want exactly its lethal 2", v.DamageMarked)
	}
	if got := before - opp.Life; got != 4 {
		t.Errorf("controller lost %d, want 4 (6 power less 2 lethal)", got)
	}
}
