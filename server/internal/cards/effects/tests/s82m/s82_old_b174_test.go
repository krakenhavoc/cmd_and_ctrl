package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB174SarkhanOracle = "1d37b453-f10c-42f2-8390-626a8a8eb27c"
	oldB174RalOracle     = "f29f4c73-5e5c-4341-8582-2873fbe5c5a3"
)

func TestOldB174AreFullWithNoCaveats(t *testing.T) {
	for _, id := range []string{oldB174SarkhanOracle, oldB174RalOracle} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Fatalf("%s is not registered", id)
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %v with caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB174SarkhanUnbrokenPlusOneDrawsThenAddsAnyColorMana(t *testing.T) {
	g, seat, me, _ := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Sarkhan Unbroken", oldB174SarkhanOracle, 4)
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	})
	effectstest.AdvanceToMainOf(t, g, seat)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceMana, me.ID)
	if pick == nil {
		t.Fatal("no colour pick for the any-colour mana")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := me.ManaPool.CountColor("R"); got != 1 {
		t.Errorf("red mana %d, want 1", got)
	}
}

func TestOldB174SarkhanUnbrokenMinusTwoMakesADragon(t *testing.T) {
	g, seat, me, _ := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Sarkhan Unbroken", oldB174SarkhanOracle, 4)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.IsToken() && c.HasSubtype("Dragon") && c.Power == 4 && c.Toughness == 4 {
				found = true
			}
		}
	})
	if !found {
		t.Error("no 4/4 Dragon token")
	}
}

func TestOldB174SarkhanUnbrokenMinusEightFetchesEveryDragonOnly(t *testing.T) {
	g, seat, me, _ := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Sarkhan Unbroken", oldB174SarkhanOracle, 8)
	var a, b uuid.UUID
	g.WithWriteLock(func() {
		me.Library.Cards = nil
		a = effectstest.PushLibraryCardForTest(me, game.Card{Name: "Dragon A", TypeLine: "Creature — Dragon", Power: 2, Toughness: 2})
		b = effectstest.PushLibraryCardForTest(me, game.Card{Name: "Dragon B", TypeLine: "Creature — Dragon", Power: 2, Toughness: 2})
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Dragon Tribal", TypeLine: "Kindred Sorcery — Dragon"})
	})
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−8: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	for _, name := range []string{"Bear", "Dragon Tribal"} {
		if effectstest.SearchOptionNamed(g, c, name) != uuid.Nil {
			t.Errorf("search offered %q, want only Dragon creature cards", name)
		}
	}
	effectstest.AnswerSearchByID(t, g, me.ID, a, b)
	for _, name := range []string{"Dragon A", "Dragon B"} {
		if _, ok := effectstest.SearchFetchedCard(g, name); !ok {
			t.Errorf("%s did not reach the battlefield", name)
		}
	}
}

func TestOldB174RalMinusThreeCountsInstantsAndSorceriesInGraveyardAndExile(t *testing.T) {
	g, seat, me, opp := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Ral, Izzet Viceroy", oldB174RalOracle, 5)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 1, 9)
	effectstest.PushGraveyardCardTyped(me, "Bolt", "Instant")
	effectstest.PushGraveyardCardTyped(me, "Wrath", "Sorcery")
	effectstest.PushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	effectstest.PushGraveyardCardTyped(opp, "Their Bolt", "Instant")
	g.WithWriteLock(func() {
		g.Exile.Cards = append(g.Exile.Cards,
			game.Card{InstanceID: uuid.New(), Name: "Exiled Instant", TypeLine: "Instant", Owner: me.ID, Controller: me.ID},
			game.Card{InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID},
			game.Card{InstanceID: uuid.New(), Name: "Face Down", TypeLine: "Instant", Owner: me.ID, Controller: me.ID, FaceDown: true})
	})
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim), Strict: true}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var damage int
	g.WithWriteLock(func() { damage = effectstest.FindBattlefieldCardForTest(g, victim).DamageMarked })
	if damage != 3 {
		t.Errorf("damage %d, want 3 (Bolt, Wrath, one exiled instant)", damage)
	}
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(pw), Strict: true}); err == nil {
		t.Error("−3 accepted a planeswalker as its target")
	}
}

func TestOldB174RalEmblemDealsFourAndDrawsTwoOnInstantsAndSorceriesOnly(t *testing.T) {
	g, seat, me, opp := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Ral, Izzet Viceroy", oldB174RalOracle, 8)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 1, 9)
	g.WithWriteLock(func() {
		for i := 0; i < 4; i++ {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		}
	})
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−8: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Emblems.Size() != 1 {
		t.Fatalf("emblems %d, want 1", me.Emblems.Size())
	}

	effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	g.RunStateChecksForTest()
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a creature spell triggered the emblem")
	}
	effectstest.PassPriorityAroundTable(t, g)

	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Some Sorcery", "Sorcery", "", nil)
	g.RunStateChecksForTest()
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	var damage int
	g.WithWriteLock(func() { damage = effectstest.FindBattlefieldCardForTest(g, victim).DamageMarked })
	if damage != 4 {
		t.Errorf("damage %d, want 4", damage)
	}
	// The sorcery left the hand when cast (-1 from the pushed card), two were drawn.
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+2)
	}
}

func TestOldB174RalPlusOneTakesOneOfTheTopTwoAndMillsTheOther(t *testing.T) {
	g, seat, me, _ := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Ral, Izzet Viceroy", oldB174RalOracle, 5)
	var ids []uuid.UUID
	g.WithWriteLock(func() { ids = effectstest.SeedLibrary(me, "A", "B", "C", "D") })
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	other := me.Library.Cards[len(me.Library.Cards)-2].InstanceID
	effectstest.AnswerChooseCards(t, g, me.ID, top)
	g.SettleResolution()
	if !me.Hand.Contains(top) || !me.Graveyard.Contains(other) {
		t.Errorf("the chosen card is not in hand or the other not in the graveyard (ids %v)", ids)
	}
}
