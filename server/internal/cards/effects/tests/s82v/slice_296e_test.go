package testss82v

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Toski, Bearer of Secrets --------------------------------------

func TestToskiDrawsWhenACreatureYouControlDealsCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toski := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Toski, Bearer of Secrets", TypeLine: "Legendary Creature — Squirrel",
		OracleID: effectstest.ToskiOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	before := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, toski)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand size = %d, want %d after Toski connects", got, before+1)
	}
}

func TestToskiMustAttackEachCombatIfAble(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Toski, Bearer of Secrets", TypeLine: "Legendary Creature — Squirrel",
		OracleID: effectstest.ToskiOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackRequirement) {
		t.Errorf("passing without attacking Toski: err = %v, want ErrAttackRequirement", err)
	}
}

// --- Displacer Kitten ------------------------------------------------

func TestDisplacerKittenFlickersAChosenPermanentOnNoncreatureCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Displacer Kitten", TypeLine: "Creature — Cat Beast",
		OracleID: effectstest.DisplacerKittenOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	target := effectstest.PushPermanentForTest(g, me.ID, "My Rock", "", "Artifact")

	effectstest.CastCatalogSpell(t, g, "Shock", "Instant", "", nil)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatalf("no pick_target prompt for Displacer Kitten's trigger")
	}
	if !effectstest.HasID(prompt.PickTargetCards, target) {
		t.Fatalf("legal set = %v, want to include %s", prompt.PickTargetCards, target)
	}
	effectstest.PickCard(t, g, me.ID, target)
	effectstest.PassPriorityAroundTable(t, g)

	newID := effectstest.FindBattlefieldByName(g, "My Rock")
	if newID == uuid.Nil {
		t.Fatalf("My Rock did not return to the battlefield")
	}
	if newID == target {
		t.Errorf("My Rock kept the same instance ID — it should be a new object after the flicker")
	}
}

// --- Changeling Outcast ------------------------------------------------

func TestChangelingOutcastCantBlockAndCantBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Changeling Outcast", TypeLine: "Creature — Shapeshifter",
		OracleID: effectstest.ChangelingOutcastOra, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AssertRestrictions(t, g, id, game.CantBlock|game.CantBeBlocked)
}

// --- Mockingbird ------------------------------------------------------

// The printed baseline. The copy choice — and its "amount of mana
// spent" ceiling, #1735 — is pinned in mana_spent_total_test.go.
func TestMockingbirdIsAPrintedFlier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mockingbird", TypeLine: "Creature — Bird Bard",
		OracleID: effectstest.MockingbirdOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	if !effectstest.HasEffectiveKeyword(t, g, id, "flying") {
		t.Error("Mockingbird lacks its printed flying")
	}
	if got := effectstest.EffectiveTypes(t, g, id); !effectstest.ContainsString(got, "Creature") {
		t.Errorf("Mockingbird's types = %v, want Creature (no copy took place)", got)
	}
}

// --- Enduring Innocence -------------------------------------------------

func TestEnduringInnocenceDrawsOnceWhenALowPowerCreatureEntersAndOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Enduring Innocence", TypeLine: "Enchantment Creature — Sheep Glimmer",
		OracleID: effectstest.EnduringInnocenceOra, Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	before := me.Hand.Size()
	effectstest.CastWithCost(t, g, "Little Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Fatalf("hand size after the first low-power ETB = %d, want %d", got, before+1)
	}

	effectstest.CastWithCost(t, g, "Another Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand size after the second low-power ETB this turn = %d, want %d (once-per-turn)", got, before+1)
	}
}

// TestEnduringInnocenceCountsTheCountersACreatureEntersWith: a printed
// 1/1 that enters with two +1/+1 counters has power 3 (CR 122.1a) and
// is not "power 2 or less". #2401's class of bug: Effective().Power
// stops before the counters.
func TestEnduringInnocenceCountsTheCountersACreatureEntersWith(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Enduring Innocence", TypeLine: "Enchantment Creature — Sheep Glimmer",
		OracleID: effectstest.EnduringInnocenceOra, Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	enter := func(name string, counters map[string]int) {
		id := uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: "Creature — Bear",
			Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
			Counters: counters,
		})
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventETB, CardID: id, Actor: me.ID})
		})
		effectstest.PassPriorityAroundTable(t, g)
	}
	before := me.Hand.Size()
	enter("Grown Bear", map[string]int{game.CounterPlusOne: 2})
	if got := me.Hand.Size(); got != before {
		t.Fatalf("hand size = %d, want %d: a 1/1 with two +1/+1 counters has power 3", got, before)
	}
	// The control: the same 1/1 with no counters does draw.
	enter("Small Bear", nil)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand size = %d, want %d: a 1/1 with no counters is power 2 or less", got, before+1)
	}
}

// --- Peregrin Took ------------------------------------------------------

func TestPeregrinTookAddsAFoodWhenATokenIsCreated(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Peregrin Took", TypeLine: "Legendary Creature — Halfling Citizen",
		OracleID: effectstest.PeregrinTookOracle, Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	before := effectstest.B16CountNamed(g, "Food")
	beforeTreasure := effectstest.B16CountNamed(g, "Treasure")
	var applyErr error
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{Controller: me.ID})
		applyErr = (effects.CreateToken{Controller: me.ID, Template: effects.TreasureToken(), N: 1}).Apply(ctx)
	})
	if applyErr != nil {
		t.Fatalf("CreateToken: %v", applyErr)
	}
	if got := effectstest.B16CountNamed(g, "Treasure"); got != beforeTreasure+1 {
		t.Errorf("Treasures = %d, want %d", got, beforeTreasure+1)
	}
	if got := effectstest.B16CountNamed(g, "Food"); got != before+1 {
		t.Errorf("Foods = %d, want %d (Peregrin Took's extra)", got, before+1)
	}
}

func TestPeregrinTookSacrificeThreeFoodsDrawsACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	took := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Peregrin Took", TypeLine: "Legendary Creature — Halfling Citizen",
		OracleID: effectstest.PeregrinTookOracle, Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	var foodIDs []uuid.UUID
	for i := 0; i < 3; i++ {
		food := effects.FoodToken()
		food.InstanceID = uuid.New()
		food.Owner = me.ID
		food.Controller = me.ID
		foodIDs = append(foodIDs, effectstest.PushBattlefieldCardWithTimestamp(g, food))
	}
	before := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, took, 0, game.ActivateAbilityParams{
		SacrificeIDs: foodIDs,
	}); err != nil {
		t.Fatalf("activate Peregrin Took: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand size = %d, want %d", got, before+1)
	}
	if got := effectstest.B16CountNamed(g, "Food"); got != 0 {
		t.Errorf("Foods left = %d, want 0", got)
	}
}

// --- Ledger Shredder ------------------------------------------------

func TestLedgerShredderConnivesOnTheCasterSecondSpellPuttingACounterOnANonlandDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	shredder := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ledger Shredder", TypeLine: "Creature — Bird Advisor",
		OracleID: effectstest.LedgerShredderOracle, Power: 1, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	effectstest.HandCardForTest(me, "Pitch Spell", "Sorcery", "")

	// First spell of the turn: no connive yet.
	effectstest.CastCatalogSpell(t, g, "Shock", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Fatalf("connive fired on the first spell of the turn")
	}

	nonland := effectstest.HandCardForTest(me, "Nonland Card", "Artifact", "")
	// Second spell of the turn: connive fires (draw, then the
	// discard prompt opens and holds the table).
	effectstest.CastCatalogSpell(t, g, "Shock", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	handBefore := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID, nonland)
	if got := me.Hand.Size(); got != handBefore-1 {
		t.Fatalf("hand size after discard = %d, want %d", got, handBefore-1)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, shredder)
	if got := c.CurrentPower(); got != 2 {
		t.Errorf("Ledger Shredder power after discarding a nonland card = %d, want 2 (a +1/+1 counter)", got)
	}
	// CR 701.50f, #4641: it connived once, and said so.
	if evs := effectstest.EventsOfKind(g, game.EventConnived); len(evs) != 1 || evs[0].CardID != shredder || evs[0].Actor != me.ID {
		t.Errorf("connive events = %+v, want one for the Shredder under its controller", evs)
	}
}

// --- The Gitrog Monster ------------------------------------------------

func TestGitrogMonsterUpkeepSacrificesALandWhenOneIsOffered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	gitrog := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Gitrog Monster", TypeLine: "Legendary Creature — Frog Horror",
		OracleID: effectstest.GitrogMonsterOracle, Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	land := effectstest.PushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	if effectstest.TriggerOnStack(g, gitrog) == nil {
		t.Fatalf("upkeep trigger not on the stack")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0) // "Sacrifice a land"
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(land) {
		t.Errorf("the land was not sacrificed")
	}
	if !g.Battlefield.Contains(gitrog) {
		t.Errorf("The Gitrog Monster was sacrificed even though a land was offered")
	}
}

func TestGitrogMonsterUpkeepSacrificesItselfWithNoLandToOffer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	gitrog := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Gitrog Monster", TypeLine: "Legendary Creature — Frog Horror",
		OracleID: effectstest.GitrogMonsterOracle, Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(gitrog) {
		t.Errorf("The Gitrog Monster survived its own upkeep with no land to sacrifice")
	}
}

func TestGitrogMonsterDrawsWhenALandGoesToYourGraveyardFromAnywhere(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Gitrog Monster", TypeLine: "Legendary Creature — Frog Horror",
		OracleID: effectstest.GitrogMonsterOracle, Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	landInHand := effectstest.HandCardForTest(me, "Some Land", "Land", "")
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
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before-1 {
		t.Errorf("hand size after the discard resolved = %d, want %d", got, before-1)
	}
}

// --- Taurean Mauler ------------------------------------------------

func TestTaureanMaulerMayGetACounterWhenAnOpponentCastsASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	mauler := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Taurean Mauler", TypeLine: "Creature — Shapeshifter",
		OracleID: effectstest.TaureanMaulerOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	spell := effectstest.HandCardForTest(opp, "Their Spell", "Sorcery", "")
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	g.Turn.ActiveSeat = (seat + 1) % len(g.Seats)
	if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, mauler)
	if got := c.CurrentPower(); got != 3 {
		t.Errorf("Taurean Mauler power = %d, want 3 after accepting the counter", got)
	}
}

// --- Crashing Drawbridge ------------------------------------------------

func TestCrashingDrawbridgeGrantsHasteToCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bridge := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Crashing Drawbridge", TypeLine: "Artifact Creature — Wall",
		OracleID: effectstest.CrashingDrawbridgeOra, Power: 0, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.PushDiesCreatureForTest(g, me.ID, "Sick Bear", "", "Creature — Bear", 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Fatalf("bear already has haste before the ability resolves")
	}
	if err := g.ActivateCatalogAbility(me.ID, bridge, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Crashing Drawbridge: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Errorf("bear lacks haste after Crashing Drawbridge resolved")
	}
}

// --- Mirror Entity ------------------------------------------------------

func TestMirrorEntitySetsBasePowerToughnessAndGrantsAllCreatureTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	entity := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mirror Entity", TypeLine: "Creature — Shapeshifter",
		OracleID: effectstest.MirrorEntityOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.PushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, entity, 0, game.ActivateAbilityParams{XValue: 4}); err != nil {
		t.Fatalf("activate Mirror Entity: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tg != 4 {
		t.Errorf("bear P/T = %d/%d, want 4/4", p, tg)
	}
	var hasElf bool
	var gotSubtypes []string
	g.ReadSnapshot(func() {
		c, _ := effectstest.BattlefieldCard(g, bear)
		hasElf = c.HasSubtype("Elf")
		gotSubtypes = c.Effective().Subtypes
	})
	if !hasElf {
		t.Errorf("bear subtypes = %v, want every creature type (e.g. Elf)", gotSubtypes)
	}
}

// --- Blightsteel Colossus ------------------------------------------------

func TestBlightsteelColossusIsShuffledIntoItsLibraryInsteadOfDying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	colossus := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blightsteel Colossus", TypeLine: "Artifact Creature — Phyrexian Golem",
		OracleID: effectstest.BlightsteelOracle, Power: 11, Toughness: 11, Owner: me.ID, Controller: me.ID,
	})
	// Indestructible refuses an ordinary destroy, so the battlefield
	// exit this pins is a sacrifice — sacrifice ignores indestructible
	// (CR 701.21a) and still opens the same "put into a graveyard"
	// replacement window.
	var applyErr error
	g.WithWriteLock(func() {
		applyErr = (effects.SacrificePermanent{Target: colossus}).Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID}))
	})
	if applyErr != nil {
		t.Fatalf("SacrificePermanent: %v", applyErr)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Elesh Norn, Mother of Machines", TypeLine: "Legendary Creature — Phyrexian Praetor",
		OracleID: effectstest.EleshNornOracle, Power: 4, Toughness: 7, Owner: me.ID, Controller: me.ID,
	})
	before := me.Hand.Size()
	effectstest.CastWithCost(t, g, "Mulldrifter", "Creature — Elemental", "{4}{U}", "24d0f5e7-0d9e-4b76-900e-a7274e80312d")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+4 {
		t.Errorf("hand size after Mulldrifter's doubled ETB = %d, want %d (draw two, twice)", got, before+4)
	}
}

// --- Heliod, Sun-Crowned ------------------------------------------------

func TestHeliodIsNotACreatureBelowDevotionFiveAndBecomesOneAtFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	heliod := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Heliod, Sun-Crowned", TypeLine: "Legendary Enchantment Creature — God",
		OracleID: effectstest.HeliodSunCrownedOra, ManaCost: "{2}{W}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	if got := effectstest.EffectiveTypes(t, g, heliod); effectstest.ContainsString(got, "Creature") {
		t.Fatalf("Heliod is a creature at devotion 1: %v", got)
	}
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", ManaCost: "{W}{W}{W}{W}",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	if got := effectstest.EffectiveTypes(t, g, heliod); !effectstest.ContainsString(got, "Creature") {
		t.Fatalf("Heliod is not a creature at devotion 5: %v", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, heliod, "indestructible") {
		t.Error("Heliod lacks indestructible")
	}
}

func TestHeliodPutsACounterOnTargetWhenYouGainLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Heliod, Sun-Crowned", TypeLine: "Legendary Enchantment Creature — God",
		OracleID: effectstest.HeliodSunCrownedOra, ManaCost: "{2}{W}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.PushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatalf("no pick_target prompt for Heliod's lifegain trigger")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, bear)
	if got := c.CurrentPower(); got != 3 {
		t.Errorf("bear power = %d, want 3 after a +1/+1 counter", got)
	}
}

func TestHeliodActivatedAbilityGrantsLifelinkToAnotherTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	heliod := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Heliod, Sun-Crowned", TypeLine: "Legendary Enchantment Creature — God",
		OracleID: effectstest.HeliodSunCrownedOra, ManaCost: "{2}{W}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.PushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Fatalf("bear already has lifelink")
	}
	if err := g.ActivateCatalogAbility(me.ID, heliod, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate Heliod: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Errorf("bear lacks lifelink after Heliod's ability resolved")
	}
}

// --- Apex Devastator ------------------------------------------------

func TestApexDevastatorCascadesFourTimes(t *testing.T) {
	if got := len(game.CatalogTriggers(effectstest.ApexDevastatorOracle)); got != 4 {
		t.Fatalf("Apex Devastator has %d triggered abilities, want 4 (cascade x4)", got)
	}

	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedCheapLibrary(g, me, 8)

	effectstest.CastWithCost(t, g, "Apex Devastator", "Creature — Chimera Hydra", "{8}{G}{G}", effectstest.ApexDevastatorOracle)
	effectstest.PassPriorityAroundTable(t, g)

	// Each of the four cascade triggers resolves one at a time with a
	// priority pass between them, so drain-then-pass four times.
	total := 0
	for i := 0; i < 4; i++ {
		total += effectstest.AnswerAllMayCast(t, g, me.ID, false)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if total != 4 {
		t.Errorf("Apex Devastator made %d cascade offers, want 4", total)
	}
}
