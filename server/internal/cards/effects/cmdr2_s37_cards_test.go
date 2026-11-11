package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	whisperwoodOracle = "8ec7c7ce-23a5-4ba7-8cfd-4caf8601a9c8"
	temurShamanOracle = "fb683e18-50b8-4821-9a1a-86c6dea63ac9"
	masteryUnseenOrac = "ae0701c1-7a99-46fe-bcb1-ea33ad2a0fc9"
)

// s37Top puts a creature card on top of p's library and returns its id.
func s37Top(p *game.Player, name string) uuid.UUID {
	return mlTop(p, name, "Creature — Bear", "{1}")
}

// s37ManifestNow manifests the top card of p's library through the engine
// and returns the manifested permanent's id.
func s37ManifestNow(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
	t.Helper()
	id := s37Top(p, "Hidden Bear")
	if _, err := g.ManifestForEffect(p.ID); err != nil {
		t.Fatalf("manifest: %v", err)
	}
	if c, ok := battlefieldCard(g, id); !ok || !c.FaceDown {
		t.Fatal("the card is not a face-down permanent")
	}
	return id
}

func TestWhisperwoodElementalManifestsAtYourEndStepOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Whisperwood Elemental", "Creature — Elemental", whisperwoodOracle, false)
	top := s37Top(me, "Top Bear")
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	mlManifested(t, g, top, me, me)
}

func TestWhisperwoodElementalSacrificeGivesFaceUpNontokenCreaturesADiesManifest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	elem := pushCatalogPermanent(g, me.ID, "Whisperwood Elemental", "Creature — Elemental", whisperwoodOracle, false)
	bear := b12Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	down := s37ManifestNow(t, g, me)
	if err := g.ActivateCatalogAbility(me.ID, elem, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, elem); ok {
		t.Fatal("the Elemental is still on the battlefield")
	}
	// The face-down manifest did not get the grant; the bear did.
	g.WithWriteLock(func() { _ = down })
	top := s37Top(me, "Replacement")
	if err := g.DestroyPermanentForEffect(bear); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	mlManifested(t, g, top, me, me)
}

func TestWhisperwoodElementalGrantSkipsFaceDownCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	elem := pushCatalogPermanent(g, me.ID, "Whisperwood Elemental", "Creature — Elemental", whisperwoodOracle, false)
	down := s37ManifestNow(t, g, me)
	if err := g.ActivateCatalogAbility(me.ID, elem, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)
	top := s37Top(me, "Replacement")
	if err := g.DestroyPermanentForEffect(down); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, top); ok {
		t.Fatal("a face-down creature got the dies trigger")
	}
}

func TestMasteryOfTheUnseenManifestsForFourAndGainsLifeWhenAPermanentTurnsUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	mastery := pushCatalogPermanent(g, me.ID, "Mastery of the Unseen", "Enchantment", masteryUnseenOrac, false)
	top := s37Top(me, "Top Bear")
	// Refused without the mana.
	if err := g.ActivateCatalogAbility(me.ID, mastery, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without paying {3}{W}")
	}
	floatForTest(g, me, "WCCC")
	if err := g.ActivateCatalogAbility(me.ID, mastery, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	mlManifested(t, g, top, me, me)
	b12Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	life := me.Life
	g.WithWriteLock(func() { me.ManaPool.AddMana(manaTokens("C")...) })
	if err := g.PerformSpecialAction(me.ID, top, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up: %v", err)
	}
	g.RunStateChecksForTest()
	b30SettleOrder(t, g)
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2 (the bear and the turned-up creature)", got)
	}
}

func TestTemurWarShamanManifestsOnEntryAndTurnedUpCreatureFights(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Temur War Shaman", "Creature — Human Shaman", temurShamanOracle, false)
	victim := b12Push(g, opp.ID, "Victim", "Creature — Bear", "", 2, 2)
	top := s37Top(me, "Fighter")
	g.WithWriteLock(func() { me.Library.Cards[0].Power, me.Library.Cards[0].Toughness = 4, 4 })
	if _, err := g.ManifestForEffect(me.ID); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { me.ManaPool.AddMana(manaTokens("C")...) })
	if err := g.PerformSpecialAction(me.ID, top, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up: %v", err)
	}
	g.RunStateChecksForTest()
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: victim}}); err != nil {
			t.Fatalf("pick target: %v", err)
		}
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, victim); ok {
		t.Fatal("the fight did not kill the victim")
	}
	if c, ok := battlefieldCard(g, top); !ok || c.DamageMarked < 2 {
		t.Errorf("the turned-up creature did not take the victim's damage: %+v", c.DamageMarked)
	}
}

const (
	uginsMasteryOracle  = "d6158860-ada8-4e39-8156-23dee2f68f79"
	primordialMistOracl = "afc9e1d3-b012-49e9-8d22-89abbee03c22"
	trailOfMysteryOracl = "0e09aaba-5c32-4a45-8a07-b69e873510ef"
	obscuringAetherOrac = "70bfb805-2009-44c8-b5bd-2ca281637566"
	panopticOracle      = "1e1a787a-2294-472d-908d-d4d72821a7cc"
)

func TestUginsMasteryManifestsOnlyForColorlessCreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Ugin's Mastery", "Enchantment", uginsMasteryOracle, false)
	top := s37Top(me, "Top Bear")
	// A green creature spell does not trigger it.
	green := handCardFull(me, "Bear", "Creature — Bear", "{1}{G}", "", []string{"G"})
	floatForTest(g, me, "CG")
	if err := g.CastSpell(me.ID, green, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast green: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, top); ok {
		t.Fatal("a colored creature spell manifested")
	}
	golem := handCardFull(me, "Golem", "Artifact Creature — Golem", "{2}", "", nil)
	floatForTest(g, me, "CC")
	if err := g.CastSpell(me.ID, golem, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast golem: %v", err)
	}
	passPriorityAroundTable(t, g)
	mlManifested(t, g, top, me, me)
}

func TestUginsMasteryAttackWithSixPowerOffersATurnUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Ugin's Mastery", "Enchantment", uginsMasteryOracle, false)
	a := b12Push(g, me.ID, "Big A", "Creature — Giant", "", 3, 3)
	b := b12Push(g, me.ID, "Big B", "Creature — Giant", "", 3, 3)
	down := s37ManifestNow(t, g, me)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if c := chooseCardsChoiceFor(g, me.ID); c != nil {
		t.Fatal("3 total power triggered it")
	}
	_ = b
	_ = down
}

func TestUginsMasteryTurnsAFaceDownCreatureUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Ugin's Mastery", "Enchantment", uginsMasteryOracle, false)
	a := b12Push(g, me.ID, "Big A", "Creature — Giant", "", 3, 3)
	b := b12Push(g, me.ID, "Big B", "Creature — Giant", "", 3, 3)
	down := s37ManifestNow(t, g, me)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, down)
	passPriorityAroundTable(t, g)
	if c := findCardAnywhere(t, g, down); c.FaceDown {
		t.Fatal("the face-down creature is still face down")
	}
}

func TestTrailOfMysterySearchesOnFaceDownEntryAndPumpsTurnedUpCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Trail of Mystery", "Enchantment", trailOfMysteryOracl, false)
	forest := stapleLibraryCard(me, "Forest", "Basic Land — Forest")
	down := s37Top(me, "Hidden Bear")
	if _, err := g.ManifestForEffect(me.ID); err != nil {
		t.Fatal(err)
	}
	g.RunStateChecksForTest()
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) == nil {
		t.Fatal("no library search was offered")
	}
	answerSearchByID(t, g, me.ID, forest)
	if !me.Hand.Contains(forest) {
		t.Fatal("the basic land was not put into hand")
	}
	g.WithWriteLock(func() { me.ManaPool.AddMana(manaTokens("C")...) })
	if err := g.PerformSpecialAction(me.ID, down, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up: %v", err)
	}
	g.RunStateChecksForTest()
	b30SettleOrder(t, g)
	if p, tt := mkmMPPT(t, g, down); p != 6 || tt != 6 {
		t.Errorf("turned-up creature is %d/%d, want 6/6 (4/4 +2/+2)", p, tt)
	}
}

func TestObscuringAetherReducesFaceDownSpellsAndTurnsItselfFaceDown(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	aether := pushCatalogPermanent(g, me.ID, "Obscuring Aether", "Enchantment", obscuringAetherOrac, false)
	id := handCardFull(me, "Pyrotechnic Performer", mkmMPPerformerTypeLine, "{1}{R}", mkmMPPerformerOracle, []string{"R"})
	floatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "disguise", Strict: true}); err == nil {
		t.Fatal("a face-down cast with one mana (cost {3}-{1}={2}) was allowed")
	}
	floatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "disguise", Strict: true}); err != nil {
		t.Fatalf("face-down cast for {2}: %v", err)
	}
	// Turn the enchantment face down.
	floatForTest(g, me, "CG")
	if err := g.ActivateCatalogAbility(me.ID, aether, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, aether)
	if !ok || !c.FaceDown || !c.IsCreature() || c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Fatalf("Obscuring Aether did not become a face-down 2/2 creature: %+v ok=%v", c.FaceDown, ok)
	}
}

func TestPanopticProjektorMakesTheNextFaceDownSpellCostThreeLess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	proj := pushCatalogPermanent(g, me.ID, "Panoptic Projektor", "Artifact", panopticOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, proj, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("tap: %v", err)
	}
	passPriorityAroundTable(t, g)
	id := handCardFull(me, "Pyrotechnic Performer", mkmMPPerformerTypeLine, "{1}{R}", mkmMPPerformerOracle, []string{"R"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "disguise", Strict: true}); err != nil {
		t.Fatalf("free face-down cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := findCardAnywhere(t, g, id); !c.FaceDownIsPermanent() {
		t.Fatal("it did not enter face down")
	}
	// Only the next one.
	id2 := handCardFull(me, "Pyrotechnic Performer", mkmMPPerformerTypeLine, "{1}{R}", mkmMPPerformerOracle, []string{"R"})
	if err := g.CastSpell(me.ID, id2, game.CastSpellParams{AlternativeCost: "disguise", Strict: true}); err == nil {
		t.Fatal("a second face-down spell was free")
	}
}

func TestPanopticProjektorDoublesATurnUpTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Panoptic Projektor", "Artifact", panopticOracle, false)
	pushCatalogPermanent(g, me.ID, "Mastery of the Unseen", "Enchantment", masteryUnseenOrac, false)
	perf := mkmMPCastFaceDown(t, g, "Pyrotechnic Performer", mkmMPPerformerTypeLine, mkmMPPerformerOracle, 3, 2)
	life := me.Life
	mkmMPTurnFaceUp(t, g, me, perf, "R")
	g.RunStateChecksForTest()
	b30SettleOrder(t, g)
	b30SettleOrder(t, g)
	// One creature: each Mastery trigger gains 1; doubled gives 2.
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2 (one creature, trigger twice)", got)
	}
}

func TestPrimordialMistManifestsAndExilesAFaceDownPermanentToPlayIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mist := pushCatalogPermanent(g, me.ID, "Primordial Mist", "Enchantment", primordialMistOracl, false)
	// A face-up permanent is not a cost.
	advanceToMain(t, g)
	forest := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	if _, err := g.ManifestForEffect(me.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, mist, 0, game.ActivateAbilityParams{Strict: true, ExilePermanentIDs: []uuid.UUID{mist}}); err == nil {
		t.Fatal("exiled a face-up permanent for the cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, mist, 0, game.ActivateAbilityParams{Strict: true, ExilePermanentIDs: []uuid.UUID{forest}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := findCardAnywhere(t, g, forest); c.FaceDown {
		t.Fatal("the exiled card is still face down")
	}
	if err := g.CastSpell(me.ID, forest, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("play the exiled land: %v", err)
	}
	if _, ok := battlefieldCard(g, forest); !ok {
		t.Fatal("the land is not on the battlefield")
	}
}

func TestPrimordialMistEndStepManifestIsOptional(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Primordial Mist", "Enchantment", primordialMistOracl, false)
	top := s37Top(me, "Top Bear")
	advanceToEndStep(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	mlManifested(t, g, top, me, me)
}
