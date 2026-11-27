package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b620_cards_test.go — S82 OLD batch 620 (#5607): Eighth Edition
// mana and ramp, +1/+1 counters, counterspells and bounce, token makers. Every
// test plays the real catalog card through the engine.

const (
	oldB620Elder     = "507bdb1a-90b4-4fd3-a1a3-e8be316e97f3"
	oldB620Trellis   = "57de8fe7-3d1b-41cd-8354-38aff3d2d052"
	oldB620Fungusaur = "3e771f5b-2de3-4a1b-8281-ab1f7491e5a1"
	oldB620Apprent   = "863ad4b8-094b-46b3-b6cd-0640e1f01d6d"
	oldB620Egg       = "98116aec-2ab1-4bee-b727-9feff6274825"
)

func TestOldB620CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB620Elder: "Fyndhorn Elder", oldB620Trellis: "Vine Trellis", oldB620Fungusaur: "Fungusaur",
		oldB620Apprent: "Daring Apprentice", oldB620Egg: "Rukh Egg",
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

func oldB620PoolColors(g *game.Game, p *game.Player) []string {
	var out []string
	g.ReadSnapshot(func() { out = effectstest.Batch01PoolColors(p) })
	return out
}

func TestOldB620FyndhornElderTapsForTwoGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elder := effectstest.B12Push(g, me.ID, "Fyndhorn Elder", "Creature — Elf Druid", oldB620Elder, 1, 1)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := g.ActivateManaAbility(me.ID, elder, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if cols := oldB620PoolColors(g, me); len(cols) != 2 || cols[0] != "G" || cols[1] != "G" {
		t.Errorf("pool %v, want {G}{G}", cols)
	}
	if err := g.ActivateManaAbility(me.ID, elder, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a tapped Elder made mana again")
	}
}

func TestOldB620VineTrellisIsADefenderThatTapsForOneGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wall := effectstest.B12Push(g, me.ID, "Vine Trellis", "Creature — Plant Wall", oldB620Trellis, 0, 4)
	if !effectstest.HasEffectiveKeyword(t, g, wall, "defender") {
		t.Error("Vine Trellis has no defender")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := g.ActivateManaAbility(me.ID, wall, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if cols := oldB620PoolColors(g, me); len(cols) != 1 || cols[0] != "G" {
		t.Errorf("pool %v, want {G}", cols)
	}
}

func TestOldB620FungusaurGrowsEachTimeItIsDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fungus := effectstest.B12Push(g, me.ID, "Fungusaur", "Creature — Fungus Dinosaur", oldB620Fungusaur, 2, 2)
	src := effectstest.B12Creature(g, g.Seats[1].ID, "Pinger", "Creature — Human", 1, 1)
	effectstest.DealDamageTo(t, g, src, fungus, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, fungus, "+1/+1"); n != 1 {
		t.Fatalf("%d +1/+1 counters after the first damage, want 1", n)
	}
	effectstest.DealDamageTo(t, g, src, fungus, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, fungus, "+1/+1"); n != 2 {
		t.Errorf("%d +1/+1 counters after the second damage, want 2", n)
	}
	if p, tough := effectstest.PtOf(t, g, fungus); p != 4 || tough != 4 {
		t.Errorf("Fungusaur is %d/%d, want 4/4", p, tough)
	}
}

func TestOldB620DaringApprenticeCountersASpellBySacrificingItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	apprentice := effectstest.B12Push(g, me.ID, "Daring Apprentice", "Creature — Human Wizard", oldB620Apprent, 1, 1)
	spell := effectstest.Batch01OpponentCasts(t, g, opp, "Cantrip", "", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, apprentice, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(apprentice) || !me.Graveyard.Contains(apprentice) {
		t.Error("the Apprentice was not sacrificed")
	}
	if !opp.Graveyard.Contains(spell) {
		t.Error("the spell was not countered into its owner's graveyard")
	}
}

func TestOldB620DaringApprenticeCannotActivateWithSummoningSickness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	apprentice := effectstest.PushCatalogPermanent(g, me.ID, "Daring Apprentice", "Creature — Human Wizard", oldB620Apprent, true)
	spell := effectstest.Batch01OpponentCasts(t, g, opp, "Cantrip", "", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, apprentice, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}}}); err == nil {
		t.Error("a summoning-sick Apprentice paid {T}")
	}
}

func TestOldB620RukhEggMakesTheBirdAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	egg := effectstest.B12Push(g, me.ID, "Rukh Egg", "Creature — Bird Egg", oldB620Egg, 0, 3)
	effectstest.P1Destroy(t, g, egg)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(egg) || !me.Graveyard.Contains(egg) {
		t.Fatal("the Egg did not die")
	}
	if n := effectstest.NamedOnBattlefield(g, "Bird", me.ID); n != 0 {
		t.Fatalf("%d Birds before the end step, want none", n)
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.NamedOnBattlefield(g, "Bird", me.ID); n != 1 {
		t.Fatalf("%d Birds at the end step, want 1", n)
	}
	var bird uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Bird" && c.Controller == me.ID {
				bird = c.InstanceID
			}
		}
	})
	if p, tough := effectstest.PtOf(t, g, bird); p != 4 || tough != 4 {
		t.Errorf("the Bird is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bird, "flying") {
		t.Error("the Bird has no flying")
	}
	if cols := effectstest.EffectiveColors(t, g, bird); len(cols) != 1 || cols[0] != "R" {
		t.Errorf("the Bird is %v, want red", cols)
	}
}

func TestOldB620RukhEggMakesNothingWhileItLives(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Rukh Egg", "Creature — Bird Egg", oldB620Egg, 0, 3)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.NamedOnBattlefield(g, "Bird", me.ID); n != 0 {
		t.Errorf("%d Birds with the Egg alive, want none", n)
	}
}
