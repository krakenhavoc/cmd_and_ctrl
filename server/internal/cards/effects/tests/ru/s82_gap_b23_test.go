package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82g23Velukan   = "667aef37-a47a-4593-a252-807c02416863"
	s82g23Recruiter = "792e1072-00bf-4f63-b956-5f20cdf09f11"
)

func TestS82GapB23EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82g23Velukan:   "Velukan Dragon",
		s82g23Recruiter: "Goblin Recruiter",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82GapB23VelukanDragonPumpsByTheRollMinusOne(t *testing.T) {
	for _, tc := range []struct {
		name  string
		want  func(int) bool
		power int
	}{{"six", func(r int) bool { return r == 6 }, 10}, {"one", func(r int) bool { return r == 1 }, 5}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			// The roll is drawn for the Dragon as its source; find a key that shows the wanted face.
			dragon := uuid.New()
			key := effectstest.D5681Key(t, dragon, 6, 1, func(r []int) bool { return tc.want(r[0]) })
			g.SetRNGKeyForTest(key)
			effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: dragon, Name: "Velukan Dragon", OracleID: s82g23Velukan,
				TypeLine: "Creature — Dragon", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID, Keywords: []string{"flying", "haste"}})
			effectstest.DeclareAttack(t, g, opp.ID, dragon)
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.PtOf(t, g, dragon); p != tc.power || tough != 5 {
				t.Errorf("Dragon is %d/%d, want %d/5", p, tough, tc.power)
			}
		})
	}
}

func TestS82GapB23GoblinRecruiterPutsAnyNumberOfGoblinsOnTopInOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	var goblins []uuid.UUID
	for _, name := range []string{"Alpha", "Beta", "Gamma"} {
		goblins = append(goblins, effectstest.PushLibraryCardForTest(me, game.Card{Name: name, TypeLine: "Creature — Goblin"}))
	}
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	effectstest.CastCatalogSpell(t, g, "Goblin Recruiter", "Creature — Goblin", s82g23Recruiter, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	if effectstest.SearchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("the search offered a non-Goblin")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, goblins); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	order := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID)
	if order == nil {
		t.Fatal("no prompt to order the Goblins")
	}
	want := []uuid.UUID{goblins[2], goblins[0], goblins[1]}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, want); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := effectstest.RavB3LibraryTop(me, 3); !effectstest.RavB3SameOrder(got, want) {
		t.Fatalf("library top %v, want %v", got, want)
	}
}
