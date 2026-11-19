package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82g23Settler   = "04dd449f-ff3a-4ed7-8241-244d58b371dc"
	s82g23Velukan   = "667aef37-a47a-4593-a252-807c02416863"
	s82g23Gerrard   = "8a1e8435-16f2-475b-a8a9-7b9f0cf4594a"
	s82g23Recruiter = "792e1072-00bf-4f63-b956-5f20cdf09f11"
)

func TestS82GapB23EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82g23Settler: "Goblin Settler", s82g23Velukan: "Velukan Dragon",
		s82g23Gerrard: "Gerrard, Weatherlight Hero", s82g23Recruiter: "Goblin Recruiter",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82GapB23GoblinSettlerDestroysTargetLandAndNothingElse(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := jmpB7Seats(g)
	forest := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Goblin Settler", "Creature — Goblin", s82g23Settler, nil)
	passPriorityAroundTable(t, g)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: forest}, true)
	if g.Battlefield.Contains(forest) || !opp.Graveyard.Contains(forest) {
		t.Error("the targeted land was not destroyed")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("a nonland permanent was destroyed")
	}
}

func TestS82GapB23GoblinSettlerCantTargetANonland(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := jmpB7Seats(g)
	b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Goblin Settler", "Creature — Goblin", s82g23Settler, nil)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(pick.ID, pick.Chooser, game.TargetRef{Kind: game.TargetCard, ID: rock}); err == nil {
		t.Error("an artifact was accepted as 'target land'")
	}
}

func TestS82GapB23VelukanDragonPumpsByTheRollMinusOne(t *testing.T) {
	for _, tc := range []struct {
		name  string
		want  func(int) bool
		power int
	}{{"six", func(r int) bool { return r == 6 }, 10}, {"one", func(r int) bool { return r == 1 }, 5}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			// The roll is drawn for the Dragon as its source; find a key that shows the wanted face.
			dragon := uuid.New()
			key := d5681Key(t, dragon, 6, 1, func(r []int) bool { return tc.want(r[0]) })
			g.SetRNGKeyForTest(key)
			pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: dragon, Name: "Velukan Dragon", OracleID: s82g23Velukan,
				TypeLine: "Creature — Dragon", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID, Keywords: []string{"flying", "haste"}})
			declareAttack(t, g, opp.ID, dragon)
			passPriorityAroundTable(t, g)
			if p, tough := ptOf(t, g, dragon); p != tc.power || tough != 5 {
				t.Errorf("Dragon is %d/%d, want %d/5", p, tough, tc.power)
			}
		})
	}
}

func TestS82GapB23GerrardExilesItselfAndReturnsThisTurnsFallenArtifactsAndCreatures(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	gerrard := b12Push(g, me.ID, "Gerrard, Weatherlight Hero", "Legendary Creature — Human Soldier", s82g23Gerrard, 3, 3)
	fallenCreature := b12Creature(g, me.ID, "Fallen Bear", "Creature — Bear", 2, 2)
	fallenArtifact := b12Permanent(g, me.ID, "Fallen Rock", "Artifact")
	fallenLand := b12Permanent(g, me.ID, "Fallen Forest", "Basic Land — Forest")
	// A creature card discarded this turn, and one that has sat in the graveyard (no event this turn).
	discarded := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: discarded, Name: "Discarded Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, discarded); err != nil {
		t.Fatal(err)
	}
	old := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: old, Name: "Old Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	for _, id := range []uuid.UUID{fallenCreature, fallenArtifact, fallenLand} {
		killCreature(t, g, me.ID, id)
	}
	killCreature(t, g, me.ID, gerrard)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(gerrard) {
		t.Error("Gerrard was not exiled")
	}
	for _, id := range []uuid.UUID{fallenCreature, fallenArtifact} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s did not return", id)
		}
	}
	for name, id := range map[string]uuid.UUID{"land": fallenLand, "discarded": discarded, "old": old} {
		if g.Battlefield.Contains(id) || !me.Graveyard.Contains(id) {
			t.Errorf("the %s card should have stayed in the graveyard", name)
		}
	}
}

func TestS82GapB23GoblinRecruiterPutsAnyNumberOfGoblinsOnTopInOrder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	var goblins []uuid.UUID
	for _, name := range []string{"Alpha", "Beta", "Gamma"} {
		goblins = append(goblins, pushLibraryCardForTest(me, game.Card{Name: name, TypeLine: "Creature — Goblin"}))
	}
	pushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	castCatalogSpell(t, g, "Goblin Recruiter", "Creature — Goblin", s82g23Recruiter, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	if searchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("the search offered a non-Goblin")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, goblins); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	order := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID)
	if order == nil {
		t.Fatal("no prompt to order the Goblins")
	}
	want := []uuid.UUID{goblins[2], goblins[0], goblins[1]}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, want); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := ravB3LibraryTop(me, 3); !ravB3SameOrder(got, want) {
		t.Fatalf("library top %v, want %v", got, want)
	}
}
