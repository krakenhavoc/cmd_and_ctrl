package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b249_cards_test.go — Amonkhet cards, S82 OLD batch 249 (#5535).
// Every test plays the real catalog entry through the engine.

const (
	b249SupplyCaravan  = "1abcfe8d-4beb-4910-b353-1227e1256329"
	b249StirTheSands   = "52fa7e9b-36d9-409c-8981-6f8dfd8bb031"
	b249TrialStrength  = "b272cb56-dcba-479e-9fba-5b31e616a95f"
	b249Companion      = "3aa654cb-c0a0-48e7-901a-c690120b1510"
	b249Oketra         = "646e68c8-cc42-4d6c-8363-a18b10a10eae"
	b249Watchers       = "9ac46f46-3099-4ad9-ae58-bf25d7270e87"
	b249Vizier         = "008d5896-6fc9-4aaa-8c6f-a44c2feb98bb"
	b249TimeToReflect  = "7ce28a3c-5f93-4434-918e-1e74f9fc71d8"
	b249Cradle         = "36d06c91-5080-4f97-8e4c-ca8ac390e808"
	b249GraspingDunes  = "47d16c11-3033-44f3-9a12-2daf3453cc5b"
	b249AnglerDrake    = "620d7146-b5b3-487c-a24b-f36b10f48df4"
	b249WindsOfRebuke  = "6840a63d-d627-4a18-afa9-018406767642"
	b249CreatureBear   = "Creature — Bear"
	b249ZombieTypeLine = "Creature — Zombie"
)

func TestB249SupplyCaravanMakesAWarriorOnlyWithATappedCreature(t *testing.T) {
	for _, tapped := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		bear := effectstest.B12Creature(g, me.ID, "Bear", b249CreatureBear, 2, 2)
		if tapped {
			effectstest.TapCard(g, bear)
		}
		effectstest.CastHoldCreature(t, g, "Supply Caravan", "Creature — Camel", b249SupplyCaravan, 3, 5)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if tapped {
			want = 1
		}
		if got := effectstest.OnBattlefieldNamed(g, "Warrior"); got != want {
			t.Errorf("tapped=%v: %d Warrior tokens, want %d", tapped, got, want)
		}
	}
}

func TestB249StirTheSandsMakesThreeZombies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Stir the Sands", "Sorcery", b249StirTheSands, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.OnBattlefieldNamed(g, "Zombie"); got != 3 {
		t.Errorf("%d Zombie tokens, want 3", got)
	}
}

func TestB249StirTheSandsCycledMakesAZombieAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	handBefore := me.Hand.Size()
	effectstest.FillPool(me, 3)
	id, _ := effectstest.CycleFromHand(t, g, "Stir the Sands", "Sorcery", b249StirTheSands, "{B}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.OnBattlefieldNamed(g, "Zombie"); got != 1 {
		t.Errorf("%d Zombie tokens, want 1", got)
	}
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand = %d, want %d (the card drawn)", me.Hand.Size(), handBefore+1)
	}
	if !effectstest.GraveyardHas(me, id) {
		t.Error("Stir the Sands is not in the graveyard")
	}
}

func TestB249TrialOfStrengthMakesABeastAndReturnsForACartouche(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	trial := effectstest.B12PlayFromHand(t, g, "Trial of Strength", "Enchantment", b249TrialStrength, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.OnBattlefieldNamed(g, "Beast"); got != 1 {
		t.Fatalf("%d Beast tokens, want 1", got)
	}
	if p, tough := effectstest.PtOf(t, g, effectstest.BattlefieldIDsNamed(g, "Beast")[0]); p != 4 || tough != 2 {
		t.Errorf("the Beast is %d/%d, want 4/2", p, tough)
	}
	effectstest.B12PlayFromHand(t, g, "Test Cartouche", "Enchantment — Cartouche", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(trial) || !me.Hand.Contains(trial) {
		t.Error("Trial of Strength did not return to its owner's hand")
	}
}

func TestB249TrialOfStrengthStaysForAnythingElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	trial := effectstest.B12PlayFromHand(t, g, "Trial of Strength", "Enchantment", b249TrialStrength, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B12PlayFromHand(t, g, "Plain Aura", "Enchantment", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(trial) {
		t.Error("Trial of Strength left for a non-Cartouche")
	}
}

func TestB249CompanionOfTheTrialsUntapsOnlyWithAGideon(t *testing.T) {
	for _, gideon := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		companion := effectstest.B12Push(g, me.ID, "Companion of the Trials", "Creature — Bird Soldier", b249Companion, 2, 2)
		bear := effectstest.B12Creature(g, me.ID, "Bear", b249CreatureBear, 2, 2)
		effectstest.TapCard(g, bear)
		kind := "Jace"
		if gideon {
			kind = "Gideon"
		}
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: kind + " Test", TypeLine: "Legendary Planeswalker — " + kind,
			Counters: map[string]int{game.CounterLoyalty: 4}, Owner: me.ID, Controller: me.ID,
		})
		effectstest.AdvanceToMain(t, g)
		effectstest.FillPool(me, 1)
		effectstest.AddPool(t, g, me.ID, "{W}")
		err := g.ActivateCatalogAbility(me.ID, companion, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)})
		if !gideon {
			if err == nil {
				t.Error("activated without a Gideon planeswalker")
			}
			continue
		}
		if err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.HmlTapped(t, g, bear) {
			t.Error("the Bear is still tapped")
		}
	}
}

func TestB249OketraNeedsThreeOtherCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oketra := effectstest.B12Push(g, me.ID, "Oketra the True", "Legendary Creature — God", b249Oketra, 3, 6)
	effectstest.AssertKeywords(t, g, oketra, "double strike", "indestructible")
	for i := 0; i < 2; i++ {
		effectstest.B12Creature(g, me.ID, "Bear", b249CreatureBear, 2, 2)
	}
	effectstest.AdvanceToMain(t, g)
	if c, ok := g.LookupCardForEffect(oketra); !ok || !game.Restricted(&c, game.CantAttack) || !game.Restricted(&c, game.CantBlock) {
		t.Fatal("with two other creatures Oketra can neither attack nor block")
	}
	effectstest.CastHoldCreature(t, g, "Cub", "Creature — Cat", "", 1, 1)
	if c, ok := g.LookupCardForEffect(oketra); !ok || game.Restricted(&c, game.CantAttack) || game.Restricted(&c, game.CantBlock) {
		t.Fatal("with three other creatures Oketra is free")
	}
}

func TestB249OketraMakesAWarrior(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oketra := effectstest.B12Push(g, me.ID, "Oketra the True", "Legendary Creature — God", b249Oketra, 3, 6)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPool(me, 3)
	effectstest.AddPool(t, g, me.ID, "{W}")
	effectstest.ApaActivate(t, g, me, oketra, 0, game.ActivateAbilityParams{})
	if got := effectstest.OnBattlefieldNamed(g, "Warrior"); got != 1 {
		t.Errorf("%d Warrior tokens, want 1", got)
	}
}

func TestB249WatchersOfTheDeadLeavesEachOpponentTwoCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	small := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	watchers := effectstest.B12Push(g, me.ID, "Watchers of the Dead", "Artifact Creature — Cat", b249Watchers, 2, 2)
	var cards []uuid.UUID
	for _, n := range []string{"A", "B", "C", "D"} {
		cards = append(cards, effectstest.DftABInGraveyard(opp, "Card "+n, "Creature — Bear"))
	}
	smallID := effectstest.DftABInGraveyard(small, "Lone Card", "Creature — Bear")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, watchers, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(watchers) {
		t.Error("Watchers of the Dead was not exiled as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, opp.ID)
	if pick == nil {
		t.Fatal("the opponent was not asked which two cards to keep")
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{cards[0], cards[2]}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	if got := opp.Graveyard.Size(); got != 2 {
		t.Errorf("the opponent's graveyard holds %d cards, want 2", got)
	}
	if !effectstest.GraveyardHas(opp, cards[0]) || !effectstest.GraveyardHas(opp, cards[2]) {
		t.Error("the chosen cards are not the ones left in the graveyard")
	}
	if !effectstest.GraveyardHas(small, smallID) {
		t.Error("an opponent with fewer than three cards lost one")
	}
}

func TestB249VizierOfDefermentExilesAnAttackerOrBlockerAndItReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", b249CreatureBear, 2, 2)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", b249CreatureBear, 2, 2)
	effectstest.WoeS2Block(g, opp.ID, blocker, attacker)
	effectstest.CastHoldCreature(t, g, "Vizier of Deferment", "Creature — Human Cleric", b249Vizier, 2, 2)
	effectstest.AnswerPickTargetIfOpen(t, g, blocker)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blocker) {
		t.Fatal("the blocker was not exiled")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByName(g, "Blocker") {
		t.Error("the exiled creature did not return at the beginning of the next end step")
	}
}

func TestB249VizierOfDefermentLeavesAnIdleCreatureAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	idle := effectstest.B12Creature(g, opp.ID, "Idle", b249CreatureBear, 2, 2)
	effectstest.CastHoldCreature(t, g, "Vizier of Deferment", "Creature — Human Cleric", b249Vizier, 2, 2)
	effectstest.AnswerPickTargetIfOpen(t, g, idle)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(idle) {
		t.Error("a creature that neither attacked nor blocked was exiled")
	}
}

func TestB249TimeToReflectNeedsAZombieBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	zombie := effectstest.B12Creature(g, opp.ID, "Zombie", b249ZombieTypeLine, 2, 2)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", b249CreatureBear, 2, 2)
	bystander := effectstest.B12Creature(g, me.ID, "Bystander", b249CreatureBear, 2, 2)
	effectstest.WoeS2Block(g, opp.ID, zombie, attacker)
	if err := effectstest.CastCatalogSpellErr(t, g, "Time to Reflect", "Instant", b249TimeToReflect, effectstest.CardRefs(bystander)); err == nil {
		t.Fatal("Time to Reflect targeted a creature that never met a Zombie")
	}
	effectstest.CastCatalogSpell(t, g, "Time to Reflect", "Instant", b249TimeToReflect, effectstest.CardRefs(attacker))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the blocked creature was not exiled")
	}
	if !g.Battlefield.Contains(zombie) {
		t.Error("the Zombie was exiled too")
	}
}

func TestB249CradleOfTheAccursedMakesAZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.B12Push(g, me.ID, "Cradle of the Accursed", "Land — Desert", b249Cradle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPool(me, 3)
	effectstest.ApaActivate(t, g, me, land, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	if got := effectstest.OnBattlefieldNamed(g, "Zombie"); got != 1 {
		t.Errorf("%d Zombie tokens, want 1", got)
	}
}

func TestB249GraspingDunesPutsAMinusCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	land := effectstest.B12Push(g, me.ID, "Grasping Dunes", "Land — Desert", b249GraspingDunes, 0, 0)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", b249CreatureBear, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPool(me, 1)
	effectstest.ApaActivate(t, g, me, land, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)})
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	if got := effectstest.CounterCount(g, victim, game.CounterMinusOne); got != 1 {
		t.Errorf("the creature has %d -1/-1 counters, want 1", got)
	}
}

func TestB249AnglerDrakeBouncesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", b249CreatureBear, 2, 2)
	effectstest.CastHoldCreature(t, g, "Angler Drake", "Creature — Drake", b249AnglerDrake, 4, 4)
	effectstest.AnswerPickTargetIfOpen(t, g, theirs)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the creature did not return to its owner's hand")
	}
}

func TestB249WindsOfRebukeBouncesAndMillsEveryPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	before := make([]int, len(g.Seats))
	for i, s := range g.Seats {
		before[i] = s.Library.Size()
	}
	effectstest.CastCatalogSpell(t, g, "Winds of Rebuke", "Instant", b249WindsOfRebuke, effectstest.CardRefs(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the permanent did not return to its owner's hand")
	}
	for i, s := range g.Seats {
		if s.Library.Size() != before[i]-2 {
			t.Errorf("seat %d library %d, want %d", i, s.Library.Size(), before[i]-2)
		}
	}
}
