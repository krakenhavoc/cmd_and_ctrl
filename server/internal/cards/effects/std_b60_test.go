package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b60_test.go — Standard leftovers, batch 60 (#4452): Lorwyn Eclipsed.

const (
	stdB60PercherOracle    = "3dc73517-0767-46ae-bc8d-ee378237a509"
	stdB60AjaniOracle      = "3ba809ce-3556-415c-9fcc-291cb10a43cd"
	stdB60BarkOracle       = "db3ba239-2239-4e9d-ad70-cff32ed24578"
	stdB60ColossusOracle   = "c8de2e1b-f2fa-4b6a-a3c5-9811c7f5f051"
	stdB60EirduOracle      = "36af3c2c-49d7-46ea-ab02-c254b332448e"
	stdB60KeepOutOracle    = "61cb927a-4cf8-49d6-ab4b-9e882bc08774"
	stdB60SentryOracle     = "fdf0fba8-f335-419c-b96b-e99037727c43"
	stdB60KithkeeperOracle = "5c5766f9-befd-4426-9527-d50af0699ca1"
	stdB60HoldOracle       = "101407b4-e7e8-4aff-84a2-a97947358b88"
	stdB60GuideOracle      = "8229954b-e1ef-4503-8443-2356e2a87663"
)

func TestStdB60EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		stdB60PercherOracle:      "Rooftop Percher",
		stdB60AjaniOracle:        "Ajani, Outland Chaperone",
		stdB60BarkOracle:         "Bark of Doran",
		stdB60ColossusOracle:     "Curious Colossus",
		stdB60EirduOracle:        "Eirdu, Carrier of Dawn",
		stdB60EirduOracle + "#1": "Isilu, Carrier of Twilight",
		stdB60KeepOutOracle:      "Keep Out",
		stdB60SentryOracle:       "Kinscaer Sentry",
		stdB60KithkeeperOracle:   "Kithkeeper",
		stdB60HoldOracle:         "Liminal Hold",
		stdB60GuideOracle:        "Meanders Guide",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Rooftop Percher ----------------------------------------------------------

func TestStdB60PercherGainsThreeEvenWithNothingToExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	castHoldCreature(t, g, "Rooftop Percher", "Creature — Shapeshifter", stdB60PercherOracle, 3, 3)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
}

func TestStdB60PercherExilesAGraveyardCardAndGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushCatalogGraveyardCard(opp, "Their Card", "Creature — Bear", "", 2, 2)
	life := me.Life
	castHoldCreature(t, g, "Rooftop Percher", "Creature — Shapeshifter", stdB60PercherOracle, 3, 3)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirs)
	for i := 0; i < 3 && latestPickTarget(g, me.ID) != nil; i++ {
		p := latestPickTarget(g, me.ID)
		_ = g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{}) // decline the second "up to"
	}
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(theirs) {
		t.Error("the targeted card stayed in the graveyard")
	}
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
}

// --- Ajani, Outland Chaperone -------------------------------------------------

func TestStdB60AjaniPlusOneMakesAKithkin(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	ajani := pushCatalogWalkerWithLine(g, me.ID, "Ajani, Outland Chaperone", "Legendary Planeswalker — Ajani", stdB60AjaniOracle, 3)
	b16Activate(t, g, me.ID, ajani, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := mshB1Tokens(g, me.ID, "Kithkin"); got != 1 {
		t.Fatalf("%d Kithkin, want 1", got)
	}
	k := mshB1Token(t, g, me.ID, "Kithkin")
	if k.Power != 1 || k.Toughness != 1 || len(k.Colors) != 2 {
		t.Errorf("Kithkin = %d/%d %v, want 1/1 green and white", k.Power, k.Toughness, k.Colors)
	}
	if n := loyaltyCount(g, ajani); n != 4 {
		t.Errorf("loyalty %d, want 4", n)
	}
}

func TestStdB60AjaniMinusTwoDealsFourToATappedCreatureOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMain(t, g)
	ajani := pushCatalogWalkerWithLine(g, me.ID, "Ajani, Outland Chaperone", "Legendary Planeswalker — Ajani", stdB60AjaniOracle, 3)
	untapped := b12Creature(g, opp.ID, "Untapped", "Creature — Bear", 4, 4)
	tapped := b12Creature(g, opp.ID, "Tapped", "Creature — Bear", 4, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == tapped {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, ajani, 1, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: untapped}}}); err == nil {
		t.Fatal("an untapped creature was accepted as the target")
	}
	b16Activate(t, g, me.ID, ajani, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: tapped}}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped 4/4 survived 4 damage")
	}
	if n := loyaltyCount(g, ajani); n != 1 {
		t.Errorf("loyalty %d, want 1", n)
	}
}

func TestStdB60AjaniUltimateLooksAtLifeTotalCardsAndPutsCheapPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	g.WithWriteLock(func() { me.Life = 3 })
	ajani := pushCatalogWalkerWithLine(g, me.ID, "Ajani, Outland Chaperone", "Legendary Planeswalker — Ajani", stdB60AjaniOracle, 8)
	mk := func(name, typeLine, cost string) uuid.UUID {
		c := game.NewCard(name, me.ID)
		c.TypeLine, c.ManaCost = typeLine, cost
		me.Library.PushTop(c)
		return c.InstanceID
	}
	deep := mk("Deep Cheap", "Creature — Bear", "{1}")      // below the top three
	dear := mk("Too Dear", "Creature — Giant", "{3}{G}{G}") // mana value 5
	land := mk("Forest", "Basic Land — Forest", "")
	cheap := mk("Cheap Bear", "Creature — Bear", "{1}{G}")
	b16Activate(t, g, me.ID, ajani, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Error("the cheap permanent was not put onto the battlefield")
	}
	for name, id := range map[string]uuid.UUID{"too dear": dear, "land": land, "deeper than X": deep} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s card entered", name)
		}
	}
}

// --- Bark of Doran ------------------------------------------------------------

func TestStdB60BarkGivesPlusZeroPlusOneAndDamageByToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMain(t, g)
	bark := pushCatalogPermanent(g, me.ID, "Bark of Doran", "Artifact — Equipment", stdB60BarkOracle, false)
	wall := pushVanillaCreature(g, me.ID, "Wall", 0, 4)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}"); err != nil {
		t.Fatal(err)
	}
	equipTo(t, g, me.ID, bark, wall)
	w := gaLayered(t, g, wall)
	if w.CurrentPower() != 0 || w.CurrentToughness() != 5 {
		t.Errorf("equipped wall = %d/%d, want 0/5", w.CurrentPower(), w.CurrentToughness())
	}
	before := opp.Life
	attackWith(t, g, opp.ID, wall)
	passPriorityAroundTable(t, g)
	if want := before - 5; opp.Life != want {
		t.Errorf("opponent at %d, want %d (the equipped 0/5 deals 5)", opp.Life, want)
	}
}

func TestStdB60BarkDoesNotChangeAnUnequippedCreatureOrOneWithPowerAtLeastToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMain(t, g)
	bark := pushCatalogPermanent(g, me.ID, "Bark of Doran", "Artifact — Equipment", stdB60BarkOracle, false)
	plain := pushVanillaCreature(g, me.ID, "Plain Bear", 3, 3)
	big := pushVanillaCreature(g, me.ID, "Big Bear", 4, 1)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}"); err != nil {
		t.Fatal(err)
	}
	equipTo(t, g, me.ID, bark, big)
	before := opp.Life
	attackWith(t, g, opp.ID, plain, big)
	passPriorityAroundTable(t, g)
	// 4/1 equipped is 4/2: power is still greater, so it deals 4.
	if want := before - (3 + 4); opp.Life != want {
		t.Errorf("opponent at %d, want %d", opp.Life, want)
	}
}

// --- Curious Colossus ---------------------------------------------------------

func TestStdB60ColossusShrinksEveryCreatureTheOpponentHasForGood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flier := b12Creature(g, opp.ID, "Dragon", "Creature — Dragon", 5, 5)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == flier {
				g.Battlefield.Cards[i].Keywords = []string{"flying"}
			}
		}
	})
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Curious Colossus", "Creature — Giant Warrior", stdB60ColossusOracle, 7, 7)
	for i := 0; i < 4 && latestPickTarget(g, me.ID) != nil; i++ {
		p := latestPickTarget(g, me.ID)
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatal(err)
		}
	}
	passPriorityAroundTable(t, g)
	d := gaLayered(t, g, flier)
	if d.CurrentPower() != 1 || d.CurrentToughness() != 1 {
		t.Errorf("Dragon = %d/%d, want 1/1", d.CurrentPower(), d.CurrentToughness())
	}
	if hasKeywordOnBattlefield(t, g, flier, "flying") {
		t.Error("it kept flying")
	}
	if !d.HasSubtype("Coward") || !d.HasSubtype("Dragon") {
		t.Errorf("subtypes of the Dragon: Coward=%v Dragon=%v, want both", d.HasSubtype("Coward"), d.HasSubtype("Dragon"))
	}
	if m := gaLayered(t, g, mine); m.CurrentPower() != 2 {
		t.Errorf("my own creature was shrunk (%d/%d)", m.CurrentPower(), m.CurrentToughness())
	}
	// A creature that enters afterwards is not affected.
	late := b12Creature(g, opp.ID, "Late Bear", "Creature — Bear", 3, 3)
	if l := gaLayered(t, g, late); l.CurrentPower() != 3 {
		t.Errorf("a creature that entered later is %d/%d, want 3/3", l.CurrentPower(), l.CurrentToughness())
	}
}

// --- Keep Out -----------------------------------------------------------------

func TestStdB60KeepOutDealsFourToATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	tapped := b12Creature(g, opp.ID, "Tapped", "Creature — Bear", 4, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == tapped {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	me := g.Seats[0]
	id := handCardForTest(me, "Keep Out", "Instant", stdB60KeepOutOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0}, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: tapped}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped 4/4 survived 4 damage")
	}
}

func TestStdB60KeepOutDestroysAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	ench := b12Permanent(g, opp.ID, "Charm", "Enchantment")
	me := g.Seats[0]
	id := handCardForTest(me, "Keep Out", "Instant", stdB60KeepOutOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ench}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived")
	}
}

// --- Liminal Hold -------------------------------------------------------------

func TestStdB60LiminalHoldExilesUntilItLeavesAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := me.Life
	hold := castHoldCreature(t, g, "Liminal Hold", "Enchantment", stdB60HoldOracle, 0, 0)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the permanent was not exiled")
	}
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(hold) })
	g.RunStateChecksForTest()
	if tmtB1Count(g, opp.ID, "Their Bear") != 1 {
		t.Error("the permanent did not return when Liminal Hold left")
	}
}

// --- Kithkeeper ---------------------------------------------------------------

func TestStdB60KithkeeperMakesOneKithkinPerColorAmongYourPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	red := game.NewCard("Red Thing", me.ID)
	red.TypeLine, red.Colors = "Creature — Goblin", []string{"R"}
	red.Power, red.Toughness = 1, 1
	pushBattlefieldCardWithTimestamp(g, red)
	keeper := game.NewCard("Kithkeeper", me.ID)
	keeper.TypeLine, keeper.OracleID, keeper.Colors = "Creature — Elemental", stdB60KithkeeperOracle, []string{"W"}
	keeper.Power, keeper.Toughness = 3, 3
	me.Hand.PushTop(keeper)
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, keeper.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	// The Kithkeeper itself is white; the Goblin is red: two colors.
	if got := mshB1Tokens(g, me.ID, "Kithkin"); got != 2 {
		t.Errorf("%d Kithkin, want 2 (white and red)", got)
	}
}

func TestStdB60KithkeeperPumpsAndFliesForThreeTappedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	keeper := pushCatalogPermanent(g, me.ID, "Kithkeeper", "Creature — Elemental", stdB60KithkeeperOracle, true)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	c := b12Creature(g, me.ID, "C", "Creature — Bear", 1, 1)
	b16Activate(t, g, me.ID, keeper, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{a, b, c}})
	passPriorityAroundTable(t, g)
	k := gaLayered(t, g, keeper)
	if k.CurrentPower() != 4 {
		t.Errorf("power %d, want 4 (the test permanent is a 1-power card, +3)", k.CurrentPower())
	}
	if !hasKeywordOnBattlefield(t, g, keeper, "flying") {
		t.Error("no flying")
	}
	for _, id := range []uuid.UUID{a, b, c} {
		if !b16Tapped(t, g, id) {
			t.Error("a creature was not tapped to pay")
		}
	}
}

// --- Kinscaer Sentry ----------------------------------------------------------

func TestStdB60SentryPutsACreatureWithValueAtMostTheAttackersTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sentry := pushCatalogPermanent(g, me.ID, "Kinscaer Sentry", "Creature — Kithkin Soldier", stdB60SentryOracle, false)
	cheap := game.NewCard("Cheap", me.ID)
	cheap.TypeLine, cheap.ManaCost, cheap.Power, cheap.Toughness = "Creature — Bear", "{1}", 1, 1
	me.Hand.PushTop(cheap)
	dear := game.NewCard("Dear", me.ID)
	dear.TypeLine, dear.ManaCost, dear.Power, dear.Toughness = "Creature — Giant", "{3}{G}", 4, 4
	me.Hand.PushTop(dear)

	declareAttack(t, g, opp.ID, sentry)
	lockInAttacks(t, g)
	pc := chooseCardsChoiceFor(g, me.ID)
	if pc == nil {
		passPriorityAroundTable(t, g)
		pc = chooseCardsChoiceFor(g, me.ID)
	}
	if pc == nil {
		t.Fatal("no choose-cards prompt")
	}
	for _, id := range pc.ChooseCards {
		if id == dear.InstanceID {
			t.Error("a mana value 4 card was offered with one attacker")
		}
	}
	answerChooseCards(t, g, me.ID, cheap.InstanceID)
	answerOptionPick(t, g, me.ID, 0) // the Sentry's own defender comes first
	passPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(cheap.InstanceID)
	if !ok || !c.Tapped || c.AttackingTarget == uuid.Nil {
		t.Errorf("the put creature: found=%v tapped=%v attacking=%v, want tapped and attacking", ok, c.Tapped, c.AttackingTarget)
	}
}

// --- Meanders Guide -----------------------------------------------------------

func TestStdB60MeandersGuideTapsAMerfolkAndReturnsACheapCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guide := pushCatalogPermanent(g, me.ID, "Meanders Guide", "Creature — Merfolk Scout", stdB60GuideOracle, false)
	friend := b12Creature(g, me.ID, "Merfolk Friend", "Creature — Merfolk", 1, 1)
	bear := b12Creature(g, me.ID, "Not Merfolk", "Creature — Bear", 1, 1)
	cheap := pushCatalogGraveyardCard(me, "Cheap Dead", "Creature — Bear", "", 1, 1)
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == cheap {
				me.Graveyard.Cards[i].ManaCost = "{2}"
			}
		}
	})
	declareAttack(t, g, opp.ID, guide)
	lockInAttacks(t, g)
	pc := chooseCardsChoiceFor(g, me.ID)
	if pc == nil {
		passPriorityAroundTable(t, g)
		pc = chooseCardsChoiceFor(g, me.ID)
	}
	if pc == nil {
		t.Fatal("no choose-cards prompt")
	}
	for _, id := range pc.ChooseCards {
		if id == bear || id == guide {
			t.Error("a non-Merfolk (or Guide itself) was offered")
		}
	}
	answerChooseCards(t, g, me.ID, friend)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, friend) {
		t.Error("the Merfolk was not tapped")
	}
	if !g.Battlefield.Contains(cheap) {
		t.Error("the creature card did not return")
	}
}

func TestStdB60MeandersGuideDecliningReturnsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guide := pushCatalogPermanent(g, me.ID, "Meanders Guide", "Creature — Merfolk Scout", stdB60GuideOracle, false)
	b12Creature(g, me.ID, "Merfolk Friend", "Creature — Merfolk", 1, 1)
	cheap := pushCatalogGraveyardCard(me, "Cheap Dead", "Creature — Bear", "", 1, 1)
	declareAttack(t, g, opp.ID, guide)
	lockInAttacks(t, g)
	if chooseCardsChoiceFor(g, me.ID) == nil {
		passPriorityAroundTable(t, g)
	}
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) {
		t.Error("a creature returned although no Merfolk was tapped")
	}
}

// --- Eirdu, Carrier of Dawn // Isilu, Carrier of Twilight -----------------------

func TestStdB60IsiluGivesPersistToOtherNontokenCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	isilu := pushCatalogPermanent(g, me.ID, "Isilu, Carrier of Twilight", "Legendary Creature — Elemental God", stdB60EirduOracle+"#1", false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.RunStateChecksForTest()
	if !hasKeywordOnBattlefield(t, g, bear, "persist") {
		t.Error("another nontoken creature I control lacks persist")
	}
	if hasKeywordOnBattlefield(t, g, isilu, "persist") {
		t.Error("Isilu itself has persist (\"each other\")")
	}
	if hasKeywordOnBattlefield(t, g, theirs, "persist") {
		t.Error("an opponent's creature has persist")
	}
}

func TestStdB60EirduGrantsConvokeToCreatureSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Eirdu, Carrier of Dawn", "Legendary Creature — Elemental God", stdB60EirduOracle, false)
	elves := cgTwoGreenCreatures(g, me.ID)
	bear := cgBear(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(bear) })
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true, TapIDs: elves}); err != nil {
		t.Fatalf("convoking a creature spell under Eirdu: %v", err)
	}
	if !cgTapped(g, elves) {
		t.Error("the convoking creatures were not tapped")
	}
	// An instant gets no convoke.
	more := cgTwoGreenCreatures(g, me.ID)
	bolt := game.Card{InstanceID: uuid.New(), Name: "Bolt", TypeLine: "Instant", ManaCost: "{1}{G}",
		Colors: []string{"G"}, Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() { me.Hand.PushTop(bolt) })
	if err := g.CastSpell(me.ID, bolt.InstanceID, game.CastSpellParams{Strict: true, TapIDs: more}); err == nil {
		t.Error("an instant was convoked under Eirdu")
	}
}
