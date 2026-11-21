package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rad_counter_batch2_test.go pins the second rad-counter batch (#2809):
// the batched "one or more nonland cards are milled" trigger, "a card in
// a graveyard that was milled this turn", a trigger that works from the
// graveyard, Strong's once-per-damage-event enrage and its radiation
// replacement, and surveil no longer counting as a mill.

const (
	radMirelurkQueen          = "116ef9ea-d199-49b8-89c7-d61dd92711f5"
	radScreechingScorchbeast  = "ed5889c9-e524-40c1-9011-7559a992bb1b"
	radWiseMothman            = "7656f93b-336f-49cb-b8fe-c2b5c014258f"
	radTatoFarmer             = "71b5a3b4-31bb-44d8-a80c-36efff2a621e"
	radMasterTranscendent     = "40262ae5-1472-4777-b759-d1752c0b0148"
	radInfestingRadroach      = "064f371c-e4f2-423b-9f35-443191dfbb4c"
	radStrongBrutishThespian  = "f0ef48e1-c3d7-4ce2-8f7c-653342778cae"
	radBatch2NonlandCreature  = "Creature — Bear"
	radBatch2BasicForestTyped = "Basic Land — Forest"
)

// radStack puts named cards on top of p's library, the first listed on
// top, replacing what was there. A name starting "Land " is a basic
// Forest; anything else is a vanilla creature card.
func radStack(p *game.Player, names ...string) []uuid.UUID {
	cards := make([]game.Card, 0, len(names))
	for _, n := range names {
		tl := radBatch2NonlandCreature
		if len(n) > 5 && n[:5] == "Land " {
			tl = radBatch2BasicForestTyped
		}
		cards = append(cards, game.Card{Name: n, TypeLine: tl, Power: 2, Toughness: 2})
	}
	return effectstest.SeedSearchLibrary(p, cards...)
}

// radMill has p mill n cards as an effect would, then settles the table.
func radMill(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.MillNForEffect(p.ID, n); err != nil {
			t.Fatalf("MillNForEffect: %v", err)
		}
	})
	g.RunStateChecksForTest()
}

// radSurveilAll has p surveil n and put every looked-at card into the
// graveyard.
func radSurveilAll(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	g.WithWriteLock(func() { g.SurveilKeyedForEffect(p.ID, uuid.Nil, n, game.KeyedTail{}) })
	c := effectstest.SurveilChoiceFor(g, p.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if err := g.ResolveSurveil(c.ID, p.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	g.RunStateChecksForTest()
}

// radGraveyardCard puts c in p's graveyard as a card p owns, as if it
// had been there since before the turn began.
func radGraveyardCard(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Graveyard.PushTop(c)
	return c.InstanceID
}

func radCreatureOnField(g *game.Game, owner uuid.UUID, name, oracle, typeLine string, p, tough int) uuid.UUID {
	return effectstest.ApaPush(g, owner, owner, game.Card{Name: name, OracleID: oracle, TypeLine: typeLine, Power: p, Toughness: tough})
}

// --- the batched mill trigger -----------------------------------------

// One mill of three cards, two of them nonland, is ONE trigger; a second
// mill the same turn triggers nothing ("only once each turn").
func TestMirelurkQueenTriggersOncePerMillAndOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	queen := radCreatureOnField(g, me.ID, "Mirelurk Queen", radMirelurkQueen, "Creature — Crab Mutant", 4, 4)
	radStack(opp, "Bear", "Land Forest", "Ogre", "Elf", "Wolf")
	hand := me.Hand.Size()

	radMill(t, g, opp, 3)
	if n := effectstest.RadTriggersFrom(g, queen); n != 1 {
		t.Fatalf("%d triggers for one mill of two nonland cards, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if n := effectstest.Pr8Counters(g, queen, game.CounterPlusOne); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1", n)
	}

	radMill(t, g, opp, 2)
	if n := effectstest.RadTriggersFrom(g, queen); n != 0 {
		t.Errorf("a second mill the same turn made %d triggers, want 0", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d after the second mill, want still 1", got)
	}
}

// A mill of only lands is not "nonland cards milled", and it does not use
// up the turn; a surveil is not a mill at all (CR 701.25a).
func TestMirelurkQueenIgnoresLandOnlyMillsAndSurveil(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	queen := radCreatureOnField(g, me.ID, "Mirelurk Queen", radMirelurkQueen, "Creature — Crab Mutant", 4, 4)
	radStack(opp, "Land Forest", "Land Forest 2", "Bear", "Ogre")

	radMill(t, g, opp, 2)
	if n := effectstest.RadTriggersFrom(g, queen); n != 0 {
		t.Fatalf("a mill of two lands made %d triggers, want 0", n)
	}
	radSurveilAll(t, g, opp, 1)
	if opp.Graveyard.Size() != 3 {
		t.Fatalf("setup: the surveil should have binned the Bear (graveyard %d)", opp.Graveyard.Size())
	}
	if n := effectstest.RadTriggersFrom(g, queen); n != 0 {
		t.Fatalf("a surveil made %d triggers, want 0: surveil is not a mill", n)
	}
	radMill(t, g, opp, 1)
	if n := effectstest.RadTriggersFrom(g, queen); n != 1 {
		t.Errorf("the first nonland mill of the turn made %d triggers, want 1", n)
	}
}

func TestMirelurkQueenEntersAndIrradiatesTargetPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Mirelurk Queen", "Creature — Crab Mutant", radMirelurkQueen, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RadOf(opp) != 2 || effectstest.RadOf(me) != 0 {
		t.Errorf("rad: opponent %d (want 2), me %d (want 0)", effectstest.RadOf(opp), effectstest.RadOf(me))
	}
}

// "That many" is the nonland cards of the mill; "do this only once each
// turn" is used up by a "yes", not by a "no".
func TestScreechingScorchbeastMakesThatManyZombieMutantsOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := radCreatureOnField(g, me.ID, "Screeching Scorchbeast", radScreechingScorchbeast, "Creature — Bat Mutant", 5, 5)
	radStack(opp, "Bear", "Land Forest", "Ogre", "Elf", "Wolf", "Cat", "Rat")

	radMill(t, g, opp, 3)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Zombie Mutant"); n != 0 {
		t.Fatalf("declined, but %d tokens", n)
	}

	radMill(t, g, opp, 2)
	if n := effectstest.RadTriggersFrom(g, beast); n != 1 {
		t.Fatalf("a no did not use up the turn: want 1 trigger, got %d", n)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Zombie Mutant"); n != 2 {
		t.Fatalf("%d Zombie Mutants for two nonland cards, want 2", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Zombie Mutant" && (c.CurrentPower() != 2 || c.CurrentToughness() != 2 || !c.HasSubtype("Mutant") || !c.HasSubtype("Zombie")) {
			t.Errorf("token %s is %d/%d %s, want a 2/2 Zombie Mutant", c.Name, c.CurrentPower(), c.CurrentToughness(), c.TypeLine)
		}
	}

	radMill(t, g, opp, 2)
	if n := effectstest.RadTriggersFrom(g, beast); n != 0 {
		t.Errorf("tokens already made this turn: want no trigger, got %d", n)
	}
}

func TestScreechingScorchbeastAttackIrradiatesEachPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := radCreatureOnField(g, me.ID, "Screeching Scorchbeast", radScreechingScorchbeast, "Creature — Bat Mutant", 5, 5)
	effectstest.DeclareAttack(t, g, opp.ID, beast)
	effectstest.PassPriorityAroundTable(t, g)
	radEach(t, g, 2)
}

// X is the nonland cards milled; the counters go on up to X targets.
func TestWiseMothmanCountersUpToXTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	moth := radCreatureOnField(g, me.ID, "The Wise Mothman", radWiseMothman, "Legendary Creature — Insect Mutant", 3, 3)
	a := effectstest.Pr7Creature(g, me.ID, "A", 2, "G")
	b := effectstest.Pr7Creature(g, opp.ID, "B", 2, "G")
	c := effectstest.Pr7Creature(g, me.ID, "C", 2, "G")
	radStack(opp, "Bear", "Land Forest", "Ogre", "Elf")

	radMill(t, g, opp, 3)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target pick for the Mothman's trigger")
	}
	if p.PickTargetMax != 2 || p.PickTargetMin != 0 {
		t.Fatalf("pick bounds %d..%d, want 0..2 (two nonland cards)", p.PickTargetMin, p.PickTargetMax)
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c}}); err == nil {
		t.Fatal("three targets for X = 2 were accepted")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	if n := effectstest.RadTriggersFrom(g, moth); n != 1 {
		t.Fatalf("%d triggers for one mill, want 1", n)
	}

	// A restore while the trigger waits on the stack keeps its X.
	restored := effectstest.RestoreThroughJSON(t, g)
	for _, tbl := range []*game.Game{g, restored} {
		effectstest.PassPriorityAroundTable(t, tbl)
		if effectstest.Pr8Counters(tbl, a, game.CounterPlusOne) != 1 || effectstest.Pr8Counters(tbl, b, game.CounterPlusOne) != 1 || effectstest.Pr8Counters(tbl, c, game.CounterPlusOne) != 0 {
			t.Errorf("counters a=%d b=%d c=%d, want 1 1 0", effectstest.Pr8Counters(tbl, a, game.CounterPlusOne),
				effectstest.Pr8Counters(tbl, b, game.CounterPlusOne), effectstest.Pr8Counters(tbl, c, game.CounterPlusOne))
		}
	}
}

func TestWiseMothmanEntersOrAttacksIrradiatesEachPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "The Wise Mothman", "Legendary Creature — Insect Mutant", radWiseMothman, nil)
	effectstest.PassPriorityAroundTable(t, g)
	radEach(t, g, 1)
	moth := uuid.Nil
	for _, c := range g.Battlefield.Cards {
		if c.Name == "The Wise Mothman" {
			moth = c.InstanceID
		}
	}
	effectstest.B10Awake(g, moth)
	effectstest.DeclareAttack(t, g, opp.ID, moth)
	effectstest.PassPriorityAroundTable(t, g)
	radEach(t, g, 2)
	_ = me
}

// Each player milling in one resolution is one occurrence (CR 603.2c):
// one trigger counting every player's nonland cards.
func TestOneOrMoreNonlandCardsMilledIsOneTriggerAcrossPlayers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	moth := radCreatureOnField(g, me.ID, "The Wise Mothman", radWiseMothman, "Legendary Creature — Insect Mutant", 3, 3)
	radStack(me, "Mine", "Land Forest")
	radStack(opp, "Theirs", "Their Other")
	g.WithWriteLock(func() {
		_ = g.MillNForEffect(me.ID, 2)
		_ = g.MillNForEffect(opp.ID, 2)
	})
	g.RunStateChecksForTest()
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target pick: the mill did not trigger the Mothman")
	}
	if p.PickTargetMax != 3 {
		t.Errorf("X = %d, want 3 nonland cards across both mills", p.PickTargetMax)
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolvePickTargets (none): %v", err)
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a second target pick: the two mills should be one trigger")
	}
	if n := effectstest.RadTriggersFrom(g, moth); n != 1 {
		t.Errorf("%d triggers, want 1", n)
	}
}

// --- milled this turn --------------------------------------------------

func TestTatoFarmerReanimatesOnlyALandMilledThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	farmer := radCreatureOnField(g, me.ID, "Tato Farmer", radTatoFarmer, "Creature — Zombie Mutant Peasant", 1, 4)
	ids := radStack(opp, "Land Milled", "Land Surveilled", "Bear")
	old := radGraveyardCard(opp, game.Card{Name: "Old Forest", TypeLine: radBatch2BasicForestTyped})

	radMill(t, g, opp, 1)
	radSurveilAll(t, g, opp, 1)
	for _, bad := range []uuid.UUID{old, ids[1]} {
		if err := effectstest.ActivateOn(t, g, me.ID, farmer, 0, bad); err == nil {
			t.Fatalf("a land not milled this turn (%s) was a legal target", bad)
		}
	}
	// A restore keeps the provenance: it is read off this turn's log.
	g = effectstest.RestoreThroughJSON(t, g)
	if err := effectstest.ActivateOn(t, g, me.ID, farmer, 0, ids[0]); err != nil {
		t.Fatalf("the milled land: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	land := effectstest.FindBattlefieldCardForTest(g, ids[0])
	if land == nil || !land.Tapped || land.Controller != me.ID {
		t.Fatalf("the milled land should be on the battlefield tapped under my control: %+v", land)
	}
}

func TestTatoFarmerLandfallMayGetTwoRadCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	radCreatureOnField(g, me.ID, "Tato Farmer", radTatoFarmer, "Creature — Zombie Mutant Peasant", 1, 4)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RadOf(me) != 2 {
		t.Errorf("rad %d, want 2", effectstest.RadOf(me))
	}
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RadOf(me) != 2 {
		t.Errorf("declined, rad %d, want still 2", effectstest.RadOf(me))
	}
}

func TestTheMasterReanimatesAMilledCreatureAsAGreenMutant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	master := radCreatureOnField(g, me.ID, "The Master, Transcendent", radMasterTranscendent, "Legendary Artifact Creature — Mutant", 2, 4)
	ids := effectstest.SeedSearchLibrary(opp,
		game.Card{Name: "Blue Wizard", TypeLine: "Artifact Creature — Human Wizard", Power: 1, Toughness: 1, Colors: []string{"U"}},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	dead := radGraveyardCard(opp, game.Card{Name: "Dead Ogre", TypeLine: "Creature — Ogre", Power: 5, Toughness: 5})
	radMill(t, g, opp, 1)
	if err := effectstest.ActivateOn(t, g, me.ID, master, 0, dead); err == nil {
		t.Fatal("a creature card not milled this turn was a legal target")
	}
	if err := effectstest.ActivateOn(t, g, me.ID, master, 0, ids[0]); err != nil {
		t.Fatalf("the milled creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.FindBattlefieldCardForTest(g, ids[0])
	if c == nil || c.Controller != me.ID {
		t.Fatalf("the milled creature should be mine on the battlefield: %+v", c)
	}
	ch := c.Effective()
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("%d/%d, want base 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if !c.HasSubtype("Mutant") || c.HasSubtype("Wizard") || c.HasSubtype("Human") {
		t.Errorf("creature types %v, want Mutant only", ch.Subtypes)
	}
	if len(ch.Colors) != 1 || ch.Colors[0] != "G" {
		t.Errorf("colours %v, want green only", ch.Colors)
	}
	if !c.IsArtifact() {
		t.Error("an artifact creature stays an artifact")
	}
}

func TestTheMasterEntersAndIrradiatesTargetPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "The Master, Transcendent", "Legendary Artifact Creature — Mutant", radMasterTranscendent, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RadOf(opp) != 2 {
		t.Errorf("rad %d, want 2", effectstest.RadOf(opp))
	}
}

// --- a trigger from the graveyard --------------------------------------

func TestInfestingRadroachReturnsFromTheGraveyardOnAnOpponentsNonlandMill(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	roach := radGraveyardCard(me, game.Card{Name: "Infesting Radroach", OracleID: radInfestingRadroach, TypeLine: "Creature — Insect Mutant", Power: 2, Toughness: 2})
	radStack(me, "My Bear")
	radStack(opp, "Land Forest", "Bear", "Ogre")

	radMill(t, g, me, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("my own mill triggered the Radroach")
	}
	radMill(t, g, opp, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("an opponent's land mill triggered the Radroach")
	}
	// Two nonland cards: two triggers; the first "yes" returns it and the
	// second is removed without asking (CR 603.4).
	radMill(t, g, opp, 2)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(roach) {
		t.Fatal("the Radroach should be in my hand")
	}
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Error("the second trigger asked again after the card left the graveyard")
	}
}

func TestInfestingRadroachGivesThatManyRadCountersAndDoesNotTriggerOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	roach := radCreatureOnField(g, me.ID, "Infesting Radroach", radInfestingRadroach, "Creature — Insect Mutant", 2, 2)
	effectstest.DealCombatDamageToPlayer(g, roach, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RadOf(opp) != 3 {
		t.Fatalf("rad %d, want 3 (the damage dealt)", effectstest.RadOf(opp))
	}
	radStack(opp, "Bear")
	radMill(t, g, opp, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Error("the graveyard ability triggered on the battlefield")
	}
}

// --- Strong -----------------------------------------------------------

func TestStrongEnrageTriggersOncePerDamageEvent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	strong := radCreatureOnField(g, me.ID, "Strong, the Brutish Thespian", radStrongBrutishThespian, "Legendary Creature — Mutant Berserker", 7, 7)
	srcA := effectstest.Pr7Creature(g, opp.ID, "A", 1, "R")
	srcB := effectstest.Pr7Creature(g, opp.ID, "B", 1, "R")

	g.WithWriteLock(func() {
		_ = g.DamageInstanceForEffect(func() error {
			_ = g.DealDamageToCreatureForEffect(srcA, strong, 1)
			return g.DealDamageToCreatureForEffect(srcB, strong, 1)
		})
	})
	g.RunStateChecksForTest()
	if n := effectstest.RadTriggersFrom(g, strong); n != 1 {
		t.Fatalf("two sources at once made %d enrage triggers, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RadOf(me) != 3 || effectstest.Pr8Counters(g, strong, game.CounterPlusOne) != 3 {
		t.Fatalf("rad %d counters %d, want 3 and 3", effectstest.RadOf(me), effectstest.Pr8Counters(g, strong, game.CounterPlusOne))
	}

	// Two separate damage instructions are two events.
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(srcA, strong, 1)
		_ = g.DealDamageToCreatureForEffect(srcB, strong, 1)
	})
	g.RunStateChecksForTest()
	if n := effectstest.RadTriggersFrom(g, strong); n != 2 {
		t.Errorf("two separate instructions made %d triggers, want 2", n)
	}
}

// "You gain life rather than lose life from radiation": the rad
// counters' own trigger heals Strong's controller, and still takes the
// counters off. Life lost any other way is untouched.
func TestStrongGainsLifeRatherThanLosingItFromRadiation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	radCreatureOnField(g, opp.ID, "Strong, the Brutish Thespian", radStrongBrutishThespian, "Legendary Creature — Mutant Berserker", 7, 7)
	radStack(opp, "Bear", "Land Forest", "Ogre", "Elf")
	if err := g.AddPlayerCounter(opp.ID, game.CounterRad, 3); err != nil {
		t.Fatal(err)
	}
	life := opp.Life

	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life+2 {
		t.Errorf("life %d -> %d, want +2 (two nonland cards milled)", life, opp.Life)
	}
	if effectstest.RadOf(opp) != 1 {
		t.Errorf("rad %d, want 1 (two of three removed)", effectstest.RadOf(opp))
	}
	life = opp.Life
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -2) })
	if opp.Life != life-2 {
		t.Errorf("a loss not from radiation should be untouched: %d -> %d", life, opp.Life)
	}
}

// --- Struggle for Project Purity ----------------------------------------

const radStruggleForProjectPurity = "5958e9e3-9457-48e1-afc1-a5c89e3b0ed0"

// Brotherhood: each opponent draws, and you draw one per card they
// really drew — an opponent with an empty library draws nothing.
func TestStruggleBrotherhoodDrawsOnePerCardOpponentsDrew(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.CastSiege(t, g, "Struggle for Project Purity", radStruggleForProjectPurity, "Brotherhood")
	effectstest.AdvanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	empty := g.Seats[(seat+1)%len(g.Seats)]
	empty.Library.Cards = nil
	hands := map[uuid.UUID]int{}
	opps := 0
	for _, p := range g.Seats {
		hands[p.ID] = p.Hand.Size()
		if p.ID != me.ID && p.ID != empty.ID {
			opps++
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hands[me.ID]; got != opps {
		t.Errorf("drew %d, want %d (one per opponent who drew)", got, opps)
	}
	for _, p := range g.Seats {
		// The empty-library player drew nothing and loses the game for
		// it (CR 704.5b); their hand is not the point.
		if p.ID == me.ID || p.ID == empty.ID {
			continue
		}
		if got := p.Hand.Size() - hands[p.ID]; got != 1 {
			t.Errorf("%s drew %d, want 1", p.Name, got)
		}
	}
}

// Enclave: twice the creatures that attacked you, not your planeswalker
// and not another player.
func TestStruggleEnclaveIrradiatesAPlayerWhoAttacksYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	oppSeat := (seat + 1) % len(g.Seats)
	opp, other := g.Seats[oppSeat], g.Seats[(seat+2)%len(g.Seats)]
	effectstest.CastSiege(t, g, "Struggle for Project Purity", radStruggleForProjectPurity, "Enclave")
	walker := effectstest.P1g2Planeswalker(g, me.ID, 5)
	a := effectstest.Pr7Creature(g, opp.ID, "A", 2, "R")
	b := effectstest.Pr7Creature(g, opp.ID, "B", 2, "R")
	c := effectstest.Pr7Creature(g, opp.ID, "C", 2, "R")
	d := effectstest.Pr7Creature(g, opp.ID, "D", 2, "R")
	effectstest.AdvanceToMainOf(t, g, oppSeat)
	for _, id := range []uuid.UUID{a, b, c, d} {
		effectstest.B10Awake(g, id)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, at := range []struct{ who, target uuid.UUID }{{a, me.ID}, {b, me.ID}, {c, walker}, {d, other.ID}} {
		if err := g.DeclareAttacker(at.who, at.target); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RadOf(opp) != 4 {
		t.Errorf("attacker rad %d, want 4 (twice the two creatures attacking you)", effectstest.RadOf(opp))
	}
	if effectstest.RadOf(me) != 0 || effectstest.RadOf(other) != 0 {
		t.Errorf("only the attacking player gets them: me %d, other %d", effectstest.RadOf(me), effectstest.RadOf(other))
	}
}

// --- surveil is not a mill ----------------------------------------------

// Colossal Grave-Reaver reads the batch through the shared
// library-to-graveyard helper, so a surveilled creature card comes back.
func TestColossalGraveReaverReturnsASurveilledCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Colossal Grave-Reaver", "Creature — Dragon", effectstest.B17ColossalGraveReaverOracle, false)
	ids := radStack(me, "Surveilled Bear")
	radSurveilAll(t, g, me, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, ids[0]); c == nil {
		t.Error("the surveilled creature card should be on the battlefield")
	}
}

// A "put into your graveyard from your library" payoff sees a surveil; a
// "mills" payoff does not.
func TestSurveilIsPutIntoTheGraveyardFromTheLibraryButNotAMill(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Sidisi, Brood Tyrant", "Legendary Creature — Snake Shaman", effectstest.B33SidisiOracle, false)
	radCreatureOnField(g, me.ID, "Glowing One", radGlowingOne, "Creature — Zombie Mutant", 2, 2)
	radStack(me, "Surveilled Bear")
	life := me.Life

	radSurveilAll(t, g, me, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Zombie"); n != 1 {
		t.Errorf("Sidisi made %d Zombies for a surveilled creature card, want 1", n)
	}
	if me.Life != life {
		t.Errorf("Glowing One gained life for a surveil: %d -> %d", life, me.Life)
	}
}
