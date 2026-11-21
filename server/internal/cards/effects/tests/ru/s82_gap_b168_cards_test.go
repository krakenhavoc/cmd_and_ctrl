package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB168HydrasGrowth   = "559903cd-5cba-47a5-8b5b-7067e393e76a"
	gapB168PresenceOfGond = "ab42398c-f0a1-4b94-ac5f-b8768e1b4e05"
	gapB168DaringPiracy   = "ae74efc6-4649-4783-a3a5-7a34d4667f64"
	gapB168ImpendingDoom  = "94366283-3b28-4d6a-96e7-c04e5aead5a6"
	gapB168OrdealPurph    = "652d282d-6dec-4043-a531-480f12ba471e"
	gapB168Reptilian      = "73f6619c-9d92-434e-9f06-52e0ff15cb5e"
	gapB168Aquatic        = "e4212b05-d397-49fb-af8b-9e52a60e6d1e"
	gapB168EternitySnare  = "b9bd1833-d4f9-46c0-bb94-331c456193fc"
	gapB168AuraType       = "Enchantment — Aura"
)

func TestGapB168EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB168HydrasGrowth: "Hydra's Growth", gapB168PresenceOfGond: "Presence of Gond",
		gapB168DaringPiracy: "Daring Piracy", gapB168ImpendingDoom: "Impending Doom",
		gapB168OrdealPurph: "Ordeal of Purphoros", gapB168Reptilian: "Reptilian Reflection",
		gapB168Aquatic: "Aquatic Incursion", gapB168EternitySnare: "Eternity Snare",
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

// gapB168Attach puts an Aura onto the battlefield attached to host.
func gapB168Attach(g *game.Game, owner uuid.UUID, name, oracle string, host uuid.UUID) uuid.UUID {
	aura := effectstest.PushCatalogPermanent(g, owner, name, gapB168AuraType, oracle, false)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, aura).AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: host}
	})
	return aura
}

func TestGapB168HydrasGrowthAddsACounterThenDoublesThemAtYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Ogre", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Hydra's Growth", gapB168AuraType, gapB168HydrasGrowth, effectstest.CardRefs(host))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, host, game.CounterPlusOne); got != 1 {
		t.Fatalf("counters after entering: %d, want 1", got)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, host, game.CounterPlusOne); got != 2 {
		t.Fatalf("counters after one upkeep: %d, want 2", got)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, host, game.CounterPlusOne); got != 4 {
		t.Errorf("counters after two upkeeps: %d, want 4", got)
	}
}

func TestGapB168PresenceOfGondLetsTheEnchantedCreatureMakeElfWarriors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Ogre", 3, 3)
	gapB168Attach(g, me.ID, "Presence of Gond", gapB168PresenceOfGond, host)
	effectstest.AdvanceToMain(t, g)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, host)
	if idx < 0 {
		t.Fatal("the enchanted creature has no granted ability")
	}
	effectstest.B16Activate(t, g, me.ID, host, idx, game.ActivateAbilityParams{Ref: ref})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, host) {
		t.Error("the ability did not tap the creature")
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Elf Warrior" && c.IsToken() {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d Elf Warrior tokens, want 1", n)
	}
}

func TestGapB168ImpendingDoomPumpsForcesAttackAndPunishesItsDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.B12Creature(g, opp.ID, "Host", "Creature — Ogre", 2, 2)
	gapB168Attach(g, me.ID, "Impending Doom", gapB168ImpendingDoom, host)
	if p, tough := effectstest.PtOf(t, g, host); p != 5 || tough != 5 {
		t.Errorf("the host is %d/%d, want 5/5", p, tough)
	}
	if len(effectstest.GaLayered(t, g, host).Effective().AttackRequirements) == 0 {
		t.Error("the host does not have to attack each combat")
	}
	life := opp.Life
	effectstest.B27Kill(g, host)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("the creature's controller is at %d, want %d", opp.Life, life-3)
	}
}

func TestGapB168OrdealOfPurphorosGrowsThenSacrificesItselfForThreeDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Ogre", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(host, game.CounterPlusOne, 2) })
	aura := gapB168Attach(g, me.ID, "Ordeal of Purphoros", gapB168OrdealPurph, host)
	life := opp.Life
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(host, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, host, game.CounterPlusOne); got != 3 {
		t.Errorf("counters %d, want 3", got)
	}
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura was not sacrificed at three counters")
	}
	if opp.Life != life-3 {
		t.Errorf("the target is at %d, want %d", opp.Life, life-3)
	}
}

func TestGapB168OrdealOfPurphorosStaysBelowThreeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Ogre", 2, 2)
	aura := gapB168Attach(g, me.ID, "Ordeal of Purphoros", gapB168OrdealPurph, host)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(host, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, host, game.CounterPlusOne); got != 1 {
		t.Errorf("counters %d, want 1", got)
	}
	if !g.Battlefield.Contains(aura) {
		t.Error("the Aura was sacrificed below three counters")
	}
}

func TestGapB168ReptilianReflectionBecomesAFiveFourWhenYouCycleAndMayDecline(t *testing.T) {
	const auraExtraction = "84100106-1f97-4859-aadd-a35796ee000f"
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		refl := effectstest.PushCatalogPermanent(g, me.ID, "Reptilian Reflection", "Enchantment", gapB168Reptilian, false)
		effectstest.SeedLibrary(me, "A", "B")
		effectstest.CycleFromHand(t, g, "Aura Extraction", "Instant", auraExtraction, "{C}{C}")
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		layered := effectstest.GaLayered(t, g, refl)
		if got := layered.IsCreature(); got != yes {
			t.Fatalf("yes=%v: creature=%v", yes, got)
		}
		if !layered.IsEnchantment() {
			t.Errorf("yes=%v: it stopped being an enchantment", yes)
		}
		if yes {
			if p, tough := effectstest.PtOf(t, g, refl); p != 5 || tough != 4 {
				t.Errorf("it is %d/%d, want 5/4", p, tough)
			}
			for _, kw := range []string{"trample", "haste"} {
				if !effectstest.HasEffectiveKeyword(t, g, refl, kw) {
					t.Errorf("no %s", kw)
				}
			}
			if !layered.HasSubtype("Dinosaur") {
				t.Error("it is not a Dinosaur")
			}
			effectstest.AdvanceToUpkeepOfSeat(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
			if effectstest.GaLayered(t, g, refl).IsCreature() {
				t.Error("it is still a creature after the turn")
			}
		}
	}
}

func TestGapB168DaringPiracyMakesAHastyMenacingPirateThatLeavesAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Daring Piracy", "Enchantment", gapB168DaringPiracy, false)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	var pirate uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Pirate" && c.IsToken() {
			pirate = c.InstanceID
		}
	}
	if pirate == uuid.Nil {
		t.Fatal("no Pirate token at the beginning of combat")
	}
	if p, tough := effectstest.PtOf(t, g, pirate); p != 1 || tough != 1 {
		t.Errorf("the Pirate is %d/%d, want 1/1", p, tough)
	}
	for _, kw := range []string{"menace", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, pirate, kw) {
			t.Errorf("the Pirate has no %s", kw)
		}
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pirate) {
		t.Error("the Pirate was not exiled at the end step")
	}
}

func TestGapB168DaringPiracyOnlyTriggersOnYourOwnCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Daring Piracy", "Enchantment", gapB168DaringPiracy, false)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.BattlefieldNamed(g, me.ID, "Pirate"); n != 0 {
		t.Errorf("%d Pirates on an opponent's combat, want 0", n)
	}
}

func TestGapB168EternitySnareDrawsAndKeepsTheCreatureTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.B12Creature(g, opp.ID, "Host", "Creature — Ogre", 2, 2)
	effectstest.TapForTest(t, g, host)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Eternity Snare", gapB168AuraType, gapB168EternitySnare, effectstest.CardRefs(host))
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d (cast and drew)", got, hand+1)
	}
	effectstest.AdvanceToUpkeepOf(t, g, g.Turn.ActiveSeat+1)
	if !effectstest.TappedNow(g, host) {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
}

func TestGapB168AquaticIncursionMakesTwoHexproofMerfolkAndOneCanBeUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	inc := effectstest.CastCatalogSpell(t, g, "Aquatic Incursion", "Enchantment", gapB168Aquatic, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var merfolk []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Merfolk" && c.IsToken() {
			merfolk = append(merfolk, c.InstanceID)
			if !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "hexproof") {
				t.Error("a Merfolk has no hexproof")
			}
		}
	}
	if len(merfolk) != 2 {
		t.Fatalf("%d Merfolk, want 2", len(merfolk))
	}
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	if err := effectstest.BroB01Activate(t, g, me, inc, 0, "{C}{C}{C}{U}"); err == nil {
		// no Targets given: must be refused
		t.Error("activated with no target")
	}
	if err := g.ActivateCatalogAbility(me.ID, inc, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(other)}); err == nil {
		t.Error("targeted a non-Merfolk")
	}
	if err := g.ActivateCatalogAbility(me.ID, inc, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(merfolk[0])}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, merfolk[0]).SummonedThisTurn = false })
	if err := effectstest.PorBlock(t, g, merfolk[0], opp.ID, bear); err == nil {
		t.Error("the targeted Merfolk was blocked")
	}
}

// The control for the test above: an untargeted Merfolk can be blocked.
func TestGapB168AnUntargetedMerfolkCanBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Aquatic Incursion", "Enchantment", gapB168Aquatic, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var merfolk uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Merfolk" && c.IsToken() {
			merfolk = c.InstanceID
		}
	}
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, merfolk).SummonedThisTurn = false })
	if err := effectstest.PorBlock(t, g, merfolk, opp.ID, bear); err != nil {
		t.Errorf("an untargeted Merfolk could not be blocked: %v", err)
	}
}
