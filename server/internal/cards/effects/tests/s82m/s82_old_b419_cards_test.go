package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b419_cards_test.go — Return to Ravnica, batch 419 (S82 older sets,
// #5566): eight Auras. Every test plays the real catalog entry through the
// engine.

const (
	oldB419Glee      = "13aeccd5-cca2-4249-904d-d5516f977440"
	oldB419Burgeon   = "2772e3cc-10e0-420f-b183-71ebaf3580e1"
	oldB419Flooding  = "474db5db-914e-474d-b0c2-7987d5986dc3"
	oldB419Tithe     = "60ff5af4-eb95-4544-9f63-d42c22112486"
	oldB419Blockade  = "b4b397a9-85cb-4bdf-8c2e-a151a02a4208"
	oldB419Fury      = "b7093835-0f84-4ea4-b228-2e39bd2642ab"
	oldB419Grasp     = "c1459d75-2bc2-4dae-8c7e-4e759a4dffc9"
	oldB419Pursuit   = "cdce8b86-81e3-443c-af91-f6b4f62f5b59"
	oldB419BearType  = "Creature — Bear"
	oldB419ForestTyp = "Basic Land — Forest"
)

func TestOldB419EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB419Glee: "Deviant Glee", oldB419Burgeon: "Urban Burgeoning", oldB419Flooding: "Chronic Flooding",
		oldB419Tithe: "Soul Tithe", oldB419Blockade: "Security Blockade", oldB419Fury: "Racecourse Fury",
		oldB419Grasp: "Paralyzing Grasp", oldB419Pursuit: "Pursuit of Flight",
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

func oldB419Activate(t *testing.T, g *game.Game, me *game.Player, host uuid.UUID, mana string, targets []game.TargetRef) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	if mana != "" {
		effectstest.FloatMana(t, g, me, mana)
	}
	idx, ref := effectstest.GrantedActivatedIndex(t, g, host)
	if idx < 0 {
		t.Fatal("the host has no granted ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, host, idx, game.ActivateAbilityParams{Strict: true, Ref: ref, Targets: targets}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB419DeviantGleePumpsAndGrantsTrampleForR(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB419BearType, 2, 2)
	effectstest.Enchant(t, g, "Deviant Glee", oldB419Glee, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 3 {
		t.Fatalf("the Bear is %d/%d, want 4/3", p, tough)
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Fatal("the Bear has trample before the ability")
	}
	oldB419Activate(t, g, me, bear, "{R}", nil)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the Bear did not gain trample")
	}
}

func TestOldB419PursuitOfFlightPumpsAndGrantsFlyingForU(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB419BearType, 2, 2)
	effectstest.Enchant(t, g, "Pursuit of Flight", oldB419Pursuit, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Fatalf("the Bear is %d/%d, want 4/4", p, tough)
	}
	oldB419Activate(t, g, me, bear, "{U}", nil)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the Bear did not gain flying")
	}
}

func TestOldB419ParalyzingGraspKeepsATappedCreatureTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	victim := effectstest.B12Creature(g, opp.ID, "Bear", oldB419BearType, 2, 2)
	free := effectstest.B12Creature(g, opp.ID, "Free Bear", oldB419BearType, 2, 2)
	_ = me
	effectstest.Enchant(t, g, "Paralyzing Grasp", oldB419Grasp, victim)
	if effectstest.B16Tapped(t, g, victim) {
		t.Fatal("Paralyzing Grasp tapped the creature as it entered")
	}
	effectstest.TapForTest(t, g, victim, free)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	for g.Turn.ActiveSeat != 1 {
		effectstest.AdvanceToNextSeatsTurn(t, g)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.B16Tapped(t, g, free) {
		t.Error("an unenchanted creature stayed tapped through its untap step")
	}
	if !effectstest.B16Tapped(t, g, victim) {
		t.Error("the enchanted creature untapped")
	}
}

func TestOldB419UrbanBurgeoningUntapsTheLandInOtherPlayersUntapSteps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	land := effectstest.B12Permanent(g, me.ID, "Forest", oldB419ForestTyp)
	effectstest.Enchant(t, g, "Urban Burgeoning", oldB419Burgeon, land)
	effectstest.TapForTest(t, g, land)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	for g.Turn.ActiveSeat != 1 {
		effectstest.AdvanceToNextSeatsTurn(t, g)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.B16Tapped(t, g, land) {
		t.Error("the enchanted land stayed tapped through another player's untap step")
	}

	// Control: with no Aura the land stays tapped through the same step.
	g = effectstest.NewCatalogGame(t)
	me, _ = effectstest.SeatsForCombat(g)
	bare := effectstest.B12Permanent(g, me.ID, "Forest", oldB419ForestTyp)
	effectstest.TapForTest(t, g, bare)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	for g.Turn.ActiveSeat != 1 {
		effectstest.AdvanceToNextSeatsTurn(t, g)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if !effectstest.B16Tapped(t, g, bare) {
		t.Error("a bare land untapped in another player's untap step")
	}
}

func TestOldB419ChronicFloodingMillsTheLandsControllerWhenItTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	land := effectstest.B12Permanent(g, opp.ID, "Forest", oldB419ForestTyp)
	for i := 0; i < 6; i++ {
		effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Lib", TypeLine: "Instant"})
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Lib", TypeLine: "Instant"})
	}
	effectstest.Enchant(t, g, "Chronic Flooding", oldB419Flooding, land)
	lib, mine := opp.Library.Size(), me.Library.Size()
	effectstest.TapForTest(t, g, land)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != lib-3 {
		t.Errorf("the land's controller's library is %d, want %d", got, lib-3)
	}
	if got := me.Library.Size(); got != mine {
		t.Errorf("the Aura's controller's library is %d, want %d", got, mine)
	}
}

func TestOldB419SecurityBlockadeMakesAKnightAndShieldsTheLandsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.B12Permanent(g, me.ID, "Forest", oldB419ForestTyp)
	effectstest.Enchant(t, g, "Security Blockade", oldB419Blockade, land)
	effectstest.PassPriorityAroundTable(t, g)
	knight := effectstest.FindBattlefieldByName(g, "Knight")
	if knight == uuid.Nil {
		t.Fatal("no Knight token")
	}
	if p, tough := effectstest.PtOf(t, g, knight); p != 2 || tough != 2 || !effectstest.HasEffectiveKeyword(t, g, knight, "vigilance") {
		t.Errorf("the Knight is %d/%d", p, tough)
	}
	oldB419Activate(t, g, me, land, "", nil)
	life := me.Life
	effectstest.Pr7Hit(t, g, effectstest.B12Creature(g, g.Seats[1].ID, "Src", oldB419BearType, 1, 1), me.ID, 3)
	if me.Life != life-2 {
		t.Errorf("life %d, want %d (1 of the 3 damage prevented)", me.Life, life-2)
	}
}

func TestOldB419RacecourseFuryGrantsHasteWithTheLandsTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.B12Permanent(g, me.ID, "Forest", oldB419ForestTyp)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB419BearType, 2, 2)
	effectstest.Enchant(t, g, "Racecourse Fury", oldB419Fury, land)
	oldB419Activate(t, g, me, land, "", effectstest.CardTarget(bear))
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the Bear did not gain haste")
	}
	if !effectstest.B16Tapped(t, g, land) {
		t.Error("the land did not tap")
	}
}

func TestOldB419SoulTitheSacrificesUnlessTheControllerPaysItsManaValue(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		_ = me
		host := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Costly Bear", TypeLine: oldB419BearType, ManaCost: "{2}{G}",
			Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
		})
		effectstest.Enchant(t, g, "Soul Tithe", oldB419Tithe, host)
		for g.Turn.ActiveSeat != 1 {
			effectstest.AdvanceToNextSeatsTurn(t, g)
		}
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		effectstest.PassPriorityUntilChoice(t, g)
		prompt := effectstest.PayUnlessFor(g, opp.ID)
		if prompt == nil {
			t.Fatalf("pay=%v: the enchanted permanent's controller was not offered the tax", pay)
		}
		if pay {
			effectstest.FloatMana(t, g, opp, "{C}{C}{C}")
		}
		effectstest.AnswerPayUnless(t, g, opp.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if got := g.Battlefield.Contains(host); got != pay {
			t.Errorf("pay=%v: the permanent is on the battlefield %v", pay, got)
		}
	}
}
