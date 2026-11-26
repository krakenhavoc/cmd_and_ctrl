package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 486 (#5581): Rise of the Eldrazi, equipment and planeswalkers.

const (
	roe486OgresCleaver = "e32c2ed2-be8b-497c-8b94-9c104f203b49"
	roe486SarkhanMad   = "e70a93e4-2fe7-46d3-8958-153ab2fcfd99"
	roe486Bear         = "Creature — Bear"
)

func TestRoeB486CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		roe486OgresCleaver: "Ogre's Cleaver", roe486SarkhanMad: "Sarkhan the Mad",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", name, spec)
		}
	}
}

// --- Ogre's Cleaver ----------------------------------------------------------

func TestRoeB486OgresCleaverGivesPlusFivePlusZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", roe486Bear, 2, 2)
	gear := effectstest.SeedEquipment(g, me.ID, "Ogre's Cleaver", roe486OgresCleaver)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, gear, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(bear)}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 7 || tough != 2 {
		t.Errorf("the equipped Bear is %d/%d, want 7/2", p, tough)
	}
}

func TestRoeB486OgresCleaverCannotEquipWithoutFiveMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", roe486Bear, 2, 2)
	gear := effectstest.SeedEquipment(g, me.ID, "Ogre's Cleaver", roe486OgresCleaver)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, gear, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(bear)}); err == nil {
		t.Fatal("equip {5} was paid with four mana")
	}
}

// --- Sarkhan the Mad ---------------------------------------------------------

func roe486Sarkhan(t *testing.T, loyalty int) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sarkhan := effectstest.PushWalkerForTest(g, me.ID, "Sarkhan the Mad", roe486SarkhanMad, loyalty)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	return g, me, opp, sarkhan
}

func TestRoeB486SarkhanZeroPutsTheCardInHandAndHurtsHimByItsManaValue(t *testing.T) {
	g, me, _, sarkhan := roe486Sarkhan(t, 7)
	top := effectstest.SeedLibraryTop(me, "Top", "Sorcery", "{2}{R}")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, sarkhan, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(top) || me.Hand.Size() != hand+1 {
		t.Error("the revealed card should be in hand")
	}
	if got := effectstest.LoyaltyOf(g, sarkhan); got != 4 {
		t.Errorf("loyalty %d, want 4 (7 - 3)", got)
	}
}

func TestRoeB486SarkhanZeroWithALandCostsNothing(t *testing.T) {
	g, me, _, sarkhan := roe486Sarkhan(t, 7)
	effectstest.SeedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	if err := g.ActivateCatalogAbility(me.ID, sarkhan, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, sarkhan); got != 7 {
		t.Errorf("loyalty %d, want 7", got)
	}
}

func TestRoeB486SarkhanZeroCanKillHimself(t *testing.T) {
	g, me, _, sarkhan := roe486Sarkhan(t, 2)
	effectstest.SeedLibraryTop(me, "Big", "Creature — Giant", "{4}{R}")
	if err := g.ActivateCatalogAbility(me.ID, sarkhan, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(sarkhan) {
		t.Error("5 damage to a loyalty-2 planeswalker should have killed it")
	}
}

func TestRoeB486SarkhanMinusTwoMakesTheVictimsControllerADragon(t *testing.T) {
	g, me, opp, sarkhan := roe486Sarkhan(t, 7)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", roe486Bear, 2, 2)
	indestructible := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Rock", TypeLine: "Creature — Golem", Power: 1, Toughness: 1, Keywords: []string{"indestructible"}})
	if err := g.ActivateCatalogAbility(me.ID, sarkhan, 1, game.ActivateAbilityParams{Targets: effectstest.CardTarget(victim)}); err != nil {
		t.Fatalf("-2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature was not sacrificed")
	}
	if n, _ := effectstest.CountTokensNamed(g, opp.ID, "Dragon"); n != 1 {
		t.Errorf("the creature's controller has %d Dragons, want 1", n)
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Dragon"); n != 0 {
		t.Errorf("the activator has %d Dragons, want 0", n)
	}
	if !g.Battlefield.Contains(indestructible) {
		t.Error("an untargeted creature was sacrificed")
	}
	if got := effectstest.LoyaltyOf(g, sarkhan); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
	for i := range g.Battlefield.Cards {
		c := g.Battlefield.Cards[i]
		if c.Name == "Dragon" && (c.CurrentPower() != 5 || c.CurrentToughness() != 5 || !game.HasKeyword(&g.Battlefield.Cards[i], "flying")) {
			t.Errorf("the Dragon is %d/%d or has no flying", c.CurrentPower(), c.CurrentToughness())
		}
	}
}

func TestRoeB486SarkhanMinusTwoSacrificesEvenAnIndestructibleCreature(t *testing.T) {
	g, me, opp, sarkhan := roe486Sarkhan(t, 7)
	rock := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Rock", TypeLine: "Creature — Golem", Power: 1, Toughness: 1, Keywords: []string{"indestructible"}})
	if err := g.ActivateCatalogAbility(me.ID, sarkhan, 1, game.ActivateAbilityParams{Targets: effectstest.CardTarget(rock)}); err != nil {
		t.Fatalf("-2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("sacrifice ignores indestructible")
	}
}

func TestRoeB486SarkhanMinusFourHitsWithEachDragonsPower(t *testing.T) {
	g, me, opp, sarkhan := roe486Sarkhan(t, 7)
	effectstest.B12Creature(g, me.ID, "Big Dragon", "Creature — Dragon", 5, 5)
	effectstest.B12Creature(g, me.ID, "Small Dragon", "Creature — Dragon", 3, 3)
	effectstest.B12Creature(g, me.ID, "Bear", roe486Bear, 4, 4)
	effectstest.B12Creature(g, opp.ID, "Their Dragon", "Creature — Dragon", 6, 6)
	life := effectstest.LifeOf(g, opp.ID)
	if err := g.ActivateCatalogAbility(me.ID, sarkhan, 2, game.ActivateAbilityParams{Targets: effectstest.RtPlayerRef(opp.ID)}); err != nil {
		t.Fatalf("-4: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-8 {
		t.Errorf("life %d, want %d (5 + 3 from your Dragons only)", got, life-8)
	}
	if got := effectstest.LoyaltyOf(g, sarkhan); got != 3 {
		t.Errorf("loyalty %d, want 3", got)
	}
}

func TestRoeB486SarkhanMinusFourCanTargetAPlaneswalker(t *testing.T) {
	g, me, opp, sarkhan := roe486Sarkhan(t, 7)
	effectstest.B12Creature(g, me.ID, "Dragon", "Creature — Dragon", 4, 4)
	walker := effectstest.PushWalkerForTest(g, opp.ID, "Walker", "", 6)
	if err := g.ActivateCatalogAbility(me.ID, sarkhan, 2, game.ActivateAbilityParams{Targets: effectstest.CardTarget(walker)}); err != nil {
		t.Fatalf("-4: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, walker); got != 2 {
		t.Errorf("their planeswalker has %d loyalty, want 2", got)
	}
}

func TestRoeB486SarkhanMinusFourRefusesACreatureTarget(t *testing.T) {
	g, me, opp, sarkhan := roe486Sarkhan(t, 7)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", roe486Bear, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, sarkhan, 2, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("a creature that is not a planeswalker was accepted")
	}
}
