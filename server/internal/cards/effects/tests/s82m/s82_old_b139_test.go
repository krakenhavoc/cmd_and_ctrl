package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b139_test.go — Game Night 2019, S82 set OLD batch 139 (#5515).
// Every test plays the real catalog entry through the engine.

const (
	oldB139SpideryGrasp  = "021ce403-cbfa-4d69-a6e9-473d58bd477a"
	oldB139AvenWindMage  = "ad5f9a18-89a5-4bcb-9cd7-5399ffd252d3"
	oldB139TatteredMummy = "066e7379-d439-41cb-a8c9-60964040b110"
	oldB139SteppeGlider  = "85eddd5e-1117-4d58-ac47-ba651050550f"
	oldB139AccursedHorde = "8f3cba20-e7a4-47f4-8bce-90168ce9cc5c"
	oldB139EngulfShore   = "a3d4ed7c-c586-4581-96dc-977636d27583"
)

func TestOldB139EveryCardIsRegisteredWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB139SpideryGrasp: "Spidery Grasp", oldB139AvenWindMage: "Aven Wind Mage",
		oldB139TatteredMummy: "Tattered Mummy", oldB139SteppeGlider: "Steppe Glider",
		oldB139AccursedHorde: "Accursed Horde", oldB139EngulfShore: "Engulf the Shore",
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

func oldB139Ref(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

func TestOldB139SpideryGraspUntapsAndPumpsTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.TapForTest(t, g, bear)
	effectstest.CastCatalogSpell(t, g, "Spidery Grasp", "Instant", oldB139SpideryGrasp, []game.TargetRef{oldB139Ref(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, bear) {
		t.Error("the target is still tapped")
	}
	if p, th := effectstest.PtOf(t, g, bear); p != 4 || th != 6 {
		t.Errorf("the target is %d/%d, want 4/6", p, th)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "reach") {
		t.Error("the target did not gain reach")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, th := effectstest.PtOf(t, g, bear); p != 2 || th != 2 {
		t.Errorf("the boost outlived the turn: %d/%d", p, th)
	}
}

func TestOldB139AvenWindMageGrowsOnInstantsAndSorceriesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mage := effectstest.B12Push(g, me.ID, "Aven Wind Mage", "Creature — Bird Wizard", oldB139AvenWindMage, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.PtOf(t, g, mage); p != 2 || th != 2 {
		t.Fatalf("a creature spell pumped the Mage to %d/%d", p, th)
	}
	effectstest.CastCatalogSpell(t, g, "Opt", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.PtOf(t, g, mage); p != 3 || th != 3 {
		t.Fatalf("an instant gave %d/%d, want 3/3", p, th)
	}
	effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.PtOf(t, g, mage); p != 4 || th != 4 {
		t.Errorf("a second spell gave %d/%d, want 4/4", p, th)
	}
}

func TestOldB139TatteredMummyDrainsEachOpponentWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mummy := effectstest.B12Push(g, me.ID, "Tattered Mummy", "Creature — Zombie Jackal", oldB139TatteredMummy, 1, 2)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{mummy}, game.DestroyOptions{}) })
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID]
		if p.ID != me.ID {
			want -= 2
		}
		if p.Life != want {
			t.Errorf("a player is at %d, want %d", p.Life, want)
		}
	}
}

func TestOldB139SteppeGliderGrantsFlyingAndVigilanceOnlyToCreaturesWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	glider := effectstest.B12Push(g, me.ID, "Steppe Glider", "Creature — Elemental", oldB139SteppeGlider, 2, 4)
	counter := effectstest.B12Creature(g, me.ID, "Counter Bear", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, counter).Counters = map[string]int{game.CounterPlusOne: 1}
	})
	effectstest.ApaMana(me, "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, glider, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{oldB139Ref(plain)}}); err == nil {
		t.Fatal("a creature without a +1/+1 counter was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, glider, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{oldB139Ref(counter)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"flying", "vigilance"} {
		if !effectstest.HasEffectiveKeyword(t, g, counter, kw) {
			t.Errorf("the target did not gain %s", kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, plain, kw) {
			t.Errorf("the creature without a counter gained %s", kw)
		}
	}
}

func TestOldB139AccursedHordeGrantsIndestructibleToAnAttackingZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	horde := effectstest.B12Push(g, me.ID, "Accursed Horde", "Creature — Zombie", oldB139AccursedHorde, 3, 3)
	zombie := effectstest.B12Creature(g, me.ID, "Zombie Token", "Creature — Zombie", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle Zombie", "Creature — Zombie", 2, 2)
	for _, id := range []uuid.UUID{horde, zombie, bear} {
		g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).SummonedThisTurn = false })
	}
	effectstest.DeclareAttack(t, g, opp.ID, zombie, bear)
	for _, id := range []uuid.UUID{bear, idle} {
		effectstest.ApaMana(me, "B", "B")
		if err := g.ActivateCatalogAbility(me.ID, horde, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{oldB139Ref(id)}}); err == nil {
			t.Error("a target that is not an attacking Zombie was accepted")
		}
	}
	effectstest.ApaMana(me, "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, horde, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{oldB139Ref(zombie)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, zombie, "indestructible") {
		t.Error("the attacking Zombie did not gain indestructible")
	}
	if effectstest.HasEffectiveKeyword(t, g, horde, "indestructible") {
		t.Error("the Horde gained indestructible without being the target")
	}
}

func TestOldB139EngulfTheShoreBouncesCreaturesUpToTheNumberOfIslands(t *testing.T) {
	for _, tc := range []struct {
		islands int
		gone    []string
	}{
		{0, nil},
		{2, []string{"Two", "One"}},
		{3, []string{"Two", "One", "Three"}},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		for i := 0; i < tc.islands; i++ {
			effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
		}
		// An Island controlled by an opponent does not count.
		effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
		creatures := map[string]uuid.UUID{
			"One":   effectstest.B12Creature(g, me.ID, "One", "Creature — Bear", 1, 1),
			"Two":   effectstest.B12Creature(g, opp.ID, "Two", "Creature — Bear", 2, 2),
			"Three": effectstest.B12Creature(g, me.ID, "Three", "Creature — Bear", 3, 3),
			"Four":  effectstest.B12Creature(g, opp.ID, "Four", "Creature — Bear", 4, 4),
		}
		effectstest.CastCatalogSpell(t, g, "Engulf the Shore", "Instant", oldB139EngulfShore, nil)
		effectstest.PassPriorityAroundTable(t, g)
		gone := map[string]bool{}
		for _, n := range tc.gone {
			gone[n] = true
		}
		for name, id := range creatures {
			if g.Battlefield.Contains(id) == gone[name] {
				t.Errorf("%d Islands: %s on the battlefield = %v, want %v", tc.islands, name, g.Battlefield.Contains(id), !gone[name])
			}
		}
		for _, name := range tc.gone {
			owner := me
			if name == "Two" {
				owner = opp
			}
			if !owner.Hand.Contains(creatures[name]) {
				t.Errorf("%d Islands: %s is not in its owner's hand", tc.islands, name)
			}
		}
	}
}
