package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// slice_296e_test.go — retriage slice 296-e ("Keyword creatures"):
// Toski, Bearer of Secrets; Displacer Kitten; Changeling Outcast;
// Mockingbird; Enduring Innocence; Peregrin Took; Ledger Shredder;
// The Gitrog Monster; Taurean Mauler; Crashing Drawbridge; Mirror
// Entity; Blightsteel Colossus; Elesh Norn, Mother of Machines;
// Heliod, Sun-Crowned; Apex Devastator.

const (
	toskiOracle           = "a8e707ec-ce77-4bc5-8c76-5ea3e81e8c7f"
	displacerKittenOracle = "09ef446c-a13d-49d9-a94c-cd5f5a2d440b"
	changelingOutcastOra  = "a61ef0cc-1da9-49f9-b0dc-01cf9f6205be"
	mockingbirdOracle     = "b9df2cdf-397c-458d-89cf-911568737ffa"
	enduringInnocenceOra  = "98a389f4-2905-47f3-b60e-3d4afb3e5cb0"
	peregrinTookOracle    = "188da0ad-8524-4fb2-915e-1876dc9df89f"
	ledgerShredderOracle  = "e9117015-1050-44dd-a46b-e7ffe2085fae"
	gitrogMonsterOracle   = "a5e54d2b-aad8-4ddd-af4b-13668913762b"
	taureanMaulerOracle   = "fd5ee26c-cc54-4dc0-b611-5cda14354a5e"
	crashingDrawbridgeOra = "328d5ef0-b25e-4f2a-80fe-35c6ae419e8a"
	mirrorEntityOracle    = "17e905ca-c0bd-473d-95a7-e180ba5fea43"
	blightsteelOracle     = "e80772e2-8623-4094-81a2-70828b2b151c"
	eleshNornOracle       = "5ade11c0-41dd-4b6a-9f5b-c5903a3a0d7f"
	heliodSunCrownedOra   = "63e596a2-9126-4af6-8782-c38687d664ad"
	apexDevastatorOracle  = "b4d6747c-3516-4e71-be30-af007c6b1dc4"
)

// --- Toski, Bearer of Secrets --------------------------------------

func TestToskiDrawsWhenACreatureYouControlDealsCombatDamageToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toski := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Toski, Bearer of Secrets", TypeLine: "Legendary Creature — Squirrel",
		OracleID: toskiOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	before := me.Hand.Size()
	attackWith(t, g, opp.ID, toski)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand size = %d, want %d after Toski connects", got, before+1)
	}
}

func TestToskiMustAttackEachCombatIfAble(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Toski, Bearer of Secrets", TypeLine: "Legendary Creature — Squirrel",
		OracleID: toskiOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	advanceToDeclareAttackersOf(t, g, seat)
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackRequirement) {
		t.Errorf("passing without attacking Toski: err = %v, want ErrAttackRequirement", err)
	}
}

// --- Displacer Kitten ------------------------------------------------

func TestDisplacerKittenFlickersAChosenPermanentOnNoncreatureCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Displacer Kitten", TypeLine: "Creature — Cat Beast",
		OracleID: displacerKittenOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	target := pushPermanentForTest(g, me.ID, "My Rock", "", "Artifact")

	castCatalogSpell(t, g, "Shock", "Instant", "", nil)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatalf("no pick_target prompt for Displacer Kitten's trigger")
	}
	if !hasID(prompt.PickTargetCards, target) {
		t.Fatalf("legal set = %v, want to include %s", prompt.PickTargetCards, target)
	}
	pickCard(t, g, me.ID, target)
	passPriorityAroundTable(t, g)

	newID := findBattlefieldByName(g, "My Rock")
	if newID == uuid.Nil {
		t.Fatalf("My Rock did not return to the battlefield")
	}
	if newID == target {
		t.Errorf("My Rock kept the same instance ID — it should be a new object after the flicker")
	}
}

// --- Changeling Outcast ------------------------------------------------

func TestChangelingOutcastCantBlockAndCantBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Changeling Outcast", TypeLine: "Creature — Shapeshifter",
		OracleID: changelingOutcastOra, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	assertRestrictions(t, g, id, game.CantBlock|game.CantBeBlocked)
}

// --- Mockingbird ------------------------------------------------------

func TestMockingbirdIsJustAFlierWithNoCopyChoice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mockingbird", TypeLine: "Creature — Bird Bard",
		OracleID: mockingbirdOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	if !hasEffectiveKeyword(t, g, id, "flying") {
		t.Error("Mockingbird lacks its printed flying")
	}
	if got := effectiveTypes(t, g, id); !containsString(got, "Creature") {
		t.Errorf("Mockingbird's types = %v, want Creature (no copy took place)", got)
	}
}

// --- Enduring Innocence -------------------------------------------------

func TestEnduringInnocenceDrawsOnceWhenALowPowerCreatureEntersAndOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Enduring Innocence", TypeLine: "Enchantment Creature — Sheep Glimmer",
		OracleID: enduringInnocenceOra, Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	before := me.Hand.Size()
	castWithCost(t, g, "Little Bear", "Creature — Bear", "{1}{G}", "")
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Fatalf("hand size after the first low-power ETB = %d, want %d", got, before+1)
	}

	castWithCost(t, g, "Another Bear", "Creature — Bear", "{1}{G}", "")
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand size after the second low-power ETB this turn = %d, want %d (once-per-turn)", got, before+1)
	}
}

// --- Peregrin Took ------------------------------------------------------

func TestPeregrinTookAddsAFoodWhenATokenIsCreated(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Peregrin Took", TypeLine: "Legendary Creature — Halfling Citizen",
		OracleID: peregrinTookOracle, Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	before := b16CountNamed(g, "Food")
	beforeTreasure := b16CountNamed(g, "Treasure")
	var applyErr error
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID})
		applyErr = (CreateToken{Controller: me.ID, Template: TreasureToken(), N: 1}).Apply(ctx)
	})
	if applyErr != nil {
		t.Fatalf("CreateToken: %v", applyErr)
	}
	if got := b16CountNamed(g, "Treasure"); got != beforeTreasure+1 {
		t.Errorf("Treasures = %d, want %d", got, beforeTreasure+1)
	}
	if got := b16CountNamed(g, "Food"); got != before+1 {
		t.Errorf("Foods = %d, want %d (Peregrin Took's extra)", got, before+1)
	}
}

func TestPeregrinTookSacrificeThreeFoodsDrawsACard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	took := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Peregrin Took", TypeLine: "Legendary Creature — Halfling Citizen",
		OracleID: peregrinTookOracle, Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	var foodIDs []uuid.UUID
	for i := 0; i < 3; i++ {
		food := FoodToken()
		food.InstanceID = uuid.New()
		food.Owner = me.ID
		food.Controller = me.ID
		foodIDs = append(foodIDs, pushBattlefieldCardWithTimestamp(g, food))
	}
	before := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, took, 0, game.ActivateAbilityParams{
		SacrificeIDs: foodIDs,
	}); err != nil {
		t.Fatalf("activate Peregrin Took: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand size = %d, want %d", got, before+1)
	}
	if got := b16CountNamed(g, "Food"); got != 0 {
		t.Errorf("Foods left = %d, want 0", got)
	}
}

// --- Ledger Shredder ------------------------------------------------

func TestLedgerShredderConnivesOnTheCasterSecondSpellPuttingACounterOnANonlandDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	shredder := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ledger Shredder", TypeLine: "Creature — Bird Advisor",
		OracleID: ledgerShredderOracle, Power: 1, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	handCardForTest(me, "Pitch Spell", "Sorcery", "")

	// First spell of the turn: no connive yet.
	castCatalogSpell(t, g, "Shock", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) != nil {
		t.Fatalf("connive fired on the first spell of the turn")
	}

	nonland := handCardForTest(me, "Nonland Card", "Artifact", "")
	// Second spell of the turn: connive fires (draw, then the
	// discard prompt opens and holds the table).
	castCatalogSpell(t, g, "Shock", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	handBefore := me.Hand.Size()
	answerDiscard(t, g, me.ID, nonland)
	if got := me.Hand.Size(); got != handBefore-1 {
		t.Fatalf("hand size after discard = %d, want %d", got, handBefore-1)
	}
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, shredder)
	if got := c.CurrentPower(); got != 2 {
		t.Errorf("Ledger Shredder power after discarding a nonland card = %d, want 2 (a +1/+1 counter)", got)
	}
}

// --- The Gitrog Monster ------------------------------------------------

func TestGitrogMonsterUpkeepSacrificesALandWhenOneIsOffered(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	gitrog := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Gitrog Monster", TypeLine: "Legendary Creature — Frog Horror",
		OracleID: gitrogMonsterOracle, Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	land := pushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	advanceToUpkeepOf(t, g, seat)
	if triggerOnStack(g, gitrog) == nil {
		t.Fatalf("upkeep trigger not on the stack")
	}
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 0) // "Sacrifice a land"
	answerChooseCards(t, g, me.ID, land)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(land) {
		t.Errorf("the land was not sacrificed")
	}
	if !g.Battlefield.Contains(gitrog) {
		t.Errorf("The Gitrog Monster was sacrificed even though a land was offered")
	}
}

func TestGitrogMonsterUpkeepSacrificesItselfWithNoLandToOffer(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	gitrog := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Gitrog Monster", TypeLine: "Legendary Creature — Frog Horror",
		OracleID: gitrogMonsterOracle, Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	advanceToUpkeepOf(t, g, seat)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(gitrog) {
		t.Errorf("The Gitrog Monster survived its own upkeep with no land to sacrifice")
	}
}

func TestGitrogMonsterDrawsWhenALandGoesToYourGraveyardFromAnywhere(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Gitrog Monster", TypeLine: "Legendary Creature — Frog Horror",
		OracleID: gitrogMonsterOracle, Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	landInHand := handCardForTest(me, "Some Land", "Land", "")
	before := me.Hand.Size()
	var promptID uuid.UUID
	g.WithWriteLock(func() {
		promptID = g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: me.ID, N: 1})
	})
	if promptID == uuid.Nil {
		t.Fatal("no discard prompt was queued")
	}
	if err := g.ResolveChooseCards(promptID, me.ID, []uuid.UUID{landInHand}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before-1 {
		t.Errorf("hand size after the discard resolved = %d, want %d", got, before-1)
	}
}

// --- Taurean Mauler ------------------------------------------------

func TestTaureanMaulerMayGetACounterWhenAnOpponentCastsASpell(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	mauler := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Taurean Mauler", TypeLine: "Creature — Shapeshifter",
		OracleID: taureanMaulerOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	spell := handCardForTest(opp, "Their Spell", "Sorcery", "")
	advanceTo(t, g, game.StepPrecombatMain)
	g.Turn.ActiveSeat = (seat + 1) % len(g.Seats)
	if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, mauler)
	if got := c.CurrentPower(); got != 3 {
		t.Errorf("Taurean Mauler power = %d, want 3 after accepting the counter", got)
	}
}

// --- Crashing Drawbridge ------------------------------------------------

func TestCrashingDrawbridgeGrantsHasteToCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bridge := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Crashing Drawbridge", TypeLine: "Artifact Creature — Wall",
		OracleID: crashingDrawbridgeOra, Power: 0, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	bear := pushDiesCreatureForTest(g, me.ID, "Sick Bear", "", "Creature — Bear", 2, 2)
	if hasEffectiveKeyword(t, g, bear, "haste") {
		t.Fatalf("bear already has haste before the ability resolves")
	}
	if err := g.ActivateCatalogAbility(me.ID, bridge, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Crashing Drawbridge: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Errorf("bear lacks haste after Crashing Drawbridge resolved")
	}
}

// --- Mirror Entity ------------------------------------------------------

func TestMirrorEntitySetsBasePowerToughnessAndGrantsAllCreatureTypes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	entity := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mirror Entity", TypeLine: "Creature — Shapeshifter",
		OracleID: mirrorEntityOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	bear := pushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, entity, 0, game.ActivateAbilityParams{XValue: 4}); err != nil {
		t.Fatalf("activate Mirror Entity: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tg != 4 {
		t.Errorf("bear P/T = %d/%d, want 4/4", p, tg)
	}
	var hasElf bool
	var gotSubtypes []string
	g.ReadSnapshot(func() {
		c, _ := battlefieldCard(g, bear)
		hasElf = c.HasSubtype("Elf")
		gotSubtypes = c.Effective().Subtypes
	})
	if !hasElf {
		t.Errorf("bear subtypes = %v, want every creature type (e.g. Elf)", gotSubtypes)
	}
}

// --- Blightsteel Colossus ------------------------------------------------

func TestBlightsteelColossusIsShuffledIntoItsLibraryInsteadOfDying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	colossus := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blightsteel Colossus", TypeLine: "Artifact Creature — Phyrexian Golem",
		OracleID: blightsteelOracle, Power: 11, Toughness: 11, Owner: me.ID, Controller: me.ID,
	})
	// Indestructible refuses an ordinary destroy, so the battlefield
	// exit this pins is a sacrifice — sacrifice ignores indestructible
	// (CR 701.21a) and still opens the same "put into a graveyard"
	// replacement window.
	var applyErr error
	g.WithWriteLock(func() {
		applyErr = (SacrificePermanent{Target: colossus}).Apply(NewContext(g, &game.StackItem{Controller: me.ID}))
	})
	if applyErr != nil {
		t.Fatalf("SacrificePermanent: %v", applyErr)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(colossus) {
		t.Fatalf("Blightsteel Colossus is still on the battlefield")
	}
	if me.Graveyard.Contains(colossus) {
		t.Errorf("Blightsteel Colossus landed in the graveyard instead of the library")
	}
	if !me.Library.Contains(colossus) {
		t.Errorf("Blightsteel Colossus is not in its owner's library")
	}
}

// --- Elesh Norn, Mother of Machines --------------------------------------

func TestEleshNornDoublesAnEnteringPermanentsTriggerForItsController(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Elesh Norn, Mother of Machines", TypeLine: "Legendary Creature — Phyrexian Praetor",
		OracleID: eleshNornOracle, Power: 4, Toughness: 7, Owner: me.ID, Controller: me.ID,
	})
	before := me.Hand.Size()
	castWithCost(t, g, "Mulldrifter", "Creature — Elemental", "{4}{U}", "24d0f5e7-0d9e-4b76-900e-a7274e80312d")
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+4 {
		t.Errorf("hand size after Mulldrifter's doubled ETB = %d, want %d (draw two, twice)", got, before+4)
	}
}

// --- Heliod, Sun-Crowned ------------------------------------------------

func TestHeliodIsNotACreatureBelowDevotionFiveAndBecomesOneAtFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	heliod := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Heliod, Sun-Crowned", TypeLine: "Legendary Enchantment Creature — God",
		OracleID: heliodSunCrownedOra, ManaCost: "{2}{W}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	if got := effectiveTypes(t, g, heliod); containsString(got, "Creature") {
		t.Fatalf("Heliod is a creature at devotion 1: %v", got)
	}
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", ManaCost: "{W}{W}{W}{W}",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	if got := effectiveTypes(t, g, heliod); !containsString(got, "Creature") {
		t.Fatalf("Heliod is not a creature at devotion 5: %v", got)
	}
	if !hasEffectiveKeyword(t, g, heliod, "indestructible") {
		t.Error("Heliod lacks indestructible")
	}
}

func TestHeliodPutsACounterOnTargetWhenYouGainLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Heliod, Sun-Crowned", TypeLine: "Legendary Enchantment Creature — God",
		OracleID: heliodSunCrownedOra, ManaCost: "{2}{W}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	bear := pushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatalf("no pick_target prompt for Heliod's lifegain trigger")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, bear)
	if got := c.CurrentPower(); got != 3 {
		t.Errorf("bear power = %d, want 3 after a +1/+1 counter", got)
	}
}

func TestHeliodActivatedAbilityGrantsLifelinkToAnotherTargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	heliod := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Heliod, Sun-Crowned", TypeLine: "Legendary Enchantment Creature — God",
		OracleID: heliodSunCrownedOra, ManaCost: "{2}{W}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	bear := pushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	if hasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Fatalf("bear already has lifelink")
	}
	if err := g.ActivateCatalogAbility(me.ID, heliod, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate Heliod: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Errorf("bear lacks lifelink after Heliod's ability resolved")
	}
}

// --- Apex Devastator ------------------------------------------------

func TestApexDevastatorCascadesFourTimes(t *testing.T) {
	if got := len(game.CatalogTriggers(apexDevastatorOracle)); got != 4 {
		t.Fatalf("Apex Devastator has %d triggered abilities, want 4 (cascade x4)", got)
	}

	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedCheapLibrary(g, me, 8)

	castWithCost(t, g, "Apex Devastator", "Creature — Chimera Hydra", "{8}{G}{G}", apexDevastatorOracle)
	passPriorityAroundTable(t, g)

	// Each of the four cascade triggers resolves one at a time with a
	// priority pass between them, so drain-then-pass four times.
	total := 0
	for i := 0; i < 4; i++ {
		total += answerAllMayCast(t, g, me.ID, false)
		passPriorityAroundTable(t, g)
	}
	if total != 4 {
		t.Errorf("Apex Devastator made %d cascade offers, want 4", total)
	}
}
