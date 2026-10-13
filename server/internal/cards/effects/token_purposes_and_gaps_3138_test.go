package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #3138 part 1: the token-making spells with a printed count declare
// it, so the bot prices them (ADR 0126 §6) and the pre-built decks may
// play them (TestCuratedDeckPurposes).
func TestTokenCountSpellsDeclareTheirPurpose(t *testing.T) {
	for _, c := range []struct {
		oracle string
		name   string
		tokens int
		draws  int
	}{
		{"5b2364d7-a811-4595-a1b4-224c70555ffa", "Raise the Alarm", 2, 0},
		{"7a5ff4d4-27b7-47d4-ba88-970c63c4e3fb", "Increasing Devotion", 5, 0},
		{"58f83528-9110-4895-b5ea-51b90af30a8d", "Seize the Spoils", 1, 2},
		{"7a2e866b-9642-495e-b47f-5b13a24373cc", "Pirate's Pillage", 2, 2},
	} {
		spec, ok := Lookup(c.oracle)
		if !ok {
			t.Fatalf("%s is not registered", c.name)
		}
		if spec.Purpose.Tokens != c.tokens || spec.Purpose.Draws != c.draws {
			t.Errorf("%s declares %+v, want tokens %d draws %d", c.name, spec.Purpose, c.tokens, c.draws)
		}
	}
}

// Flashback doubles Increasing Devotion's tokens, and the offer says
// so: castPurpose prices a flashback cast by the offer's purpose.
func TestIncreasingDevotionFlashbackDeclaresTenTokens(t *testing.T) {
	spec, _ := Lookup("7a5ff4d4-27b7-47d4-ba88-970c63c4e3fb")
	if len(spec.AlternativeCosts) != 1 || spec.AlternativeCosts[0].Key != "flashback" {
		t.Fatalf("alternative costs = %+v", spec.AlternativeCosts)
	}
	if got := spec.AlternativeCosts[0].Purpose.Tokens; got != 10 {
		t.Errorf("flashback declares %d tokens, want 10", got)
	}
}

func TestRithsCharmTokenModeDeclaresThree(t *testing.T) {
	spec, ok := Lookup("9dbccdc5-61f7-4a7d-bbbc-5cab393e24f7")
	if !ok || spec.Modes == nil || len(spec.Modes.Options) != 3 {
		t.Fatalf("Rith's Charm modes = %+v", spec.Modes)
	}
	if got := spec.Modes.Options[1].Purpose.Tokens; got != 3 {
		t.Errorf("the Saproling mode declares %d tokens, want 3", got)
	}
	if !spec.Modes.Options[0].Purpose.IsZero() || !spec.Modes.Options[2].Purpose.IsZero() {
		t.Error("only the token mode declares a purpose")
	}
}

func TestRaiseTheAlarmAndIncreasingDevotionMakeTheDeclaredTokens(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Raise the Alarm", "Instant", "5b2364d7-a811-4595-a1b4-224c70555ffa", nil)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Soldier"); got != 2 {
		t.Errorf("Raise the Alarm made %d Soldiers, want 2", got)
	}
}

// --- Stoneforge Mystic (the caveat is gone) -------------------------

const stoneforgeOracle = "358789f9-7d87-411d-919e-d597da665cbd"

func TestStoneforgeMysticPutsAnEquipmentFromHandOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mystic := pushCatalogPermanent(g, me.ID, "Stoneforge Mystic", "Creature — Kor Artificer", stoneforgeOracle, false)
	sword := handCardForTest(me, "Bonesplitter", "Artifact — Equipment", "")
	rock := handCardForTest(me, "Sol Ring", "Artifact", "")

	if err := g.ActivateCatalogAbility(me.ID, mystic, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for the Equipment")
	}
	if pick.ChooseMin != 0 {
		t.Errorf("min %d, want 0 — \"you MAY put\"", pick.ChooseMin)
	}
	if !hasID(pick.ChooseCards, sword) || hasID(pick.ChooseCards, rock) {
		t.Errorf("candidates %v: want the Equipment and not the plain artifact", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, sword)
	if !g.Battlefield.Contains(sword) || me.Hand.Contains(sword) {
		t.Error("the Equipment did not move from hand to the battlefield")
	}
	if !g.Battlefield.Contains(mystic) {
		t.Fatal("the Mystic left the battlefield")
	}
}

func TestStoneforgeMysticAbilityTapsAndCanBeDeclined(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mystic := pushCatalogPermanent(g, me.ID, "Stoneforge Mystic", "Creature — Kor Artificer", stoneforgeOracle, false)
	sword := handCardForTest(me, "Bonesplitter", "Artifact — Equipment", "")

	if err := g.ActivateCatalogAbility(me.ID, mystic, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, mystic, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a second activation went through: {T} is part of the cost")
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID) // decline
	if g.Battlefield.Contains(sword) {
		t.Error("declining still put the Equipment onto the battlefield")
	}
}

// --- Land Tax (re-checked on resolution) -----------------------------

func TestLandTaxRechecksTheLandCountOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Land Tax", "Enchantment", b03LandTaxOracle, false)
	seedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	for i := 0; i < 3; i++ {
		stapleLibraryCard(me, "Plains", "Basic Land — Plains")
	}

	advanceToUpkeepOf(t, g, 1)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	// In response, I even the count: the trigger now has nothing to do.
	seedLandOnBattlefield(g, me.ID, "Plains", "Basic Land — Plains")
	handBefore := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) != nil {
		t.Error("Land Tax searched although no opponent controlled more lands on resolution")
	}
	if me.Hand.Size() != handBefore {
		t.Errorf("hand changed by %d", me.Hand.Size()-handBefore)
	}
}

// --- Carnage Tyrant and Thrun ------------------------------------------

const (
	carnageTyrantOracle = "8c411f4e-a091-447c-9450-d895b10b4985"
	thrunLastTrollOrcl  = "1149e5ac-554a-41b1-84ae-bac42579c1aa"
)

func TestCarnageTyrantIsUncounterableTrampleHexproof(t *testing.T) {
	g := newCatalogGame(t)
	tyrant := pushCatalogPermanent(g, g.Seats[0].ID, "Carnage Tyrant", "Creature — Dinosaur", carnageTyrantOracle, false)
	abilities := effectiveAbilities(t, g, tyrant)
	for _, kw := range []string{"trample", "hexproof"} {
		if !containsString(abilities, kw) {
			t.Errorf("Carnage Tyrant lacks %q: %v", kw, abilities)
		}
	}
	spec, _ := Lookup(carnageTyrantOracle)
	if !spec.CantBeCountered {
		t.Error("Carnage Tyrant can be countered")
	}
}

func TestThrunTheLastTrollIsUncounterableHexproofAndRegenerates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	thrun := pushCatalogPermanent(g, me.ID, "Thrun, the Last Troll", "Legendary Creature — Troll Shaman", thrunLastTrollOrcl, false)
	if !containsString(effectiveAbilities(t, g, thrun), "hexproof") {
		t.Error("Thrun lacks hexproof")
	}
	spec, _ := Lookup(thrunLastTrollOrcl)
	if !spec.CantBeCountered {
		t.Error("Thrun can be countered")
	}
	if err := g.ActivateCatalogAbility(me.ID, thrun, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The shield saves it from destruction.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(thrun) })
	if !g.Battlefield.Contains(thrun) {
		t.Error("a regeneration shield did not save Thrun from destruction")
	}
}

// --- Primeval Titan -----------------------------------------------------

const primevalTitanOracle = "ae83ef2c-960f-4c5b-97cc-52465c687c18"

func TestPrimevalTitanFetchesUpToTwoLandsTappedOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Karakas", TypeLine: "Legendary Land"},
		game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear"},
		game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"},
	)
	castAndResolveCreature(t, g, "Primeval Titan", "Creature — Giant", primevalTitanOracle)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt on entering")
	}
	if c.SearchMax != 2 {
		t.Errorf("search max %d, want 2", c.SearchMax)
	}
	if searchOptionNamed(g, c, "Grizzly Bears") != uuid.Nil {
		t.Error("a nonland card was offered")
	}
	if searchOptionNamed(g, c, "Karakas") == uuid.Nil {
		t.Error("a nonbasic land was not offered: the clause says land cards")
	}
	answerSearchNamed(t, g, me.ID, "Forest", "Karakas")
	for _, name := range []string{"Forest", "Karakas"} {
		id := findBattlefieldByName(g, name)
		if id == uuid.Nil {
			t.Fatalf("%s did not reach the battlefield", name)
		}
		var tapped bool
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				tapped = c.Tapped
			}
		}
		if !tapped {
			t.Errorf("%s entered untapped", name)
		}
	}
	if findBattlefieldByName(g, "Mountain") != uuid.Nil {
		t.Error("a third land was fetched")
	}
}

func TestPrimevalTitanAttackFetchesAndMayFindNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	titan := pushDiesCreatureForTest(g, me.ID, "Primeval Titan", primevalTitanOracle, "Creature — Giant", 6, 6)

	declareAttack(t, g, opp.ID, titan)
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) == nil {
		t.Fatal("attacking did not open the search")
	}
	answerSearchFailToFind(t, g, me.ID)
	if findBattlefieldByName(g, "Forest") != uuid.Nil {
		t.Error("declining the search still fetched a land")
	}
}

// --- Luminarch Aspirant --------------------------------------------------

const luminarchOracle = "cb9994b9-924b-4e10-9075-9cfbec88f2bf"

func TestLuminarchAspirantPutsACounterOnTheChosenCreatureAtCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	aspirant := pushCatalogPermanent(g, me.ID, "Luminarch Aspirant", "Creature — Human Cleric", luminarchOracle, false)
	bear := seedCreature(g, "Grizzly Bears", me.ID)

	advanceTo(t, g, game.StepBeginCombat)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt at the beginning of combat")
	}
	if !hasID(prompt.PickTargetCards, aspirant) || !hasID(prompt.PickTargetCards, bear) {
		t.Errorf("legal targets %v: the Aspirant itself and the Bear are both creatures I control", prompt.PickTargetCards)
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("Bear has %d +1/+1 counters, want 1", got)
	}
	if got := counterCount(g, aspirant, game.CounterPlusOne); got != 0 {
		t.Errorf("Aspirant has %d counters, want 0", got)
	}
}

// --- Hero of Bladehold ---------------------------------------------------

const heroOfBladeholdOracle = "454a8902-8120-4373-96ee-bbf352b04e8d"

func TestHeroOfBladeholdMakesTwoAttackingSoldiersAndBattleCries(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hero := pushDiesCreatureForTest(g, me.ID, "Hero of Bladehold", heroOfBladeholdOracle, "Creature — Human Knight", 3, 4)
	friend := pushDiesCreatureForTest(g, me.ID, "Friend", "", "Creature — Bear", 2, 2)

	declareAttack(t, g, opp.ID, hero, friend)
	// Two simultaneous triggers of one controller: the controller orders them.
	ordered := false
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerOrder && c.Chooser == me.ID {
			if len(c.TriggerOrderIDs) != 2 {
				t.Fatalf("%d triggers to order, want 2 (battle cry and the Soldiers)", len(c.TriggerOrderIDs))
			}
			ordered = true
		}
	}
	if !ordered {
		t.Fatal("no trigger-order prompt for Hero's two attack triggers")
	}
	for i := 0; i < 20; i++ {
		if c := pendingOfKind(g, game.PendingChoiceTriggerOrder); c != nil {
			answerTriggerOrderInOfferedOrder(t, g)
			continue
		}
		if stackFullyEmpty(g) {
			break
		}
		passPriorityAroundTable(t, g)
	}
	soldiers := battlefieldIDsNamed(g, "Soldier")
	if len(soldiers) != 2 {
		t.Fatalf("%d Soldiers, want 2", len(soldiers))
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Soldier" {
			continue
		}
		if !c.Tapped {
			t.Error("a Soldier is untapped; they enter tapped and attacking")
		}
		if c.AttackingTarget != opp.ID {
			t.Errorf("a Soldier attacks %s, want %s", c.AttackingTarget, opp.ID)
		}
	}
	// Battle cry: every other attacker got +1/+0; Hero itself did not.
	if got := effectivePower(t, g, friend); got != 3 {
		t.Errorf("Friend has power %d, want 3 after battle cry", got)
	}
	if got := effectivePower(t, g, hero); got != 3 {
		t.Errorf("Hero has power %d, want 3 (battle cry is for the others)", got)
	}
}
