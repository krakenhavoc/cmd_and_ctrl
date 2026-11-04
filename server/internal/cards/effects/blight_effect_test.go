package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blight_effect_test.go — #4683: "blight N" as a resolving effect
// (CR 701.68a), end to end through the cards that print it.

const (
	sourbreadAuntieOracle    = "6b087ae1-1412-4dd9-9b96-3b45be52d386"
	gutsplitterGangOracle    = "322c1c8f-ef8a-4a9f-b9ce-927720bad1d6"
	highPerfectMorcantOracle = "dd27c5fa-4355-4e5c-8219-b74cfe209da0"
	warrenTorchmasterOracle  = "20c3c8ab-e996-462e-a1f6-87421fc3ee55"
	dreamSeizerOracle        = "c7e4bc5d-e1d9-4fbc-8d53-0cec73086828"
	blightedBlackthornOracle = "a294968e-9a7f-4d66-bde0-e775ae5f8197"
	scuzzbackScroungerOracle = "af38d683-77a3-4599-ad2a-1c293ae01d45"
	boggartMischiefOracle    = "ac65ed8c-abe8-4044-9ce5-70b1370d0e19"
)

// blightPrompt is the open blight pick addressed to chooser, or nil.
func blightPrompt(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	c := latestChooseCardsFor(g, chooser)
	if c == nil || c.ChooseDestination != game.ChooseBlight {
		return nil
	}
	return c
}

// castAuntie casts Sourbread Auntie and settles until its blight is asked.
func castAuntie(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := castAndResolveCreature(t, g, "Sourbread Auntie", "Creature — Goblin Warrior", sourbreadAuntieOracle)
	monSettle(t, g)
	return id
}

func minusCounters(g *game.Game, id uuid.UUID) int {
	c := findBattlefieldCardByID(g, id)
	if c == nil {
		return -1
	}
	return c.Counters[game.CounterMinusOne]
}

func goblinTokens(g *game.Game, controller uuid.UUID) int {
	n, _ := countTokensNamed(g, controller, "Goblin")
	return n
}

func TestSourbreadAuntieBlightsACreatureYouControlAndMakesTwoGoblins(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := pushVanillaCreature(g, me.ID, "Bear", 3, 3)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 3, 3)
	auntie := castAuntie(t, g)

	p := blightPrompt(g, me.ID)
	if p == nil {
		t.Fatal("the blight is asked as the trigger resolves")
	}
	if p.ChooseMin != 0 || p.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1 (\"you may\", one creature)", p.ChooseMin, p.ChooseMax)
	}
	if !hasID(p.ChooseCards, bear) || !hasID(p.ChooseCards, auntie) || hasID(p.ChooseCards, theirs) {
		t.Errorf("offered %v: want the creatures I control, the Auntie included, and not an opponent's", p.ChooseCards)
	}
	if goblinTokens(g, me.ID) != 0 {
		t.Fatal("tokens before the blight was answered")
	}
	answerChooseCards(t, g, me.ID, bear)
	monSettle(t, g)

	if got := minusCounters(g, bear); got != 2 {
		t.Errorf("-1/-1 counters on the blighted creature: %d, want 2", got)
	}
	if got := minusCounters(g, auntie); got != 0 {
		t.Errorf("-1/-1 counters on the Auntie: %d, want 0", got)
	}
	if got := goblinTokens(g, me.ID); got != 2 {
		t.Fatalf("Goblin tokens: %d, want 2", got)
	}
	for _, c := range g.Battlefield.Cards {
		if IsToken(c) && c.Name == "Goblin" {
			if len(c.Colors) != 2 || c.Colors[0] != "B" || c.Colors[1] != "R" || c.Power != 1 || c.Toughness != 1 {
				t.Errorf("token %+v: want a 1/1 black and red Goblin", c)
			}
		}
	}
	// The blighting player placed the counters (CR 120.3d).
	placed := false
	for _, ev := range g.Events {
		if ev.Kind == game.EventCounterPlaced && ev.Target == bear && ev.Actor == me.ID {
			placed = true
		}
	}
	if !placed {
		t.Error("no counter placement on the blighted creature names the blighting player")
	}
}

func TestSourbreadAuntieDeclinedMakesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushVanillaCreature(g, me.ID, "Bear", 3, 3)
	castAuntie(t, g)
	answerChooseCards(t, g, me.ID)
	monSettle(t, g)
	if got := minusCounters(g, bear); got != 0 {
		t.Errorf("counters after a decline: %d", got)
	}
	if got := goblinTokens(g, me.ID); got != 0 {
		t.Errorf("Goblin tokens after a decline: %d, want 0", got)
	}
}

// The choice is not a target: a hexproof creature you control can be
// blighted.
func TestBlightCanChooseAHexproofCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hex := pushVanillaCreature(g, me.ID, "Warded", 3, 3)
	findBattlefieldCardByID(g, hex).Keywords = []string{"hexproof", "shroud"}
	castAuntie(t, g)
	p := blightPrompt(g, me.ID)
	if p == nil || !hasID(p.ChooseCards, hex) {
		t.Fatal("a hexproof, shrouded creature you control is a blight candidate")
	}
	answerChooseCards(t, g, me.ID, hex)
	monSettle(t, g)
	if got := minusCounters(g, hex); got != 2 {
		t.Errorf("counters: %d, want 2", got)
	}
	if got := goblinTokens(g, me.ID); got != 2 {
		t.Errorf("Goblin tokens: %d, want 2", got)
	}
}

// An effect's blight is "an effect would put counters" (CR 614.16), so
// Doubling Season doubles it — unlike the cost blight of #1703.
func TestDoublingSeasonDoublesAnEffectBlight(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Doubling Season",
		TypeLine: "Enchantment", OracleID: twDoublingSeasonOracle, Owner: me.ID, Controller: me.ID})
	wall := pushVanillaCreature(g, me.ID, "Wall", 0, 10)
	castAuntie(t, g)
	answerChooseCards(t, g, me.ID, wall)
	monSettle(t, g)
	if got := minusCounters(g, wall); got != 4 {
		t.Errorf("counters under Doubling Season: %d, want 4", got)
	}
	// Doubling Season doubles the tokens too.
	if got := goblinTokens(g, me.ID); got != 4 {
		t.Errorf("Goblin tokens under Doubling Season: %d, want 4", got)
	}
}

// Two counter replacements ask to be ordered (CR 616.1). The tokens wait
// for the answer, and the table waiting on it is a restore point.
func TestBlightWaitsForTheReplacementOrder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Doubling Season",
		TypeLine: "Enchantment", OracleID: twDoublingSeasonOracle, Owner: me.ID, Controller: me.ID})
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Winding Constrictor",
		TypeLine: "Creature — Snake", OracleID: twWindingConstrictorOra, Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID})
	wall := pushVanillaCreature(g, me.ID, "Wall", 0, 10)
	castAuntie(t, g)
	answerChooseCards(t, g, me.ID, wall)
	order := b06ReplacementOrderFor(g, me.ID)
	if order == nil {
		t.Fatal("Doubling Season and Winding Constrictor both apply: no CR 616 ordering prompt")
	}
	if got := goblinTokens(g, me.ID); got != 0 {
		t.Fatalf("tokens before the counters landed: %d", got)
	}
	g = restoreRoundTrip(t, g, true)
	order = b06ReplacementOrderFor(g, me.ID)
	if err := g.ResolveReplacementOrder(order.ID, me.ID, order.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	monSettle(t, g)
	if got := minusCounters(g, wall); got < 4 {
		t.Errorf("counters: %d, want the doubled and incremented count", got)
	}
	if got := goblinTokens(g, me.ID); got < 2 {
		t.Errorf("Goblin tokens after the answer: %d, want them made", got)
	}
}

// The blight question is a restore point.
func TestBlightPromptIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushVanillaCreature(g, me.ID, "Bear", 3, 3)
	castAuntie(t, g)
	if blightPrompt(g, me.ID) == nil {
		t.Fatal("no blight prompt")
	}
	restored := restoreRoundTrip(t, g, true)
	answerChooseCards(t, restored, me.ID, bear)
	monSettle(t, restored)
	if got := minusCounters(restored, bear); got != 2 {
		t.Errorf("restored: counters %d, want 2", got)
	}
	if got := goblinTokens(restored, me.ID); got != 2 {
		t.Errorf("restored: Goblin tokens %d, want 2", got)
	}
}

// "If you don't" (Gutsplitter Gang): a decline loses the life, a blight
// does not.
func TestGutsplitterGangLosesLifeOnlyWhenItDoesNotBlight(t *testing.T) {
	for _, tc := range []struct {
		name    string
		blight  bool
		wantPay int
	}{
		{"blighted", true, 0},
		{"declined", false, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			seat := g.Turn.ActiveSeat
			me := g.Seats[seat]
			gang := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Gutsplitter Gang",
				TypeLine: "Creature — Goblin Berserker", OracleID: gutsplitterGangOracle, Power: 6, Toughness: 6,
				Owner: me.ID, Controller: me.ID})
			advanceToStepOf(t, g, seat, game.StepPrecombatMain)
			monSettle(t, g)
			if blightPrompt(g, me.ID) == nil {
				t.Fatal("no blight prompt at the first main phase")
			}
			before := me.Life
			if tc.blight {
				answerChooseCards(t, g, me.ID, gang)
			} else {
				answerChooseCards(t, g, me.ID)
			}
			monSettle(t, g)
			if got := before - me.Life; got != tc.wantPay {
				t.Errorf("life lost: %d, want %d", got, tc.wantPay)
			}
			if tc.blight && minusCounters(g, gang) != 2 {
				t.Errorf("counters on the Gang: %d, want 2", minusCounters(g, gang))
			}
		})
	}
}

// "Each opponent blights 1" (High Perfect Morcant): each opponent picks a
// creature of their own, in turn order; one with a single creature is not
// asked, and one with none does not blight.
func TestEachOpponentBlightsTheirOwnCreature(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	o1, o2, o3 := g.Seats[(seat+1)%4], g.Seats[(seat+2)%4], g.Seats[(seat+3)%4]
	mine := pushVanillaCreature(g, me.ID, "Mine", 3, 3)
	a := pushVanillaCreature(g, o1.ID, "A", 2, 2)
	b := pushVanillaCreature(g, o1.ID, "B", 4, 4)
	lone := pushVanillaCreature(g, o2.ID, "Lone", 2, 2)
	findBattlefieldCardByID(g, lone).Keywords = []string{"hexproof"}
	_ = o3 // controls no creature

	morcant := castAndResolveCreature(t, g, "High Perfect Morcant", "Legendary Creature — Elf Noble", highPerfectMorcantOracle)
	monSettle(t, g)
	if blightPrompt(g, me.ID) != nil {
		t.Fatal("the controller does not blight")
	}
	p := blightPrompt(g, o1.ID)
	if p == nil {
		t.Fatal("the first opponent is asked")
	}
	if p.ChooseMin != 1 || p.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want exactly one (a mandatory blight)", p.ChooseMin, p.ChooseMax)
	}
	if !hasID(p.ChooseCards, a) || !hasID(p.ChooseCards, b) || hasID(p.ChooseCards, mine) || hasID(p.ChooseCards, morcant) || hasID(p.ChooseCards, lone) {
		t.Errorf("offered %v: want only the first opponent's own creatures", p.ChooseCards)
	}
	answerChooseCards(t, g, o1.ID, b)
	monSettle(t, g)
	if blightPrompt(g, o2.ID) != nil {
		t.Error("an opponent with one creature has no choice to make")
	}
	if got := minusCounters(g, b); got != 1 {
		t.Errorf("first opponent's choice: %d counters, want 1", got)
	}
	if got := minusCounters(g, a); got != 0 {
		t.Errorf("first opponent's other creature: %d counters, want 0", got)
	}
	if got := minusCounters(g, lone); got != 1 {
		t.Errorf("second opponent's hexproof creature: %d counters, want 1", got)
	}
	if got := minusCounters(g, mine); got != 0 {
		t.Errorf("my creature: %d counters, want 0", got)
	}
}

// "When you do" (Warren Torchmaster): the reflexive trigger's target is
// chosen after the blight, as it goes on the stack.
func TestWarrenTorchmasterHasteFollowsTheBlight(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	torch := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Warren Torchmaster",
		TypeLine: "Creature — Goblin Warrior", OracleID: warrenTorchmasterOracle, Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID})
	bear := pushVanillaCreature(g, me.ID, "Bear", 3, 3)
	advanceToStepOf(t, g, seat, game.StepBeginCombat)
	monSettle(t, g)
	if blightPrompt(g, me.ID) == nil {
		t.Fatal("no blight prompt at the beginning of combat")
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("the haste target is asked before the blight")
	}
	answerChooseCards(t, g, me.ID, torch)
	monSettle(t, g)
	if got := minusCounters(g, torch); got != 1 {
		t.Fatalf("counters on the Torchmaster: %d, want 1", got)
	}
	pickCard(t, g, me.ID, bear)
	monSettle(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if !effectiveHasKeyword(findBattlefieldCardByID(g, bear), "haste") {
		t.Error("the target did not gain haste")
	}
}

// Declining Warren Torchmaster's blight makes no reflexive trigger.
func TestWarrenTorchmasterDeclinedAsksNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Warren Torchmaster",
		TypeLine: "Creature — Goblin Warrior", OracleID: warrenTorchmasterOracle, Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID})
	advanceToStepOf(t, g, seat, game.StepBeginCombat)
	monSettle(t, g)
	answerChooseCards(t, g, me.ID)
	monSettle(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a declined blight made the reflexive trigger")
	}
}

func TestBlightCardsAreRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		sourbreadAuntieOracle:    "Sourbread Auntie",
		gutsplitterGangOracle:    "Gutsplitter Gang",
		highPerfectMorcantOracle: "High Perfect Morcant",
		warrenTorchmasterOracle:  "Warren Torchmaster",
		dreamSeizerOracle:        "Dream Seizer",
		blightedBlackthornOracle: "Blighted Blackthorn",
		scuzzbackScroungerOracle: "Scuzzback Scrounger",
		boggartMischiefOracle:    "Boggart Mischief",
	} {
		s, ok := Lookup(oracle)
		if !ok || s.Name != name {
			t.Errorf("%s: registered as %q (%v), want %q", oracle, s.Name, ok, name)
		}
	}
}

// CR 701.68b: with no creature to put the counters on there is nothing to
// choose, and "if you do" is false.
func TestBoggartMischiefWithNoCreatureMakesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Boggart Mischief", "Kindred Enchantment — Goblin", boggartMischiefOracle, nil)
	monSettle(t, g)
	if blightPrompt(g, me.ID) != nil {
		t.Fatal("asked to blight with no creature to blight")
	}
	if !g.Battlefield.Contains(findBattlefieldByName(g, "Boggart Mischief")) {
		t.Fatal("Boggart Mischief did not resolve")
	}
	if got := goblinTokens(g, me.ID); got != 0 {
		t.Errorf("Goblin tokens: %d, want 0", got)
	}
}

// A Goblin blighted to death is a Goblin creature you control dying: the
// drain triggers, and the two tokens still come (the blight was done).
func TestBoggartMischiefBlightsAGoblinToDeathAndDrains(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	gob := b16Creature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 1, 1, "R")
	castCatalogSpell(t, g, "Boggart Mischief", "Kindred Enchantment — Goblin", boggartMischiefOracle, nil)
	monSettle(t, g)
	if blightPrompt(g, me.ID) == nil {
		t.Fatal("no blight prompt")
	}
	myLife, theirLife := me.Life, opp.Life
	answerChooseCards(t, g, me.ID, gob)
	monSettle(t, g)
	if findBattlefieldCardByID(g, gob) != nil {
		t.Fatal("the 1/1 survived a -1/-1 counter")
	}
	if got := goblinTokens(g, me.ID); got != 2 {
		t.Errorf("Goblin tokens: %d, want 2", got)
	}
	if me.Life != myLife+1 || opp.Life != theirLife-1 {
		t.Errorf("life: me %d to %d, opponent %d to %d; want +1 / -1", myLife, me.Life, theirLife, opp.Life)
	}
}

// Scuzzback Scrounger's Treasure follows the blight.
func TestScuzzbackScroungerMakesATreasureWhenItBlights(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	scuzz := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Scuzzback Scrounger",
		TypeLine: "Creature — Goblin Warrior", OracleID: scuzzbackScroungerOracle, Power: 3, Toughness: 2,
		Owner: me.ID, Controller: me.ID})
	advanceToStepOf(t, g, seat, game.StepPrecombatMain)
	monSettle(t, g)
	answerChooseCards(t, g, me.ID, scuzz)
	monSettle(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("Treasures: %d, want 1", n)
	}
	if got := minusCounters(g, scuzz); got != 1 {
		t.Errorf("counters: %d, want 1", got)
	}
}

// Blighted Blackthorn: a blight draws a card and loses 1 life.
func TestBlightedBlackthornDrawsAndLosesLifeWhenItBlights(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	thorn := castAndResolveCreature(t, g, "Blighted Blackthorn", "Creature — Treefolk Warlock", blightedBlackthornOracle)
	monSettle(t, g)
	hand, life := me.Hand.Size(), me.Life
	answerChooseCards(t, g, me.ID, thorn)
	monSettle(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("cards drawn: %d, want 1", got)
	}
	if got := life - me.Life; got != 1 {
		t.Errorf("life lost: %d, want 1", got)
	}
}

// Dream Seizer: a blight makes each opponent discard.
func TestDreamSeizerBlightMakesEachOpponentDiscard(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	seizer := castAndResolveCreature(t, g, "Dream Seizer", "Creature — Faerie Rogue", dreamSeizerOracle)
	monSettle(t, g)
	answerChooseCards(t, g, me.ID, seizer)
	monSettle(t, g)
	if discardChoiceFor(g, opp.ID) == nil {
		t.Error("an opponent was not asked to discard")
	}
}
