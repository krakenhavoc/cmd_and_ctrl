package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b470_test.go — Scars of Mirrodin charge-counter artifacts, S82 OLD
// batch 470 (#5577): eight cards, each driven through the engine.

const (
	b470Corruption  = "0da19cd4-7b3f-47bb-bb35-cbcc88ab53cd"
	b470Mending     = "17b1c870-ef94-49af-8f9b-a8765f1b54db"
	b470Rage        = "3b172be8-1bac-441e-98cc-f5af8da6005b"
	b470Infestation = "6af2b898-7ba0-4641-8234-5171397cf602"
	b470Thought     = "8a179d68-c9e6-4f46-99bf-11b2a2708990"
	b470Mass        = "98066238-28f8-41e3-903a-d367927f23ca"
	b470Urn         = "ddb19cac-20b0-4ab8-a31d-c5c155491780"
	b470Censer      = "f1feac23-8ba3-440c-828c-590f88008c31"
)

func TestS82OldB470EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b470Corruption: "Trigon of Corruption", b470Mending: "Trigon of Mending", b470Rage: "Trigon of Rage",
		b470Infestation: "Trigon of Infestation", b470Thought: "Trigon of Thought", b470Mass: "Chimeric Mass",
		b470Urn: "Golden Urn", b470Censer: "Necrogen Censer",
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

// b470Cast casts an artifact from the active seat's hand with the given X and
// lets it settle, so its entry counters are put on by the real cast.
func b470Cast(t *testing.T, g *game.Game, name, manaCost, oracle string, x int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Artifact", ManaCost: manaCost, OracleID: oracle,
		Owner: me.ID, Controller: me.ID})
	effectstest.ToMainForCost(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: x}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func b470Untap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).Tapped = false })
}

func TestS82OldB470TrigonsEnterWithThreeCountersAndRefill(t *testing.T) {
	for _, tc := range []struct {
		name, cost, oracle, mana string
	}{
		{"Trigon of Corruption", "{4}", b470Corruption, "B"},
		{"Trigon of Mending", "{2}", b470Mending, "W"},
		{"Trigon of Rage", "{2}", b470Rage, "R"},
		{"Trigon of Infestation", "{4}", b470Infestation, "G"},
		{"Trigon of Thought", "{5}", b470Thought, "U"},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		id := b470Cast(t, g, tc.name, tc.cost, tc.oracle, 0)
		if n := effectstest.CounterCount(g, id, game.CounterCharge); n != 3 {
			t.Fatalf("%s entered with %d charge counters, want 3", tc.name, n)
		}
		// {c}{c}, {T}: put a charge counter on it; one colour of mana is not enough.
		if err := effectstest.RavB2Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{}, tc.mana); err == nil {
			t.Fatalf("%s refilled for one mana", tc.name)
		}
		effectstest.RavB2MustActivate(t, g, me.ID, id, 0, game.ActivateAbilityParams{}, tc.mana)
		if n := effectstest.CounterCount(g, id, game.CounterCharge); n != 4 {
			t.Errorf("%s has %d charge counters after the refill, want 4", tc.name, n)
		}
	}
}

func TestS82OldB470TrigonOfCorruptionShrinksACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	id := b470Cast(t, g, "Trigon of Corruption", "{4}", b470Corruption, 0)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.RavB2MustActivate(t, g, me.ID, id, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}, "C", "C")
	if n := effectstest.CounterCount(g, bear, game.CounterMinusOne); n != 1 {
		t.Errorf("-1/-1 counters on the target = %d, want 1", n)
	}
	if n := effectstest.CounterCount(g, id, game.CounterCharge); n != 2 {
		t.Errorf("charge counters left = %d, want 2", n)
	}
}

func TestS82OldB470TrigonOfMendingGainsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := b470Cast(t, g, "Trigon of Mending", "{2}", b470Mending, 0)
	life := me.Life
	effectstest.RavB2MustActivate(t, g, me.ID, id, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(me.ID)}}, "C", "C")
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
}

func TestS82OldB470TrigonOfRageGivesPlusThreeUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := b470Cast(t, g, "Trigon of Rage", "{2}", b470Rage, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.RavB2MustActivate(t, g, me.ID, id, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}, "C", "C")
	if c := effectstest.LayeredCard(t, g, bear); c.CurrentPower() != 5 || c.CurrentToughness() != 2 {
		t.Errorf("%d/%d, want 5/2", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.DcCleanup(g)
	if c := effectstest.LayeredCard(t, g, bear); c.CurrentPower() != 2 {
		t.Errorf("power %d after the turn, want 2", c.CurrentPower())
	}
}

func TestS82OldB470TrigonOfInfestationMakesAnInfectInsect(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := b470Cast(t, g, "Trigon of Infestation", "{4}", b470Infestation, 0)
	effectstest.RavB2MustActivate(t, g, me.ID, id, 1, game.ActivateAbilityParams{}, "C", "C")
	var found bool
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.Name == "Phyrexian Insect" {
			found = c.IsCreature() && c.CurrentPower() == 1 && c.CurrentToughness() == 1 && effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "infect")
		}
	}
	if !found {
		t.Error("no 1/1 Phyrexian Insect creature with infect")
	}
}

func TestS82OldB470TrigonOfThoughtDrawsACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := b470Cast(t, g, "Trigon of Thought", "{5}", b470Thought, 0)
	hand := me.Hand.Size()
	effectstest.RavB2MustActivate(t, g, me.ID, id, 1, game.ActivateAbilityParams{}, "C", "C")
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestS82OldB470TrigonsRefuseWithoutACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B12Push(g, me.ID, "Trigon of Thought", "Artifact", b470Thought, 0, 0)
	if err := effectstest.RavB2Activate(t, g, me.ID, id, 1, game.ActivateAbilityParams{}, "C", "C"); err == nil {
		t.Error("the draw ability was paid with no charge counter on it")
	}
}

func TestS82OldB470NecrogenCenserDrainsTwiceThenIsSpent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	id := b470Cast(t, g, "Necrogen Censer", "{3}", b470Censer, 0)
	if n := effectstest.CounterCount(g, id, game.CounterCharge); n != 2 {
		t.Fatalf("entered with %d charge counters, want 2", n)
	}
	life := opp.Life
	for i := 0; i < 2; i++ {
		effectstest.RavB2MustActivate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
		b470Untap(g, id)
	}
	if opp.Life != life-4 {
		t.Errorf("life %d → %d, want -4", life, opp.Life)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err == nil {
		t.Error("a third activation was allowed with no counter left")
	}
}

func TestS82OldB470GoldenUrnMayPutACounterOnInYourUpkeep(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		urn := effectstest.B12Push(g, me.ID, "Golden Urn", "Artifact", b470Urn, 0, 0)
		effectstest.AdvanceToUpkeepOf(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, accept)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if accept {
			want = 1
		}
		if n := effectstest.CounterCount(g, urn, game.CounterCharge); n != want {
			t.Errorf("accept=%v: charge counters = %d, want %d", accept, n, want)
		}
	}
}

func TestS82OldB470GoldenUrnGainsLifePerCounterWhenSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	urn := effectstest.B12Push(g, me.ID, "Golden Urn", "Artifact", b470Urn, 0, 0)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(urn, game.CounterCharge, 3); err != nil {
			t.Errorf("AddCounterForEffect: %v", err)
		}
	})
	life := me.Life
	effectstest.RavB2MustActivate(t, g, me.ID, urn, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(urn) {
		t.Error("the Urn was not sacrificed")
	}
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3 for three counters", life, me.Life)
	}
}

func TestS82OldB470ChimericMassEntersWithXAndAnimates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := b470Cast(t, g, "Chimeric Mass", "{X}", b470Mass, 3)
	if n := effectstest.CounterCount(g, id, game.CounterCharge); n != 3 {
		t.Fatalf("entered with %d charge counters, want 3", n)
	}
	if c := effectstest.LayeredCard(t, g, id); c.IsCreature() {
		t.Fatal("a creature before it is animated")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, id, 0, game.ActivateAbilityParams{}, "C")
	c := effectstest.LayeredCard(t, g, id)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Construct") || c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Fatalf("animated: creature=%v artifact=%v %d/%d, want a 3/3 Construct artifact creature", c.IsCreature(), c.IsArtifact(), c.CurrentPower(), c.CurrentToughness())
	}
	// More counters make it bigger as they arrive.
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, game.CounterCharge, 2); err != nil {
			t.Errorf("AddCounterForEffect: %v", err)
		}
	})
	if c := effectstest.LayeredCard(t, g, id); c.CurrentPower() != 5 || c.CurrentToughness() != 5 {
		t.Errorf("with five counters it is %d/%d, want 5/5", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.DcCleanup(g)
	if c := effectstest.LayeredCard(t, g, id); c.IsCreature() {
		t.Error("still a creature after the turn ended")
	}
}
