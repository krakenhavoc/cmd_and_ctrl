package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 7, group C (#5462): Elspeth, Knight-Errant and Tibalt, the
// Fiend-Blooded, driven through the real engine.

const (
	s82OldB007ElspethOracle = "eab0d8b9-c66f-402f-b259-f46f15cefd78"
	s82OldB007TibaltOracle  = "ee9c57c6-41ef-4328-8417-e5a433eadf17"
)

func TestS82OldB007CCardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		s82OldB007ElspethOracle: "Elspeth, Knight-Errant", s82OldB007TibaltOracle: "Tibalt, the Fiend-Blooded",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v (found %v)", name, spec, ok)
		}
	}
}

func TestS82OldB007CElspethPlusOneMakesASoldier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	w := effectstest.S82Walker(g, me.ID, "Elspeth, Knight-Errant", s82OldB007ElspethOracle, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, w, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	var soldier *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Controller == me.ID && c.Name == "Soldier" {
			soldier = c
		}
	}
	if soldier == nil || soldier.Power != 1 || soldier.Toughness != 1 {
		t.Fatalf("no 1/1 Soldier token: %+v", soldier)
	}
	if got := effectstest.LoyaltyOf(g, w); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestS82OldB007CElspethPlusOnePumpsAndGrantsFlyingForTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	w := effectstest.S82Walker(g, me.ID, "Elspeth, Knight-Errant", s82OldB007ElspethOracle, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, w, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, bear)
	if c.CurrentPower() != 5 || c.CurrentToughness() != 5 {
		t.Errorf("Bear is %d/%d, want 5/5", c.CurrentPower(), c.CurrentToughness())
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("no flying")
	}
	if got := effectstest.LoyaltyOf(g, w); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestS82OldB007CElspethRefusesAnIllegalTargetAndTheUltimateWithoutLoyalty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	w := effectstest.S82Walker(g, me.ID, "Elspeth, Knight-Errant", s82OldB007ElspethOracle, 4)
	effectstest.AdvanceToMain(t, g)
	// The −8 needs eight loyalty.
	if err := g.ActivateCatalogAbility(me.ID, w, 2, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("−8 activated at loyalty 4")
	}
	// A land is not a creature.
	land := effectstest.PushCatalogPermanent(g, me.ID, "Plains", "Basic Land — Plains", "", false)
	if err := g.ActivateCatalogAbility(me.ID, w, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.S82TargetCard(land)}); err == nil {
		t.Error("the pump accepted a land")
	}
}

func TestS82OldB007CElspethEmblemMakesOwnersPermanentsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	w := effectstest.S82Walker(g, me.ID, "Elspeth, Knight-Errant", s82OldB007ElspethOracle, 8)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Plains", "Basic Land — Plains", "", false)
	enemy := effectstest.B12Creature(g, opp.ID, "Enemy", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, w, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.EmblemsOf(g, me.ID)) != 1 {
		t.Fatal("no emblem")
	}
	for _, id := range []uuid.UUID{bear, land} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "indestructible") {
			t.Errorf("%v is not indestructible", id)
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, enemy, "indestructible") {
		t.Error("an opponent's creature is indestructible")
	}
	// A permanent that arrives later is covered.
	late := effectstest.PushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
	if !effectstest.HasEffectiveKeyword(t, g, late, "indestructible") {
		t.Error("a later land is not indestructible")
	}
}

func TestS82OldB007CTibaltPlusOneDrawsThenDiscardsAtRandom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	w := effectstest.S82Walker(g, me.ID, "Tibalt, the Fiend-Blooded", s82OldB007TibaltOracle, 2)
	effectstest.AdvanceToMain(t, g)
	hand, gy := me.Hand.Size(), me.Graveyard.Size()
	effectstest.S82Activate(t, g, me.ID, w, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d (drew one, discarded one)", me.Hand.Size(), hand)
	}
	if me.Graveyard.Size() != gy+1 {
		t.Errorf("graveyard %d, want %d", me.Graveyard.Size(), gy+1)
	}
	if got := effectstest.LoyaltyOf(g, w); got != 3 {
		t.Errorf("loyalty %d, want 3", got)
	}
}

func TestS82OldB007CTibaltMinusFourDamagesByHandSize(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	w := effectstest.S82Walker(g, me.ID, "Tibalt, the Fiend-Blooded", s82OldB007TibaltOracle, 5)
	effectstest.AdvanceToMain(t, g)
	n, life := opp.Hand.Size(), opp.Life
	effectstest.S82Activate(t, g, me.ID, w, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-n {
		t.Errorf("life %d, want %d (hand of %d)", opp.Life, life-n, n)
	}
	if got := effectstest.LoyaltyOf(g, w); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
}

func TestS82OldB007CTibaltMinusFourRefusesAtLowLoyalty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	w := effectstest.S82Walker(g, me.ID, "Tibalt, the Fiend-Blooded", s82OldB007TibaltOracle, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, w, 1, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Error("−4 activated at loyalty 2")
	}
}

func TestS82OldB007CTibaltMinusSixStealsEveryCreatureUntappedWithHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	w := effectstest.S82Walker(g, me.ID, "Tibalt, the Fiend-Blooded", s82OldB007TibaltOracle, 6)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == theirs {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, w, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, theirs)
	if c.Controller != me.ID || c.Tapped {
		t.Errorf("their creature: controller %v tapped %v, want mine and untapped", c.Controller, c.Tapped)
	}
	for _, id := range []uuid.UUID{mine, theirs} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
			t.Errorf("%v lacks haste", id)
		}
	}
	if got := effectstest.LoyaltyOf(g, w); got != 0 {
		t.Logf("loyalty %d", got)
	}
}
