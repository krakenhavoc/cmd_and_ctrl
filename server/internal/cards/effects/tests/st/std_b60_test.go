package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Rooftop Percher ----------------------------------------------------------

func TestStdB60PercherGainsThreeEvenWithNothingToExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	effectstest.CastHoldCreature(t, g, "Rooftop Percher", "Creature — Shapeshifter", stdB60PercherOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
}

func TestStdB60PercherExilesAGraveyardCardAndGainsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushCatalogGraveyardCard(opp, "Their Card", "Creature — Bear", "", 2, 2)
	life := me.Life
	effectstest.CastHoldCreature(t, g, "Rooftop Percher", "Creature — Shapeshifter", stdB60PercherOracle, 3, 3)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, theirs)
	for i := 0; i < 3 && effectstest.LatestPickTarget(g, me.ID) != nil; i++ {
		p := effectstest.LatestPickTarget(g, me.ID)
		_ = g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{}) // decline the second "up to"
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(theirs) {
		t.Error("the targeted card stayed in the graveyard")
	}
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
}

// --- Ajani, Outland Chaperone -------------------------------------------------

func TestStdB60AjaniPlusOneMakesAKithkin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	ajani := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Ajani, Outland Chaperone", "Legendary Planeswalker — Ajani", stdB60AjaniOracle, 3)
	effectstest.B16Activate(t, g, me.ID, ajani, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.MshB1Tokens(g, me.ID, "Kithkin"); got != 1 {
		t.Fatalf("%d Kithkin, want 1", got)
	}
	k := effectstest.MshB1Token(t, g, me.ID, "Kithkin")
	if k.Power != 1 || k.Toughness != 1 || len(k.Colors) != 2 {
		t.Errorf("Kithkin = %d/%d %v, want 1/1 green and white", k.Power, k.Toughness, k.Colors)
	}
	if n := effectstest.LoyaltyCount(g, ajani); n != 4 {
		t.Errorf("loyalty %d, want 4", n)
	}
}

func TestStdB60AjaniMinusTwoDealsFourToATappedCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	ajani := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Ajani, Outland Chaperone", "Legendary Planeswalker — Ajani", stdB60AjaniOracle, 3)
	untapped := effectstest.B12Creature(g, opp.ID, "Untapped", "Creature — Bear", 4, 4)
	tapped := effectstest.B12Creature(g, opp.ID, "Tapped", "Creature — Bear", 4, 4)
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
	effectstest.B16Activate(t, g, me.ID, ajani, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: tapped}}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped 4/4 survived 4 damage")
	}
	if n := effectstest.LoyaltyCount(g, ajani); n != 1 {
		t.Errorf("loyalty %d, want 1", n)
	}
}

func TestStdB60AjaniUltimateLooksAtLifeTotalCardsAndPutsCheapPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	g.WithWriteLock(func() { me.Life = 3 })
	ajani := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Ajani, Outland Chaperone", "Legendary Planeswalker — Ajani", stdB60AjaniOracle, 8)
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
	effectstest.B16Activate(t, g, me.ID, ajani, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	bark := effectstest.PushCatalogPermanent(g, me.ID, "Bark of Doran", "Artifact — Equipment", stdB60BarkOracle, false)
	wall := effectstest.PushVanillaCreature(g, me.ID, "Wall", 0, 4)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}"); err != nil {
		t.Fatal(err)
	}
	effectstest.EquipTo(t, g, me.ID, bark, wall)
	w := effectstest.GaLayered(t, g, wall)
	if w.CurrentPower() != 0 || w.CurrentToughness() != 5 {
		t.Errorf("equipped wall = %d/%d, want 0/5", w.CurrentPower(), w.CurrentToughness())
	}
	before := opp.Life
	effectstest.AttackWith(t, g, opp.ID, wall)
	effectstest.PassPriorityAroundTable(t, g)
	if want := before - 5; opp.Life != want {
		t.Errorf("opponent at %d, want %d (the equipped 0/5 deals 5)", opp.Life, want)
	}
}

func TestStdB60BarkDoesNotChangeAnUnequippedCreatureOrOneWithPowerAtLeastToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	bark := effectstest.PushCatalogPermanent(g, me.ID, "Bark of Doran", "Artifact — Equipment", stdB60BarkOracle, false)
	plain := effectstest.PushVanillaCreature(g, me.ID, "Plain Bear", 3, 3)
	big := effectstest.PushVanillaCreature(g, me.ID, "Big Bear", 4, 1)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}"); err != nil {
		t.Fatal(err)
	}
	effectstest.EquipTo(t, g, me.ID, bark, big)
	before := opp.Life
	effectstest.AttackWith(t, g, opp.ID, plain, big)
	effectstest.PassPriorityAroundTable(t, g)
	// 4/1 equipped is 4/2: power is still greater, so it deals 4.
	if want := before - (3 + 4); opp.Life != want {
		t.Errorf("opponent at %d, want %d", opp.Life, want)
	}
}

// --- Curious Colossus ---------------------------------------------------------

func TestStdB60ColossusShrinksEveryCreatureTheOpponentHasForGood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flier := effectstest.B12Creature(g, opp.ID, "Dragon", "Creature — Dragon", 5, 5)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == flier {
				g.Battlefield.Cards[i].Keywords = []string{"flying"}
			}
		}
	})
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Curious Colossus", "Creature — Giant Warrior", stdB60ColossusOracle, 7, 7)
	for i := 0; i < 4 && effectstest.LatestPickTarget(g, me.ID) != nil; i++ {
		p := effectstest.LatestPickTarget(g, me.ID)
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	d := effectstest.GaLayered(t, g, flier)
	if d.CurrentPower() != 1 || d.CurrentToughness() != 1 {
		t.Errorf("Dragon = %d/%d, want 1/1", d.CurrentPower(), d.CurrentToughness())
	}
	if effectstest.HasKeywordOnBattlefield(t, g, flier, "flying") {
		t.Error("it kept flying")
	}
	if !d.HasSubtype("Coward") || !d.HasSubtype("Dragon") {
		t.Errorf("subtypes of the Dragon: Coward=%v Dragon=%v, want both", d.HasSubtype("Coward"), d.HasSubtype("Dragon"))
	}
	if m := effectstest.GaLayered(t, g, mine); m.CurrentPower() != 2 {
		t.Errorf("my own creature was shrunk (%d/%d)", m.CurrentPower(), m.CurrentToughness())
	}
	// A creature that enters afterwards is not affected.
	late := effectstest.B12Creature(g, opp.ID, "Late Bear", "Creature — Bear", 3, 3)
	if l := effectstest.GaLayered(t, g, late); l.CurrentPower() != 3 {
		t.Errorf("a creature that entered later is %d/%d, want 3/3", l.CurrentPower(), l.CurrentToughness())
	}
}

// --- Keep Out -----------------------------------------------------------------

func TestStdB60KeepOutDealsFourToATappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	tapped := effectstest.B12Creature(g, opp.ID, "Tapped", "Creature — Bear", 4, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == tapped {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	me := g.Seats[0]
	id := effectstest.HandCardForTest(me, "Keep Out", "Instant", stdB60KeepOutOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0}, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: tapped}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped 4/4 survived 4 damage")
	}
}

func TestStdB60KeepOutDestroysAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ench := effectstest.B12Permanent(g, opp.ID, "Charm", "Enchantment")
	me := g.Seats[0]
	id := effectstest.HandCardForTest(me, "Keep Out", "Instant", stdB60KeepOutOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ench}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived")
	}
}

// --- Liminal Hold -------------------------------------------------------------

func TestStdB60LiminalHoldExilesUntilItLeavesAndGainsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := me.Life
	hold := effectstest.CastHoldCreature(t, g, "Liminal Hold", "Enchantment", stdB60HoldOracle, 0, 0)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the permanent was not exiled")
	}
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(hold) })
	g.RunStateChecksForTest()
	if effectstest.TmtB1Count(g, opp.ID, "Their Bear") != 1 {
		t.Error("the permanent did not return when Liminal Hold left")
	}
}

// --- Kithkeeper ---------------------------------------------------------------

func TestStdB60KithkeeperMakesOneKithkinPerColorAmongYourPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	red := game.NewCard("Red Thing", me.ID)
	red.TypeLine, red.Colors = "Creature — Goblin", []string{"R"}
	red.Power, red.Toughness = 1, 1
	effectstest.PushBattlefieldCardWithTimestamp(g, red)
	keeper := game.NewCard("Kithkeeper", me.ID)
	keeper.TypeLine, keeper.OracleID, keeper.Colors = "Creature — Elemental", stdB60KithkeeperOracle, []string{"W"}
	keeper.Power, keeper.Toughness = 3, 3
	me.Hand.PushTop(keeper)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, keeper.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The Kithkeeper itself is white; the Goblin is red: two colors.
	if got := effectstest.MshB1Tokens(g, me.ID, "Kithkin"); got != 2 {
		t.Errorf("%d Kithkin, want 2 (white and red)", got)
	}
}

func TestStdB60KithkeeperPumpsAndFliesForThreeTappedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	keeper := effectstest.PushCatalogPermanent(g, me.ID, "Kithkeeper", "Creature — Elemental", stdB60KithkeeperOracle, true)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	c := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 1, 1)
	effectstest.B16Activate(t, g, me.ID, keeper, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{a, b, c}})
	effectstest.PassPriorityAroundTable(t, g)
	k := effectstest.GaLayered(t, g, keeper)
	if k.CurrentPower() != 4 {
		t.Errorf("power %d, want 4 (the test permanent is a 1-power card, +3)", k.CurrentPower())
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, keeper, "flying") {
		t.Error("no flying")
	}
	for _, id := range []uuid.UUID{a, b, c} {
		if !effectstest.B16Tapped(t, g, id) {
			t.Error("a creature was not tapped to pay")
		}
	}
}

// --- Kinscaer Sentry ----------------------------------------------------------

func TestStdB60SentryPutsACreatureWithValueAtMostTheAttackersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sentry := effectstest.PushCatalogPermanent(g, me.ID, "Kinscaer Sentry", "Creature — Kithkin Soldier", stdB60SentryOracle, false)
	cheap := game.NewCard("Cheap", me.ID)
	cheap.TypeLine, cheap.ManaCost, cheap.Power, cheap.Toughness = "Creature — Bear", "{1}", 1, 1
	me.Hand.PushTop(cheap)
	dear := game.NewCard("Dear", me.ID)
	dear.TypeLine, dear.ManaCost, dear.Power, dear.Toughness = "Creature — Giant", "{3}{G}", 4, 4
	me.Hand.PushTop(dear)

	effectstest.DeclareAttack(t, g, opp.ID, sentry)
	effectstest.LockInAttacks(t, g)
	pc := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pc == nil {
		effectstest.PassPriorityAroundTable(t, g)
		pc = effectstest.ChooseCardsChoiceFor(g, me.ID)
	}
	if pc == nil {
		t.Fatal("no choose-cards prompt")
	}
	for _, id := range pc.ChooseCards {
		if id == dear.InstanceID {
			t.Error("a mana value 4 card was offered with one attacker")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, cheap.InstanceID)
	effectstest.AnswerOptionPick(t, g, me.ID, 0) // the Sentry's own defender comes first
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(cheap.InstanceID)
	if !ok || !c.Tapped || c.AttackingTarget == uuid.Nil {
		t.Errorf("the put creature: found=%v tapped=%v attacking=%v, want tapped and attacking", ok, c.Tapped, c.AttackingTarget)
	}
}

// --- Meanders Guide -----------------------------------------------------------

func TestStdB60MeandersGuideTapsAMerfolkAndReturnsACheapCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guide := effectstest.PushCatalogPermanent(g, me.ID, "Meanders Guide", "Creature — Merfolk Scout", stdB60GuideOracle, false)
	friend := effectstest.B12Creature(g, me.ID, "Merfolk Friend", "Creature — Merfolk", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Not Merfolk", "Creature — Bear", 1, 1)
	cheap := effectstest.PushCatalogGraveyardCard(me, "Cheap Dead", "Creature — Bear", "", 1, 1)
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == cheap {
				me.Graveyard.Cards[i].ManaCost = "{2}"
			}
		}
	})
	effectstest.DeclareAttack(t, g, opp.ID, guide)
	effectstest.LockInAttacks(t, g)
	pc := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pc == nil {
		effectstest.PassPriorityAroundTable(t, g)
		pc = effectstest.ChooseCardsChoiceFor(g, me.ID)
	}
	if pc == nil {
		t.Fatal("no choose-cards prompt")
	}
	for _, id := range pc.ChooseCards {
		if id == bear || id == guide {
			t.Error("a non-Merfolk (or Guide itself) was offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, friend)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, friend) {
		t.Error("the Merfolk was not tapped")
	}
	if !g.Battlefield.Contains(cheap) {
		t.Error("the creature card did not return")
	}
}

func TestStdB60MeandersGuideDecliningReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guide := effectstest.PushCatalogPermanent(g, me.ID, "Meanders Guide", "Creature — Merfolk Scout", stdB60GuideOracle, false)
	effectstest.B12Creature(g, me.ID, "Merfolk Friend", "Creature — Merfolk", 1, 1)
	cheap := effectstest.PushCatalogGraveyardCard(me, "Cheap Dead", "Creature — Bear", "", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, guide)
	effectstest.LockInAttacks(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) {
		t.Error("a creature returned although no Merfolk was tapped")
	}
}

// --- Eirdu, Carrier of Dawn // Isilu, Carrier of Twilight -----------------------

func TestStdB60IsiluGivesPersistToOtherNontokenCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	isilu := effectstest.PushCatalogPermanent(g, me.ID, "Isilu, Carrier of Twilight", "Legendary Creature — Elemental God", stdB60EirduOracle+"#1", false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.RunStateChecksForTest()
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "persist") {
		t.Error("another nontoken creature I control lacks persist")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, isilu, "persist") {
		t.Error("Isilu itself has persist (\"each other\")")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, theirs, "persist") {
		t.Error("an opponent's creature has persist")
	}
}

func TestStdB60EirduGrantsConvokeToCreatureSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Eirdu, Carrier of Dawn", "Legendary Creature — Elemental God", stdB60EirduOracle, false)
	elves := effectstest.CgTwoGreenCreatures(g, me.ID)
	bear := effectstest.CgBear(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(bear) })
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true, TapIDs: elves}); err != nil {
		t.Fatalf("convoking a creature spell under Eirdu: %v", err)
	}
	if !effectstest.CgTapped(g, elves) {
		t.Error("the convoking creatures were not tapped")
	}
	// An instant gets no convoke.
	more := effectstest.CgTwoGreenCreatures(g, me.ID)
	bolt := game.Card{InstanceID: uuid.New(), Name: "Bolt", TypeLine: "Instant", ManaCost: "{1}{G}",
		Colors: []string{"G"}, Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() { me.Hand.PushTop(bolt) })
	if err := g.CastSpell(me.ID, bolt.InstanceID, game.CastSpellParams{Strict: true, TapIDs: more}); err == nil {
		t.Error("an instant was convoked under Eirdu")
	}
}
