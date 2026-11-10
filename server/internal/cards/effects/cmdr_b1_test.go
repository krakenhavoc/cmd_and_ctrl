package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Foundations Commander angel deck, batches 1 and 2 (#5368).

const (
	cmdrGiadaOracle             = "48e6d3d8-2f27-4017-acdd-40bce8cdbc02"
	cmdrAlwaysWatchingOracle    = "c2153115-bc12-458d-baeb-538ec61e1242"
	cmdrAngelicSleuthOracle     = "2579e9e0-b00f-4a1b-8fce-15d06bb7fe5a"
	cmdrDefyDeathOracle         = "dbf9c3e8-35dd-4b80-bd23-f43e21cfd991"
	cmdrEmeriaShepherdOracle    = "473da1b6-232b-4428-a66d-30262246b47a"
	cmdrGraspOfFateOracle       = "1b26a54f-6c27-4940-b77e-2a66175c5eff"
	cmdrHeraldOfWarOracle       = "1228bd6f-565d-4ad5-bd24-e5aa1ead5dee"
	cmdrInvokeTheDivineOracle   = "87bd0fdb-c5b9-46ea-9858-1a870960351f"
	cmdrLinvalaOracle           = "64f5689d-504d-40e9-bdb5-36dcefd257e2"
	cmdrMerchantOfTruthOracle   = "84275aab-f718-4896-b61c-63bf49fb483b"
	cmdrMetropolisReformerOracl = "baec04a1-3bc9-4618-afab-e42013e36648"
	cmdrNornsChoirmasterOracle  = "bad6a5a9-7010-418d-b042-182926aa999d"
	cmdrReyaDawnbringerOracle   = "08d1a5d1-141d-48e3-ac22-c2529191cb13"
	cmdrSearchThePremisesOracle = "0232418d-1c52-4573-b488-c62345ca1c8d"
	cmdrSegovianAngelOracle     = "50372c54-359c-4ee6-9e21-1e40ba53652d"
	cmdrSeraphOfTheSwordOracle  = "84ee74e9-b3e3-4f47-8b3c-5c1a70f5cad6"
	cmdrSerraAvengerOracle      = "b91b65b3-b35c-460e-a49b-333000a9a519"
	cmdrStarnheimAspirantOracle = "e7b92789-47a9-430a-ad57-be7b6c3d63cb"
	cmdrSunblastAngelOracle     = "8b34d1df-4024-497d-bd28-a7caa8c3fa24"
)

func TestCmdrB1CardsAreRegisteredFull(t *testing.T) {
	want := map[string]string{
		cmdrGiadaOracle:             "Giada, Font of Hope",
		cmdrAlwaysWatchingOracle:    "Always Watching",
		cmdrAngelicSleuthOracle:     "Angelic Sleuth",
		cmdrDefyDeathOracle:         "Defy Death",
		cmdrEmeriaShepherdOracle:    "Emeria Shepherd",
		cmdrGraspOfFateOracle:       "Grasp of Fate",
		cmdrHeraldOfWarOracle:       "Herald of War",
		cmdrInvokeTheDivineOracle:   "Invoke the Divine",
		cmdrLinvalaOracle:           "Linvala, the Preserver",
		cmdrMerchantOfTruthOracle:   "Merchant of Truth",
		cmdrMetropolisReformerOracl: "Metropolis Reformer",
		cmdrNornsChoirmasterOracle:  "Norn's Choirmaster",
		cmdrReyaDawnbringerOracle:   "Reya Dawnbringer",
		cmdrSearchThePremisesOracle: "Search the Premises",
		cmdrSegovianAngelOracle:     "Segovian Angel",
		cmdrSeraphOfTheSwordOracle:  "Seraph of the Sword",
		cmdrSerraAvengerOracle:      "Serra Avenger",
		cmdrStarnheimAspirantOracle: "Starnheim Aspirant",
		cmdrSunblastAngelOracle:     "Sunblast Angel",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s is not Full", name)
		}
	}
}

// cmdrPush puts a permanent on seat `owner`'s side with printed P/T.
func cmdrPush(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// cmdrSettle drives the stack empty, taking the offered order for any
// trigger-ordering prompt (two abilities that trigger together).
func cmdrSettle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 32; i++ {
		if ch := triggerOrderPrompt(g); ch != nil {
			if err := g.ResolveTriggerOrder(ch.ID, ch.Chooser, append([]uuid.UUID(nil), ch.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		passPriorityAroundTable(t, g)
		if triggerOrderPrompt(g) == nil {
			return
		}
	}
	t.Fatal("never settled")
}

// cmdrCountNamed counts the battlefield permanents with a name.
func cmdrCountNamed(g *game.Game, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			n++
		}
	}
	return n
}

// --- Segovian Angel ---------------------------------------------------

func TestCmdrSegovianAngelHasFlyingAndVigilance(t *testing.T) {
	g := newCatalogGame(t)
	id := cmdrPush(g, g.Seats[0].ID, "Segovian Angel", "Creature — Angel", cmdrSegovianAngelOracle, 1, 1)
	for _, kw := range []string{"flying", "vigilance"} {
		if !hasKeywordOnBattlefield(t, g, id, kw) {
			t.Errorf("no %s", kw)
		}
	}
}

// --- Giada, Font of Hope ----------------------------------------------

func TestCmdrGiadaGivesOtherAngelsACounterPerAngelAlreadyThere(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	giada := cmdrPush(g, me.ID, "Giada, Font of Hope", "Legendary Creature — Angel", cmdrGiadaOracle, 2, 2)
	if counterCount(g, giada, game.CounterPlusOne) != 0 {
		t.Fatal("Giada does not count herself as entering")
	}
	first := castAndResolveCreature(t, g, "Angel One", "Creature — Angel", "")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, first, game.CounterPlusOne); got != 1 {
		t.Errorf("first Angel: %d counters, want 1 (Giada)", got)
	}
	second := castAndResolveCreature(t, g, "Angel Two", "Creature — Angel", "")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, second, game.CounterPlusOne); got != 2 {
		t.Errorf("second Angel: %d counters, want 2 (Giada and the first Angel)", got)
	}
	bear := castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("a non-Angel got %d counters", got)
	}
	// An Angel token entering is an Angel too.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("3/3 white Angel with flying"), 1) })
	passPriorityAroundTable(t, g)
	if got := counterCount(g, findBattlefieldByName(g, "Angel"), game.CounterPlusOne); got != 3 {
		t.Errorf("Angel token: %d counters, want 3", got)
	}
}

func TestCmdrGiadaDoesNotHelpAnOpponentsAngels(t *testing.T) {
	g := newCatalogGame(t)
	cmdrPush(g, g.Seats[1].ID, "Giada, Font of Hope", "Legendary Creature — Angel", cmdrGiadaOracle, 2, 2)
	angel := castAndResolveCreature(t, g, "Angel One", "Creature — Angel", "")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, angel, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's Giada gave %d counters", got)
	}
}

func TestCmdrGiadaManaOnlyCastsAngels(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	giada := cmdrPush(g, me.ID, "Giada, Font of Hope", "Legendary Creature — Angel", cmdrGiadaOracle, 2, 2)
	activateManaFor(t, g, me.ID, giada, 0, game.ManaAbilityParams{})
	bear := handCardFull(me, "Bear", "Creature — Bear", "{W}", "", []string{"W"})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("Giada's mana paid for a Bear")
	}
	angel := handCardFull(me, "Segovian Angel", "Creature — Angel", "{W}", cmdrSegovianAngelOracle, []string{"W"})
	if err := g.CastSpell(me.ID, angel, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("Giada's mana could not pay for an Angel: %v", err)
	}
}

// --- Always Watching ----------------------------------------------------

func TestCmdrAlwaysWatchingAffectsOnlyYourNontokenCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := cmdrPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := cmdrPush(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, BlackZombieToken(), 1) })
	token := findBattlefieldByName(g, "Zombie")
	tokenBase := effectivePower(t, g, token)
	cmdrPush(g, me.ID, "Always Watching", "Enchantment", cmdrAlwaysWatchingOracle, 0, 0)

	if got := effectivePower(t, g, mine); got != 3 {
		t.Errorf("my nontoken creature: power %d, want 3", got)
	}
	if !hasKeywordOnBattlefield(t, g, mine, "vigilance") {
		t.Error("my nontoken creature has no vigilance")
	}
	if got := effectivePower(t, g, theirs); got != 2 || hasKeywordOnBattlefield(t, g, theirs, "vigilance") {
		t.Errorf("an opponent's creature was affected (power %d)", got)
	}
	if got := effectivePower(t, g, token); got != tokenBase || hasKeywordOnBattlefield(t, g, token, "vigilance") {
		t.Errorf("my token was affected (power %d, base %d)", got, tokenBase)
	}
}

// --- Angelic Sleuth ------------------------------------------------------

func cmdrClues(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == "Clue" {
			n++
		}
	}
	return n
}

func TestCmdrAngelicSleuthInvestigatesWhenAPermanentWithCountersLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sleuth := cmdrPush(g, me.ID, "Angelic Sleuth", "Creature — Angel Advisor", cmdrAngelicSleuthOracle, 2, 3)
	counted := cmdrPush(g, me.ID, "Counted", "Creature — Bear", "", 2, 2)
	plain := cmdrPush(g, me.ID, "Plain", "Creature — Bear", "", 2, 2)
	foreign := cmdrPush(g, opp.ID, "Foreign", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(counted, game.CounterPlusOne, 1)
		_ = g.AddCounterForEffect(foreign, game.CounterPlusOne, 1)
	})

	b27Kill(g, plain)
	passPriorityAroundTable(t, g)
	if got := cmdrClues(g, me.ID); got != 0 {
		t.Fatalf("a permanent with no counters made %d Clues", got)
	}
	b27Kill(g, foreign)
	passPriorityAroundTable(t, g)
	if got := cmdrClues(g, me.ID); got != 0 {
		t.Fatalf("an opponent's permanent made %d Clues", got)
	}
	b27Kill(g, counted)
	passPriorityAroundTable(t, g)
	if got := cmdrClues(g, me.ID); got != 1 {
		t.Fatalf("a permanent with a counter made %d Clues, want 1", got)
	}
	// "Another": the Sleuth itself leaving with counters is not it.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(sleuth, game.CounterPlusOne, 1) })
	b27Kill(g, sleuth)
	passPriorityAroundTable(t, g)
	if got := cmdrClues(g, me.ID); got != 1 {
		t.Fatalf("the Sleuth leaving made a Clue (%d)", got)
	}
}

// --- Merchant of Truth ----------------------------------------------------

func TestCmdrMerchantOfTruthInvestigatesForNontokenCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmdrPush(g, me.ID, "Merchant of Truth", "Creature — Angel Detective", cmdrMerchantOfTruthOracle, 2, 5)
	bear := cmdrPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	foreign := cmdrPush(g, opp.ID, "Foreign", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, BlackZombieToken(), 1) })
	token := findBattlefieldByName(g, "Zombie")

	b27Kill(g, token)
	b27Kill(g, foreign)
	passPriorityAroundTable(t, g)
	if got := cmdrClues(g, me.ID); got != 0 {
		t.Fatalf("a token or an opponent's creature made %d Clues", got)
	}
	b27Kill(g, bear)
	passPriorityAroundTable(t, g)
	if got := cmdrClues(g, me.ID); got != 1 {
		t.Fatalf("a nontoken creature of mine made %d Clues, want 1", got)
	}
}

func TestCmdrMerchantOfTruthTriggersForItselfAndForCreaturesThatDieWithIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	merchant := cmdrPush(g, me.ID, "Merchant of Truth", "Creature — Angel Detective", cmdrMerchantOfTruthOracle, 2, 5)
	bear := cmdrPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentsForEffect([]uuid.UUID{merchant, bear})
	})
	passPriorityAroundTable(t, g)
	if got := cmdrClues(g, me.ID); got != 2 {
		t.Fatalf("Merchant and a Bear dying together made %d Clues, want 2", got)
	}
}

func TestCmdrMerchantOfTruthGivesYourCluesExalted(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmdrPush(g, me.ID, "Merchant of Truth", "Creature — Angel Detective", cmdrMerchantOfTruthOracle, 2, 5)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, ClueToken(), 1)
		_ = g.CreateTokenForEffect(opp.ID, ClueToken(), 1)
	})
	var mine, theirs uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Clue" && c.Controller == me.ID {
			mine = c.InstanceID
		}
		if c.Name == "Clue" && c.Controller == opp.ID {
			theirs = c.InstanceID
		}
	}
	count := func(id uuid.UUID) int {
		n := 0
		for _, a := range effectiveAbilities(t, g, id) {
			if a == game.KeywordExalted {
				n++
			}
		}
		return n
	}
	if got := count(mine); got != 1 {
		t.Errorf("my Clue has %d instances of exalted, want 1", got)
	}
	if got := count(theirs); got != 0 {
		t.Errorf("an opponent's Clue has %d instances of exalted", got)
	}
	// A second Merchant is a second instance (CR 113.2c).
	cmdrPush(g, me.ID, "Merchant of Truth", "Creature — Angel Detective", cmdrMerchantOfTruthOracle, 2, 5)
	if got := count(mine); got != 2 {
		t.Errorf("with two Merchants my Clue has %d instances, want 2", got)
	}
}

// --- Metropolis Reformer ---------------------------------------------------

func TestCmdrMetropolisReformerGivesYouHexproofAndLifeForDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	reformer := cmdrPush(g, me.ID, "Metropolis Reformer", "Creature — Angel Cleric", cmdrMetropolisReformerOracl, 2, 3)
	if got := playerAbilities(g, me); !hasPlayerAbility(got, KeywordHexproof) {
		t.Fatalf("you do not have hexproof: %v", got)
	}
	src := cmdrPush(g, g.Seats[1].ID, "Pinger", "Creature — Bear", "", 1, 1)
	life := me.Life
	b27Damage(g, src, reformer, 2)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life from 2 damage, want 2", got)
	}
	// Dies to lethal damage, and still gains that much.
	b27Damage(g, src, reformer, 3)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := me.Life - life; got != 5 {
		t.Errorf("gained %d life in total, want 5", got)
	}
}

// --- Linvala, the Preserver -------------------------------------------------

func TestCmdrLinvalaNeedsAnOpponentAhead(t *testing.T) {
	// Behind on life and on creatures: both abilities.
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Life = me.Life + 5
	cmdrPush(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	cmdrPush(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2) // Linvala counts herself
	life := me.Life
	castAndResolveCreature(t, g, "Linvala, the Preserver", "Legendary Creature — Angel", cmdrLinvalaOracle)
	cmdrSettle(t, g)
	if got := me.Life - life; got != 5 {
		t.Errorf("gained %d life, want 5", got)
	}
	if got := cmdrCountNamed(g, "Angel"); got != 1 {
		t.Errorf("%d Angel tokens, want 1", got)
	}

	// Level on life and creatures: neither.
	g = newCatalogGame(t)
	me = g.Seats[0]
	life = me.Life
	castAndResolveCreature(t, g, "Linvala, the Preserver", "Legendary Creature — Angel", cmdrLinvalaOracle)
	cmdrSettle(t, g)
	if me.Life != life || cmdrCountNamed(g, "Angel") != 0 {
		t.Errorf("Linvala did something level with the table (life %d→%d, tokens %d)", life, me.Life, cmdrCountNamed(g, "Angel"))
	}
}

func TestCmdrLinvalaChecksEachAbilityOnItsOwn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Ahead on life, behind on creatures: only the token.
	me.Life = opp.Life + 3
	cmdrPush(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	cmdrPush(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2) // Linvala counts herself
	life := me.Life
	castAndResolveCreature(t, g, "Linvala, the Preserver", "Legendary Creature — Angel", cmdrLinvalaOracle)
	cmdrSettle(t, g)
	if me.Life != life {
		t.Errorf("gained life while ahead on life")
	}
	if got := cmdrCountNamed(g, "Angel"); got != 1 {
		t.Errorf("%d Angel tokens, want 1", got)
	}
}

// --- Search the Premises -------------------------------------------------------

func TestCmdrSearchThePremisesInvestigatesForAttacksAtYou(t *testing.T) {
	g := newCatalogGame(t)
	me, defender, other := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4], g.Seats[(g.Turn.ActiveSeat+2)%4]
	cmdrPush(g, defender.ID, "Search the Premises", "Enchantment", cmdrSearchThePremisesOracle, 0, 0)
	a := pushVanillaCreature(g, me.ID, "Attacker One", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Attacker Two", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, defender.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(b, other.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := cmdrClues(g, defender.ID); got != 1 {
		t.Fatalf("one attacker at me and one at another player made %d Clues, want 1", got)
	}
}

func TestCmdrSearchThePremisesCountsEachAttackerAndPlaneswalkers(t *testing.T) {
	g := newCatalogGame(t)
	me, defender := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	cmdrPush(g, defender.ID, "Search the Premises", "Enchantment", cmdrSearchThePremisesOracle, 0, 0)
	walker := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: walker, Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: defender.ID, Controller: defender.ID, Counters: map[string]int{game.CounterLoyalty: 3},
	})
	a := pushVanillaCreature(g, me.ID, "Attacker One", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Attacker Two", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: a, Target: defender.ID}, {Attacker: b, Target: walker}}); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	cmdrSettle(t, g)
	if got := cmdrClues(g, defender.ID); got != 2 {
		t.Fatalf("an attacker at me and one at my planeswalker made %d Clues, want 2", got)
	}
}
