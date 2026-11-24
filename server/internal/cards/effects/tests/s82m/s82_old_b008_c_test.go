package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 008, group C: Hexdrinker and Immard, the Stormcleaver (#5462).

const (
	s82OldB008Hexdrinker = "69bc2afd-9f53-47f2-b9c8-f12732784e10"
	s82OldB008Immard     = "104f65d3-bedb-4dc5-a5dd-8ab9fe796af4"
)

func TestS82OldB008CRegisteredFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{s82OldB008Hexdrinker: "Hexdrinker", s82OldB008Immard: "Immard, the Stormcleaver"} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v %v", name, spec, ok)
		}
	}
}

func s82OldB008Level(t *testing.T, g *game.Game, id uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, game.CounterLevel, n); err != nil {
			t.Fatalf("level counters: %v", err)
		}
	})
}

func s82OldB008Size(t *testing.T, g *game.Game, id uuid.UUID, p, tough int, what string) {
	t.Helper()
	if gp, gt := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); gp != p || gt != tough {
		t.Errorf("%s is %d/%d, want %d/%d", what, gp, gt, p, tough)
	}
}

func TestS82OldB008CHexdrinkerBands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	h := effectstest.B12Push(g, me.ID, "Hexdrinker", "Creature — Snake", s82OldB008Hexdrinker, 2, 1)
	have := 0
	for _, step := range []struct {
		level, p, tough int
		instants, all   bool
	}{{0, 2, 1, false, false}, {2, 2, 1, false, false}, {3, 4, 4, true, false}, {7, 4, 4, true, false}, {8, 6, 6, false, true}} {
		s82OldB008Level(t, g, h, step.level-have)
		have = step.level
		s82OldB008Size(t, g, h, step.p, step.tough, "Hexdrinker")
		if got := effectstest.HasEffectiveKeyword(t, g, h, "protection from instants"); got != (step.instants) {
			t.Errorf("level %d: protection from instants %v, want %v", step.level, got, step.instants)
		}
		if got := effectstest.HasEffectiveKeyword(t, g, h, "protection from everything"); got != step.all {
			t.Errorf("level %d: protection from everything %v, want %v", step.level, got, step.all)
		}
	}
}

func TestS82OldB008CHexdrinkerLevelUpIsSorceryOnlyAndCostsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	h := effectstest.B12Push(g, me.ID, "Hexdrinker", "Creature — Snake", s82OldB008Hexdrinker, 2, 1)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, h, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("level up activated with no mana to pay {1}")
	}
	effectstest.ApaMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, h, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("level up: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, h, game.CounterLevel); got != 1 {
		t.Errorf("%d level counters, want 1", got)
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.ApaMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, h, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("level up activated outside a main phase")
	}
}

func s82OldB008PushImmard(g *game.Game, me *game.Player, counters int) uuid.UUID {
	id := effectstest.B12Push(g, me.ID, "Immard, the Stormcleaver", "Legendary Creature — Human Soldier", s82OldB008Immard, 4, 4)
	if counters > 0 {
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterCharge, counters) })
	}
	return id
}

func s82OldB008Attack(t *testing.T, g *game.Game, me, opp *game.Player, id uuid.UUID) {
	t.Helper()
	effectstest.DeclareAttack(t, g, opp.ID, id)
}

func TestS82OldB008CImmardEntersAndPutsACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.CastCreatureCard(t, g, "Immard, the Stormcleaver", "Legendary Creature — Human Soldier", s82OldB008Immard, 4, 4)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, id, game.CounterCharge); got != 1 {
		t.Errorf("%d charge counters, want 1", got)
	}
}

func TestS82OldB008CImmardRemoveWithNoCounterDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.CastCreatureCard(t, g, "Immard, the Stormcleaver", "Legendary Creature — Human Soldier", s82OldB008Immard, 4, 4)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, id, game.CounterCharge); got != 0 {
		t.Errorf("%d charge counters, want 0", got)
	}
	if c := effectstest.ModePickChoiceFor(g, me.ID); c != nil {
		t.Error("a reflexive mode prompt opened though no counter was removed")
	}
}

func TestS82OldB008CImmardAttackRemovesACounterThenDealsFourDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	id := s82OldB008PushImmard(g, me, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 2, 6)
	s82OldB008Attack(t, g, me, opp, id)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, id, game.CounterCharge); got != 0 {
		t.Fatalf("%d charge counters, want 0", got)
	}
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode prompt for the reflexive trigger")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if v, ok := effectstest.BattlefieldCard(g, victim); !ok || v.DamageMarked != 4 {
		t.Errorf("Ogre damage %+v, want 4 marked", v.DamageMarked)
	}
}

func TestS82OldB008CImmardAttackPutsAnotherCounterOnAndNoReflexive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	id := s82OldB008PushImmard(g, me, 1)
	s82OldB008Attack(t, g, me, opp, id)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, id, game.CounterCharge); got != 2 {
		t.Errorf("%d charge counters, want 2", got)
	}
	if c := effectstest.ModePickChoiceFor(g, me.ID); c != nil {
		t.Error("putting a counter opened the reflexive mode prompt")
	}
}

func TestS82OldB008CImmardSecondModeGrantsLifelinkAndIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	id := s82OldB008PushImmard(g, me, 2)
	s82OldB008Attack(t, g, me, opp, id)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode prompt")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, id, game.CounterCharge); got != 1 {
		t.Errorf("%d charge counters, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "lifelink") || !effectstest.HasEffectiveKeyword(t, g, id, "indestructible") {
		t.Error("Immard lacks lifelink or indestructible")
	}
}
