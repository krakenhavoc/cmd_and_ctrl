package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// primeval_spawn_test.go — Primeval Spawn (#2017): the entry clause, and the
// ten exiled cards cast against one running total mana value of 10.

const primevalSpawnOracle = "a11c6601-7e91-4def-bfea-bb34023ef3c6"

// spawnLeaves puts Primeval Spawn onto the battlefield, destroys it and
// settles the leaves-the-battlefield trigger, leaving the exile offered.
func spawnLeaves(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	spawn := pushCatalogPermanent(g, me.ID, "Primeval Spawn", "Creature — Avatar", primevalSpawnOracle, false)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(spawn); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	return spawn
}

func spawnCast(g *game.Game, me *game.Player, id uuid.UUID) error {
	return g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile"})
}

func TestPrimevalSpawnExilesTenAndSharesOneBudgetOfTen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	// The top ten: mana values 9 four times, then 6, 4, 3, 1, 0 and a land.
	// Two fillers underneath stay where they are.
	me.Library.Cards = nil
	byName := map[string]uuid.UUID{}
	push := func(name, typeLine, cost string) {
		byName[name] = pushLibraryCardForTest(me, game.Card{Name: name, TypeLine: typeLine, ManaCost: cost})
	}
	for _, n := range []string{"V", "W", "X", "Y"} {
		push("Nine "+n, "Artifact", "{9}")
	}
	push("Six", "Artifact", "{6}")
	push("Four", "Artifact", "{4}")
	push("Three", "Artifact", "{3}")
	push("One", "Artifact", "{1}")
	push("Zero", "Artifact", "{0}")
	push("Forest", "Basic Land — Forest", "")
	// pushLibraryCardForTest pushes to the BOTTOM, so what is pushed last
	// is deepest: these two are below the top ten.
	push("Filler One", "Artifact", "{2}")
	push("Filler Two", "Artifact", "{2}")

	spawnLeaves(t, g, me)
	for _, n := range []string{"Six", "Four", "Three", "One", "Zero", "Forest", "Nine V", "Nine W", "Nine X", "Nine Y"} {
		if !inExile(g, byName[n]) {
			t.Errorf("%s was among the top ten and is not in exile", n)
		}
	}
	for _, n := range []string{"Filler One", "Filler Two"} {
		if inExile(g, byName[n]) || !me.Library.Contains(byName[n]) {
			t.Errorf("%s was below the top ten and left the library", n)
		}
	}

	// 6 + 3 = 9 fits, leaving 1: the 4 is refused, the 1 is not, the 0 is.
	if err := spawnCast(g, me, byName["Six"]); err != nil {
		t.Fatalf("cast the mana value 6 card: %v", err)
	}
	if err := spawnCast(g, me, byName["Three"]); err != nil {
		t.Fatalf("cast the mana value 3 card (total 9): %v", err)
	}
	if err := spawnCast(g, me, byName["Four"]); !errors.Is(err, game.ErrSpellManaValueTooHigh) {
		t.Fatalf("cast the mana value 4 card at total 9: err = %v, want ErrSpellManaValueTooHigh", err)
	}
	if err := spawnCast(g, me, byName["Nine V"]); !errors.Is(err, game.ErrSpellManaValueTooHigh) {
		t.Fatalf("cast a mana value 9 card at total 9: err = %v, want ErrSpellManaValueTooHigh", err)
	}
	if err := spawnCast(g, me, byName["One"]); err != nil {
		t.Fatalf("cast the mana value 1 card (total 10): %v", err)
	}
	if err := spawnCast(g, me, byName["Zero"]); err != nil {
		t.Fatalf("cast the mana value 0 card at total 10 (nothing left but 0 fits): %v", err)
	}
}

func TestPrimevalSpawnTotalIsOrderIndependent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	four := pushLibraryCardForTest(me, game.Card{Name: "Four", TypeLine: "Artifact", ManaCost: "{4}"})
	six := pushLibraryCardForTest(me, game.Card{Name: "Six", TypeLine: "Artifact", ManaCost: "{6}"})
	one := pushLibraryCardForTest(me, game.Card{Name: "One", TypeLine: "Artifact", ManaCost: "{1}"})
	spawnLeaves(t, g, me)
	// 4 + 6 = 10 exactly, in the other order from "biggest first".
	if err := spawnCast(g, me, four); err != nil {
		t.Fatalf("cast the 4 first: %v", err)
	}
	if err := spawnCast(g, me, six); err != nil {
		t.Fatalf("cast the 6 (total 10): %v", err)
	}
	if err := spawnCast(g, me, one); !errors.Is(err, game.ErrSpellManaValueTooHigh) {
		t.Fatalf("a third card past the total: err = %v, want ErrSpellManaValueTooHigh", err)
	}
}

// A land among the ten is exiled and cannot be cast: the grant is for
// spells, and playing a land is not casting.
func TestPrimevalSpawnLandIsExiledAndNotCastable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	forest := pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	spawnLeaves(t, g, me)
	if !inExile(g, forest) {
		t.Fatal("the land among the top ten is exiled")
	}
	if err := spawnCast(g, me, forest); err == nil {
		t.Error("a land was cast")
	}
	if err := g.PlayCard(me.ID, forest); err == nil {
		t.Error("a land was played from exile: the grant is for casting")
	}
}

// The window is the resolving ability's: once the caster passes, the
// permission is gone and the cards not cast stay in exile.
func TestPrimevalSpawnWindowClosesOnPassAndLeavesTheRestInExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	a := pushLibraryCardForTest(me, game.Card{Name: "A", TypeLine: "Artifact", ManaCost: "{2}"})
	b := pushLibraryCardForTest(me, game.Card{Name: "B", TypeLine: "Artifact", ManaCost: "{3}"})
	spawnLeaves(t, g, me)
	if err := spawnCast(g, me, a); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !inExile(g, b) {
		t.Error("the card that was not cast stays in exile (neither bottomed nor put in hand)")
	}
	if err := spawnCast(g, me, b); err == nil {
		t.Error("the card was cast after the window closed")
	}
	if len(me.CastPermissions) != 0 {
		t.Errorf("a permission survived the pass: %+v", me.CastPermissions)
	}
}

// Fewer than ten cards: the budget applies to what exists.
func TestPrimevalSpawnWithAShortLibraryExilesWhatIsThere(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	two := pushLibraryCardForTest(me, game.Card{Name: "Two", TypeLine: "Artifact", ManaCost: "{2}"})
	spawnLeaves(t, g, me)
	if !inExile(g, two) || me.Library.Size() != 0 {
		t.Fatal("the whole short library is exiled")
	}
	if err := spawnCast(g, me, two); err != nil {
		t.Fatalf("cast the only card: %v", err)
	}
}

// An empty library exiles nothing and grants nothing.
func TestPrimevalSpawnWithAnEmptyLibraryDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	spawnLeaves(t, g, me)
	if len(me.CastPermissions) != 0 {
		t.Errorf("permissions = %+v, want none", me.CastPermissions)
	}
}

// What a bot picks from follows the running total, as the gate does.
func TestPrimevalSpawnEnumeratorFollowsTheRunningTotal(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	six := pushLibraryCardForTest(me, game.Card{Name: "Six", TypeLine: "Artifact", ManaCost: "{6}"})
	five := pushLibraryCardForTest(me, game.Card{Name: "Five", TypeLine: "Artifact", ManaCost: "{5}"})
	four := pushLibraryCardForTest(me, game.Card{Name: "Four", TypeLine: "Artifact", ManaCost: "{4}"})
	spawnLeaves(t, g, me)

	offered := func(id uuid.UUID) bool {
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Kind == legal.KindCast && strings.Contains(string(m.Params), id.String()) {
				return true
			}
		}
		return false
	}
	for name, id := range map[string]uuid.UUID{"Six": six, "Five": five, "Four": four} {
		if !offered(id) {
			t.Errorf("%s is under the total of 10 and not offered", name)
		}
	}
	if err := spawnCast(g, me, six); err != nil {
		t.Fatal(err)
	}
	if offered(five) {
		t.Error("the mana value 5 card is offered with 4 left")
	}
	if !offered(four) {
		t.Error("the mana value 4 card is not offered with 4 left")
	}
}

// The running total is part of the snapshot.
func TestPrimevalSpawnTotalSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	six := pushLibraryCardForTest(me, game.Card{Name: "Six", TypeLine: "Artifact", ManaCost: "{6}"})
	five := pushLibraryCardForTest(me, game.Card{Name: "Five", TypeLine: "Artifact", ManaCost: "{5}"})
	four := pushLibraryCardForTest(me, game.Card{Name: "Four", TypeLine: "Artifact", ManaCost: "{4}"})
	spawnLeaves(t, g, me)
	if err := spawnCast(g, me, six); err != nil {
		t.Fatal(err)
	}
	r := restoreRoundTrip(t, g, false)
	rme := r.PlayerByIDForEffect(me.ID)
	if err := spawnCast(r, rme, five); !errors.Is(err, game.ErrSpellManaValueTooHigh) {
		t.Fatalf("restored: a 5 with 4 left: err = %v, want ErrSpellManaValueTooHigh", err)
	}
	if err := spawnCast(r, rme, four); err != nil {
		t.Fatalf("restored: a 4 with 4 left: %v", err)
	}
}

// --- the entry clause ----------------------------------------------

func TestPrimevalSpawnEntry(t *testing.T) {
	const typeLine = "Creature — Avatar"
	for _, tc := range []struct {
		name   string
		enters bool
		do     func(t *testing.T, g *game.Game, me *game.Player) uuid.UUID
	}{
		{"cast with its mana", true, func(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
			floatForTest(g, me, "WUBRGWUBRGW")
			return castFromHandForTest(t, g, me, "Primeval Spawn", typeLine, "{5}{W}{U}{B}{R}{G}",
				primevalSpawnOracle, game.CastSpellParams{Strict: true})
		}},
		{"cast without paying its mana cost", false, func(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
			return castFromHandForTest(t, g, me, "Primeval Spawn", typeLine, "{0}",
				primevalSpawnOracle, game.CastSpellParams{Strict: true})
		}},
		{"cast for nothing with mana charging off", false, func(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
			return castFromHandForTest(t, g, me, "Primeval Spawn", typeLine, "{0}",
				primevalSpawnOracle, game.CastSpellParams{})
		}},
		{"put onto the battlefield from a graveyard", false, func(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
			id := gyCard(me, "Primeval Spawn", typeLine, "{5}{W}{U}{B}{R}{G}", primevalSpawnOracle)
			b30Reanimate(t, g, id, me.ID)
			return id
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			me.Library.Cards = nil
			card := pushLibraryCardForTest(me, game.Card{Name: "Top", TypeLine: "Artifact", ManaCost: "{1}"})
			id := tc.do(t, g, me)
			passPriorityAroundTable(t, g)
			onBattlefield := g.Battlefield.Contains(id)
			if onBattlefield != tc.enters {
				t.Fatalf("on the battlefield = %v, want %v", onBattlefield, tc.enters)
			}
			if !tc.enters {
				if !inExile(g, id) {
					t.Error("a Spawn that does not enter is exiled instead")
				}
				// It never entered, so it never left: no ten cards.
				if inExile(g, card) {
					t.Error("an exiled Spawn triggered its leaves-the-battlefield ability")
				}
			}
		})
	}
}

// The budgeted form is its own window: an option that names another one is
// refused rather than quietly overridden.
func TestExileTopWithBudgetRefusesOptionsThatMeanAnotherWindow(t *testing.T) {
	budget := 10
	for name, e := range map[string]ExileTopWithPermission{
		"free":         {Free: true},
		"cast only":    {CastOnly: true},
		"while exiled": {WhileExiled: true},
		"face down":    {FaceDown: true},
	} {
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			top := pushLibraryCardForTest(me, game.Card{Name: "Top", TypeLine: "Artifact", ManaCost: "{1}"})
			e.From, e.GrantTo, e.N, e.TotalManaValueAtMost = me.ID, me.ID, 1, &budget
			var err error
			g.WithWriteLock(func() { err = e.Apply(NewContext(g, &game.StackItem{Controller: me.ID})) })
			if err == nil {
				t.Error("a budgeted exile with another window's option was accepted")
			}
			if inExile(g, top) {
				t.Error("the refused form exiled a card")
			}
		})
	}
}

func TestPrimevalSpawnKeywordsAndCompleteness(t *testing.T) {
	spec, ok := Lookup(primevalSpawnOracle)
	if !ok {
		t.Fatal("Primeval Spawn is not in the catalog")
	}
	want := map[string]bool{"vigilance": true, "trample": true, "lifelink": true}
	for _, k := range spec.PrintedKeywords {
		delete(want, k)
	}
	if len(want) != 0 {
		t.Errorf("missing keywords %v", want)
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Primeval Spawn is complete since a waived payment asks (#3855): %+v", spec.Caveats)
	}
}
