package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b13_cards_test.go — March of the Machine, S82 batch 13 (#5477):
// twelve dig, loot and graveyard-utility creatures and Auras. Each test plays
// the real catalog entry through the engine.

const (
	s82b13LookoutOracle    = "f8ad2f88-72ae-4e47-beb5-25a6dacc709b"
	s82b13WayfinderOracle  = "3ec767f2-40ba-46c9-b7cd-5821986dda0c"
	s82b13MolochOracle     = "f2a6dc8d-3c98-44ae-aff0-b838d7aee0b7"
	s82b13RottentailOracle = "502f8beb-0132-492f-bbfb-acf1ea532967"
	s82b13SailbackOracle   = "17e74c4b-435c-469e-a9a4-32bf0cb7634c"
	s82b13OracleOfTragedy  = "f88e1616-a7c7-4122-b626-bfa799cccaa4"
	s82b13HawkerOracle     = "215fd0b4-5817-4335-a0ed-1694e44bde41"
	s82b13BruiserOracle    = "2976f47b-9626-4934-bcee-7ee7f57944c3"
	s82b13RagerOracle      = "d1a77708-7ddf-4645-bf1e-1d8147a57f23"
	s82b13WinderOracle     = "13c4437e-48a9-4c51-9b76-9e83e3ec362f"
	s82b13DisturbingOracle = "adde6737-516c-4d6f-9363-388ad37d4114"
	s82b13FailedOracle     = "e0930791-72b0-4d8b-b242-7b4094fed14e"
)

func TestS82B13CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82b13LookoutOracle:    "Expedition Lookout",
		s82b13WayfinderOracle:  "Deeproot Wayfinder",
		s82b13MolochOracle:     "Sandstalker Moloch",
		s82b13RottentailOracle: "Placid Rottentail",
		s82b13SailbackOracle:   "Ravenous Sailback",
		s82b13OracleOfTragedy:  "Oracle of Tragedy",
		s82b13HawkerOracle:     "Omen Hawker",
		s82b13BruiserOracle:    "Scrappy Bruiser",
		s82b13RagerOracle:      "Stormclaw Rager",
		s82b13WinderOracle:     "Zephyr Winder",
		s82b13DisturbingOracle: "Disturbing Conversion",
		s82b13FailedOracle:     "Failed Conversion",
	} {
		def, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != effects.CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, def.Name, def.Completeness, def.Caveats)
		}
	}
}

// s82b13FillGraveyard puts n filler cards in p's graveyard.
func s82b13FillGraveyard(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		effectstest.PushGraveyardCardForTest(p, "Filler")
	}
}

// --- Omen Hawker -----------------------------------------------------------

func TestS82B13OmenHawkerManaPaysOnlyForAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hawker := effectstest.B12Push(g, me.ID, "Omen Hawker", "Creature — Octopus Advisor", s82b13HawkerOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, hawker, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if got := me.ManaPool.Total(); got != 2 {
		t.Fatalf("pool %d, want {C}{U}", got)
	}
	cost, _ := game.ParseCost("{1}{U}")
	bear := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{U}"}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) {
		t.Error("the mana paid for a spell")
	}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(bear)) {
		t.Error("the mana refused an activated ability")
	}
}

// --- Expedition Lookout ----------------------------------------------------

func TestS82B13ExpeditionLookoutAttacksAndIsUnblockableOnlyAgainstAFullGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lookout := effectstest.B12Push(g, me.ID, "Expedition Lookout", "Creature — Merfolk Rogue", s82b13LookoutOracle, 2, 3)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall of Stone", "Creature — Wall", 0, 8)
	if !effectstest.EffectiveAbilitiesContain(t, g, lookout, "defender") {
		t.Fatal("Expedition Lookout has no defender")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	s82b13FillGraveyard(opp, 7)
	effectstest.B12Creature(g, opp.ID, "Nudge", "Creature — Bear", 1, 1) // settles the layers
	if err := g.DeclareAttacker(lookout, opp.ID); err == nil {
		t.Fatal("Lookout attacked with seven cards in an opponent's graveyard")
	}
	// Eight cards in my own graveyard is not an opponent's.
	s82b13FillGraveyard(me, 8)
	effectstest.B12Creature(g, opp.ID, "Nudge", "Creature — Bear", 1, 1)
	if err := g.DeclareAttacker(lookout, opp.ID); err == nil {
		t.Fatal("Lookout attacked on the strength of its own controller's graveyard")
	}
	s82b13FillGraveyard(opp, 1)
	effectstest.B12Creature(g, opp.ID, "Nudge", "Creature — Bear", 1, 1)
	if err := g.DeclareAttacker(lookout, opp.ID); err != nil {
		t.Fatalf("Lookout could not attack with eight cards in an opponent's graveyard: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.BrRefusal(t, g.DeclareBlocker(blocker, lookout), game.BlockReasonCantBeBlocked)
}

func TestS82B13ExpeditionLookoutCanBeBlockedWithSevenCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lookout := effectstest.B12Push(g, me.ID, "Expedition Lookout", "Creature — Merfolk Rogue", s82b13LookoutOracle, 2, 3)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall of Stone", "Creature — Wall", 0, 8)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	s82b13FillGraveyard(opp, 8)
	effectstest.B12Creature(g, opp.ID, "Nudge", "Creature — Bear", 1, 1)
	if err := g.DeclareAttacker(lookout, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	// The graveyard drops below eight before blocks.
	g.WithWriteLock(func() { opp.Graveyard.Cards = opp.Graveyard.Cards[:7] })
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, lookout); err != nil {
		t.Fatalf("with seven cards it may be blocked: %v", err)
	}
}

// --- Zephyr Winder -----------------------------------------------------------

func TestS82B13ZephyrWinderUntapsAnUpToOneTargetCreatureOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	winder := effectstest.B12Push(g, me.ID, "Zephyr Winder", "Creature — Elemental", s82b13WinderOracle, 2, 1)
	if !effectstest.EffectiveAbilitiesContain(t, g, winder, "flying") {
		t.Fatal("Zephyr Winder has no flying")
	}
	friend := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, friend).Tapped = true })

	effectstest.AttackWith(t, g, opp.ID, winder)
	g.RunStateChecksForTest()
	effectstest.AnswerPickTargetIfOpen(t, g, friend)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, friend); c == nil || c.Tapped {
		t.Error("the targeted creature is still tapped")
	}
	if opp.Life != 38 {
		t.Errorf("the defender has %d life, want 38", opp.Life)
	}
}

// --- Failed Conversion -------------------------------------------------------

func TestS82B13FailedConversionShrinksAndSurveilsWhenTheHostDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 5, 4)
	effectstest.CastCatalogSpell(t, g, "Failed Conversion", "Enchantment — Aura", s82b13FailedOracle, effectstest.CardRefs(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(ogre) {
		t.Fatal("the 5/4 survived -4/-4")
	}
	effectstest.PassPriorityAroundTable(t, g)
	surveil := false
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID {
			surveil = true
		}
	}
	if !surveil {
		t.Error("the Aura's controller has no surveil prompt after the host died")
	}
}

func TestS82B13FailedConversionLeavesAToughCreatureShrunken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	wurm := effectstest.B12Creature(g, opp.ID, "Their Wurm", "Creature — Wurm", 6, 6)
	effectstest.CastCatalogSpell(t, g, "Failed Conversion", "Enchantment — Aura", s82b13FailedOracle, effectstest.CardRefs(wurm))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, wurm), effectstest.EffectiveToughness(t, g, wurm); p != 2 || tough != 2 {
		t.Errorf("the host is %d/%d, want 2/2", p, tough)
	}
}

// --- Disturbing Conversion ---------------------------------------------------

func TestS82B13DisturbingConversionMillsEachPlayerAndShrinksByTheControllersGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 6, 6)
	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = p.Library.Size()
	}
	effectstest.CastCatalogSpell(t, g, "Disturbing Conversion", "Enchantment — Aura", s82b13DisturbingOracle, effectstest.CardRefs(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if got := before[i] - p.Library.Size(); got != 2 {
			t.Errorf("seat %d milled %d, want 2", i, got)
		}
	}
	if got := effectstest.EffectivePower(t, g, ogre); got != 6-len(opp.Graveyard.Cards) {
		t.Errorf("host power %d with %d cards in its controller's graveyard, want %d", got, len(opp.Graveyard.Cards), 6-len(opp.Graveyard.Cards))
	}
	// More cards in its controller's graveyard, not the Aura's.
	s82b13FillGraveyard(opp, 1)
	effectstest.B12Creature(g, me.ID, "Nudge", "Creature — Bear", 1, 1) // settles the layers
	if got, want := effectstest.EffectivePower(t, g, ogre), 6-len(opp.Graveyard.Cards); got != want {
		t.Errorf("host power %d after another card hit its controller's graveyard, want %d", got, want)
	}
	if got := effectstest.EffectiveToughness(t, g, ogre); got != 6 {
		t.Errorf("host toughness %d, want 6 (-X/-0)", got)
	}
}

// --- Stormclaw Rager ---------------------------------------------------------

func TestS82B13StormclawRagerSacrificesAnotherCreatureOrArtifactForACounterAndACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rager := effectstest.B12Push(g, me.ID, "Stormclaw Rager", "Creature — Ogre Warrior", s82b13RagerOracle, 2, 2)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.Acrb1AddMana(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, rager, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err == nil {
		t.Fatal("activated outside a main phase")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.Acrb1AddMana(t, g, me.ID, "{C}")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, rager, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{rager}}); err == nil {
		t.Fatal("sacrificed the Rager to itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, rager, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("sacrifice a creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if got := effectstest.CounterOn(g, rager, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	effectstest.Acrb1AddMana(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, rager, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("sacrifice an artifact: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, rager, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters after the artifact, want 2", got)
	}
}

// --- Placid Rottentail ---------------------------------------------------------

func TestS82B13PlacidRottentailPutsTwoCountersFromTheGraveyardAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vig := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	tail := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: tail, Name: "Placid Rottentail", TypeLine: "Creature — Fungus Rabbit", OracleID: s82b13RottentailOracle,
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	effectstest.Acrb1AddMana(t, g, me.ID, "{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, tail, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(vig), Strict: true}); err == nil {
		t.Fatal("activated outside a main phase")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.Acrb1AddMana(t, g, me.ID, "{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, tail, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(vig), Strict: true}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, vig, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
	if me.Graveyard.Contains(tail) {
		t.Error("the Rottentail is still in the graveyard (exiled as a cost)")
	}
}

func TestS82B13PlacidRottentailHasVigilanceAndCantActivateFromTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tail := effectstest.B12Push(g, me.ID, "Placid Rottentail", "Creature — Fungus Rabbit", s82b13RottentailOracle, 1, 1)
	if !effectstest.EffectiveAbilitiesContain(t, g, tail, "vigilance") {
		t.Error("no vigilance")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.Acrb1AddMana(t, g, me.ID, "{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, tail, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(tail), Strict: true}); err == nil {
		t.Error("the graveyard ability activated from the battlefield")
	}
}

// --- Ravenous Sailback ---------------------------------------------------------

func TestS82B13RavenousSailbackHasteMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.CastHoldCreature(t, g, "Ravenous Sailback", "Creature — Dinosaur", s82b13SailbackOracle, 3, 4)
	_ = me
	// With nothing to destroy only the haste bullet is offered.
	effectstest.ChooseModeNow(t, g, g.Seats[g.Turn.ActiveSeat].ID, []int{0}, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, id, "haste") {
		t.Error("the Sailback did not gain haste")
	}
	effectstest.DftMPPassTurn(t, g)
	if effectstest.EffectiveAbilitiesContain(t, g, id, "haste") {
		t.Error("the haste outlasted the turn")
	}
}

func TestS82B13RavenousSailbackDestroysAnArtifactOrEnchantment(t *testing.T) {
	for _, tc := range []struct{ Name, TypeLine string }{{"Rock", "Artifact"}, {"Pact", "Enchantment"}} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		target := effectstest.B12Permanent(g, opp.ID, tc.Name, tc.TypeLine)
		bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		effectstest.CastHoldCreatureNoResolve(t, g, "Ravenous Sailback", "Creature — Dinosaur", s82b13SailbackOracle, 3, 4)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.ChooseModeNow(t, g, g.Seats[g.Turn.ActiveSeat].ID, []int{0, 1}, 1)
		effectstest.AnswerPickTargetIfOpen(t, g, target)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(target) {
			t.Errorf("%s survived", tc.Name)
		}
		if !g.Battlefield.Contains(bear) {
			t.Error("a creature was destroyed")
		}
	}
}
