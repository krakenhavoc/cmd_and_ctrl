package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b466_cards_test.go — Mirrodin Besieged, S82 OLD batch 466 (#5576).
// Every test plays the real catalog entry through the engine.

const (
	oldB466Plaguemaw = "22252186-15d0-468f-8cab-c74db9b71439"
	oldB466Fangren   = "e154e612-4041-4fb7-903f-917588dbe58e"
	oldB466Decimator = "eeabf42a-b1ac-4579-b82a-aabf9052ec20"
	oldB466Silverskn = "4eb24ae3-da1a-4565-934f-e25f56036d71"
	oldB466Mortarpod = "77248a76-373c-48fc-939f-aa378ab57f7c"
	oldB466Sledge    = "ba5d8c0d-3e4e-4652-86d2-80ce2f48dc0f"
	oldB466RotWolf   = "7409906d-7111-48fa-bfea-8b95e0fdb7dc"
	oldB466Rats      = "c7335845-ace9-402d-a5be-3bfe3fe9b089"
	oldB466Vatmother = "e9162478-0b10-422a-a11c-7b39b00f026e"
	oldB466Mage      = "0848bb49-5758-41d6-a3e8-0c62e1768f52"
	oldB466Turbine   = "b9043dcd-234f-4144-80db-1cb903c009ea"
)

func TestOldB466EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB466Plaguemaw: "Plaguemaw Beast",
		oldB466Fangren:   "Fangren Marauder",
		oldB466Decimator: "Decimator Web",
		oldB466Silverskn: "Silverskin Armor",
		oldB466Mortarpod: "Mortarpod",
		oldB466Sledge:    "Piston Sledge",
		oldB466RotWolf:   "Rot Wolf",
		oldB466Rats:      "Septic Rats",
		oldB466Vatmother: "Phyrexian Vatmother",
		oldB466Mage:      "Treasure Mage",
		oldB466Turbine:   "Myr Turbine",
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

func TestOldB466PlaguemawBeastSacrificesACreatureToProliferate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	beast := effectstest.B12Push(g, me.ID, "Plaguemaw Beast", "Creature — Phyrexian Beast", oldB466Plaguemaw, 4, 3)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	target := effectstest.B12Creature(g, me.ID, "Counted", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, target)
		c.Counters = map[string]int{"+1/+1": 1}
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, beast, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature is sacrificed as a cost")
	}
	if !effectstest.B16Tapped(t, g, beast) {
		t.Error("the ability has a {T} cost")
	}
	effectstest.SettleAnsweringProliferate(t, g)
	if got := effectstest.BattlefieldCardCopy(t, g, target).Counters["+1/+1"]; got != 2 {
		t.Errorf("+1/+1 counters %d, want 2 after proliferate", got)
	}
}

func TestOldB466PlaguemawBeastNeedsTheTapSymbol(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	beast := effectstest.PushCatalogPermanent(g, me.ID, "Plaguemaw Beast", "Creature — Phyrexian Beast", oldB466Plaguemaw, true)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, beast, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err == nil {
		t.Fatal("a summoning-sick Beast paid its {T} cost")
	}
}

func TestOldB466FangrenMarauderGainsFiveForEachArtifactToAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Fangren Marauder", "Creature — Beast", oldB466Fangren, 5, 5)
	mine := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	theirs := effectstest.B12Permanent(g, opp.ID, "Stone", "Artifact")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	effectstest.P1Destroy(t, g, mine)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Fatalf("life %d → %d, want +5 for one's own artifact", life, me.Life)
	}
	effectstest.P1Destroy(t, g, theirs)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+10 {
		t.Fatalf("life %d → %d, want +10 after an opponent's artifact too", life, me.Life)
	}
	effectstest.P1Destroy(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+10 {
		t.Errorf("a nonartifact creature dying gained life: %d", me.Life)
	}
}

func TestOldB466FangrenMarauderMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Fangren Marauder", "Creature — Beast", oldB466Fangren, 5, 5)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	life := me.Life
	effectstest.P1Destroy(t, g, rock)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("declined, yet life %d → %d", life, me.Life)
	}
}

func TestOldB466DecimatorWebDrainsPoisonsAndMills(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	web := effectstest.PushCatalogPermanent(g, me.ID, "Decimator Web", "Artifact", oldB466Decimator, false)
	effectstest.AdvanceToMain(t, g)
	life, lib, gy := opp.Life, opp.Library.Size(), opp.Graveyard.Size()
	effectstest.RavB2MustActivate(t, g, me.ID, web, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}, "C", "C", "C", "C")
	if opp.Life != life-2 {
		t.Errorf("life %d → %d, want -2", life, opp.Life)
	}
	if got := effectstest.PoisonOn(opp); got != 1 {
		t.Errorf("poison counters %d, want 1", got)
	}
	if opp.Library.Size() != lib-6 || opp.Graveyard.Size() != gy+6 {
		t.Errorf("library %d → %d, graveyard %d → %d, want six milled", lib, opp.Library.Size(), gy, opp.Graveyard.Size())
	}
	if !effectstest.B16Tapped(t, g, web) {
		t.Error("the ability has a {T} cost")
	}
}

func TestOldB466DecimatorWebTargetsOnlyAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	web := effectstest.PushCatalogPermanent(g, me.ID, "Decimator Web", "Artifact", oldB466Decimator, false)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.RavB2Activate(t, g, me.ID, web, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}, "C", "C", "C", "C"); err == nil {
		t.Fatal("Decimator Web targeted its own controller")
	}
}

func TestOldB466SilverskinArmorPumpsAndMakesTheCreatureAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	armor := effectstest.PushCatalogPermanent(g, me.ID, "Silverskin Armor", "Artifact — Equipment", oldB466Silverskn, false)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, armor).AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: bear}
	})
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("equipped Bear is %d/%d, want 3/3", p, tough)
	}
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, _ := g.LookupCardForEffect(bear)
		if !c.IsArtifact() || !c.IsCreature() {
			t.Errorf("equipped Bear is artifact %v creature %v, want both", c.IsArtifact(), c.IsCreature())
		}
	})
	oldB466Unattach(g, armor)
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, _ := g.LookupCardForEffect(bear)
		if c.IsArtifact() {
			t.Error("the Bear stayed an artifact after the Armor left it")
		}
	})
}

func oldB466Unattach(g *game.Game, equipment uuid.UUID) {
	g.WithWriteLock(func() { _ = g.UnattachForEffect(equipment) })
}

func TestOldB466MortarpodMakesAGermAndGrantsASacrificePing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Mortarpod", "Artifact — Equipment", oldB466Mortarpod, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var germ game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Phyrexian Germ" {
			germ = c
		}
	}
	if germ.InstanceID == uuid.Nil {
		t.Fatal("living weapon made no Germ")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, germ.InstanceID); p != 0 || tough != 1 {
		t.Errorf("the equipped Germ is %d/%d, want 0/1", p, tough)
	}
	idx, ref := effectstest.GrantedActivatedIndex(t, g, germ.InstanceID)
	if idx < 0 {
		t.Fatal("the equipped Germ has no granted ability")
	}
	life := opp.Life
	effectstest.B16Activate(t, g, me.ID, germ.InstanceID, idx, game.ActivateAbilityParams{
		Ref: ref, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(germ.InstanceID) {
		t.Error("the creature is sacrificed as the cost")
	}
	if opp.Life != life-1 {
		t.Errorf("life %d → %d, want 1 damage", life, opp.Life)
	}
}

func TestOldB466PistonSledgeAttachesOnEntryAndEquipsBySacrificingAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	sledge := effectstest.CastCatalogSpell(t, g, "Piston Sledge", "Artifact — Equipment", oldB466Sledge, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 5 || tough != 3 {
		t.Fatalf("the Bear is %d/%d after the entry attach, want 5/3", p, tough)
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, sledge, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{rock}, Targets: effectstest.CardRefs(other),
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is sacrificed as the equip cost")
	}
	if p, _ := effectstest.EffectivePTOf(t, g, other); p != 5 {
		t.Errorf("the moved Sledge left the other creature at power %d, want 5", p)
	}
}

func TestOldB466RotWolfDrawsWhenACreatureItDamagedDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wolf := effectstest.B12Push(g, me.ID, "Rot Wolf", "Creature — Phyrexian Wolf", oldB466RotWolf, 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 3)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 3, 3)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(wolf, victim, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	hand := me.Hand.Size()
	effectstest.P1Destroy(t, g, bystander)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatal("a creature the Wolf never damaged drew a card")
	}
	effectstest.P1Destroy(t, g, victim)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
}

func TestOldB466SepticRatsGetsBiggerOnlyAgainstAPoisonedDefender(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rats := effectstest.B12Push(g, me.ID, "Septic Rats", "Creature — Phyrexian Rat", oldB466Rats, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, rats)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, rats); p != 2 || tough != 2 {
		t.Fatalf("against an unpoisoned defender the Rats are %d/%d, want 2/2", p, tough)
	}
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	rats2 := effectstest.B12Push(g2, me2.ID, "Septic Rats", "Creature — Phyrexian Rat", oldB466Rats, 2, 2)
	g2.WithWriteLock(func() { _ = g2.AddPlayerCounterByForEffect(me2.ID, opp2.ID, game.CounterPoison, 1) })
	effectstest.DeclareAttack(t, g2, opp2.ID, rats2)
	effectstest.PassPriorityAroundTable(t, g2)
	if p, tough := effectstest.EffectivePTOf(t, g2, rats2); p != 3 || tough != 3 {
		t.Errorf("against a poisoned defender the Rats are %d/%d, want 3/3", p, tough)
	}
}

func TestOldB466PhyrexianVatmotherGivesItsControllerPoisonEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Phyrexian Vatmother", "Creature — Phyrexian Horror", oldB466Vatmother, 4, 5)
	opp := g.Seats[1]
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PoisonOn(me) != 0 || effectstest.PoisonOn(opp) != 0 {
		t.Fatal("the Vatmother fired in an opponent's upkeep")
	}
	effectstest.AdvanceToStepOf(t, g, 0, game.StepUpkeep)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PoisonOn(me); got != 1 {
		t.Errorf("poison counters %d after the controller's upkeep, want 1", got)
	}
	if effectstest.PoisonOn(opp) != 0 {
		t.Error("an opponent got poison")
	}
}

func TestOldB466TreasureMageFindsAnArtifactOfManaValueSixOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	big := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Big Artifact", TypeLine: "Artifact", ManaCost: "{6}"})
	small := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Small Artifact", TypeLine: "Artifact", ManaCost: "{5}"})
	creature := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Big Creature", TypeLine: "Creature — Golem", ManaCost: "{7}"})
	effectstest.CastCatalogSpell(t, g, "Treasure Mage", "Creature — Human Wizard", oldB466Mage, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt was offered")
	}
	if len(c.SearchCards) != 1 || c.SearchCards[0] != big {
		t.Fatalf("offered %v, want only the mana value 6 artifact (not %v or %v)", c.SearchCards, small, creature)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, big)
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, h := range me.Hand.Cards {
		if h.InstanceID == big {
			found = true
		}
	}
	if !found {
		t.Error("the artifact is not in the hand")
	}
}

func TestOldB466MyrTurbineMakesAMyrAndTutorsForFiveTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	turbine := effectstest.PushCatalogPermanent(g, me.ID, "Myr Turbine", "Artifact", oldB466Turbine, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, turbine, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.BattlefieldNamed(g, me.ID, "Myr"); n != 1 {
		t.Fatalf("%d Myr tokens, want 1", n)
	}
	if !effectstest.B16Tapped(t, g, turbine) {
		t.Error("the token ability has a {T} cost")
	}
}

func TestOldB466MyrTurbineSecondAbilityTapsFiveMyr(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	turbine := effectstest.PushCatalogPermanent(g, me.ID, "Myr Turbine", "Artifact", oldB466Turbine, false)
	var myr []uuid.UUID
	for i := 0; i < 5; i++ {
		myr = append(myr, effectstest.B12Creature(g, me.ID, "Myr Token", "Artifact Creature — Myr", 1, 1))
	}
	target := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Myr Battlesphere", TypeLine: "Artifact Creature — Myr", ManaCost: "{7}", Power: 4, Toughness: 7})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, turbine, 1, game.ActivateAbilityParams{TapIDs: myr[:4]}); err == nil {
		t.Fatal("the ability was activated tapping only four Myr")
	}
	if err := g.ActivateCatalogAbility(me.ID, turbine, 1, game.ActivateAbilityParams{TapIDs: myr}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, target)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(target) {
		t.Error("the Myr creature card is not on the battlefield")
	}
	for _, id := range myr {
		if !effectstest.B16Tapped(t, g, id) {
			t.Error("a Myr was not tapped to pay")
		}
	}
}
