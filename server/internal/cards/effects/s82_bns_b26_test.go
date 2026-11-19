package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 BNS batch 26 — Foundations Jumpstart creatures (#5929).

const (
	s82BnsWoodcrasher = "faaf9975-74df-4e9d-be0c-19a446cd507c"
	s82BnsSpy         = "39b4a47a-df51-4bd9-9a20-43ac80a5d352"
	s82BnsLieutenant  = "57d0c688-6206-4c83-9fcb-27fd29e2a9ed"
	s82BnsBurglar     = "9c975fda-471a-4d41-9bb6-ea42edc6a0e6"
	s82BnsExpert      = "341cf654-faf7-4db9-bd9a-70783f1ccff0"
	s82BnsInciter     = "826b83ba-0fe5-4f1d-91c8-799a1100d1e6"
	s82BnsAerialist   = "ae78b6f9-8334-479e-8111-8e3acb2b7ca7"
	s82BnsIntimidator = "6ad1d8d0-06c0-4804-af5f-60ffb358f371"
	s82BnsCleric      = "fe8d2874-c626-4d4e-a7f5-11c175a2b3bb"
	s82BnsNecromancer = "6494e3c4-5489-44a5-94b6-4ee2d7830896"
	s82BnsHippocamp   = "5ba04e0d-a3eb-448d-8a97-08c155ef9694"
)

func TestS82BnsB26CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		s82BnsWoodcrasher: "Baloth Woodcrasher", s82BnsSpy: "Balustrade Spy",
		s82BnsLieutenant: "Basri's Lieutenant", s82BnsBurglar: "Blood Burglar",
		s82BnsExpert: "Bloodfire Expert", s82BnsInciter: "Bloodlust Inciter",
		s82BnsAerialist: "Bloodthirsty Aerialist", s82BnsIntimidator: "Boldwyr Intimidator",
		s82BnsCleric: "Bonecaller Cleric", s82BnsNecromancer: "Boneclad Necromancer",
		s82BnsHippocamp: "Breaching Hippocamp",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82BnsB26WoodcrasherGrowsAndTramplesOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	crasher := b12Push(g, me.ID, "Baloth Woodcrasher", "Creature — Beast", s82BnsWoodcrasher, 4, 4)
	if hasEffectiveKeyword(t, g, crasher, "trample") {
		t.Fatal("trample before any land entered")
	}
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if c := apaLive(g, crasher); c.CurrentPower() != 8 || c.CurrentToughness() != 8 {
		t.Errorf("Woodcrasher is %d/%d after landfall, want 8/8", c.CurrentPower(), c.CurrentToughness())
	}
	if !hasEffectiveKeyword(t, g, crasher, "trample") {
		t.Error("no trample after landfall")
	}
}

func TestS82BnsB26SpyMillsUntilALand(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	putOnTopForTest(opp, "Beyond", "Creature — Bear")
	putOnTopForTest(opp, "Plains", "Basic Land — Plains")
	putOnTopForTest(opp, "Second Spell", "Sorcery")
	putOnTopForTest(opp, "First Spell", "Instant")
	graveyard, library := opp.Graveyard.Size(), opp.Library.Size()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Balustrade Spy", TypeLine: "Creature — Vampire Rogue",
		OracleID: s82BnsSpy, Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size() - graveyard; got != 3 {
		t.Errorf("%d cards milled, want 3 (two spells and the land)", got)
	}
	if got := library - opp.Library.Size(); got != 3 {
		t.Errorf("library fell by %d, want 3", got)
	}
}

func TestS82BnsB26LieutenantCountersThenMakesAKnightWhenACounteredCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	lt := castCreatureCard(t, g, "Basri's Lieutenant", "Creature — Human Knight", s82BnsLieutenant, 3, 4)
	_ = lt
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, bear, "+1/+1"); n != 1 {
		t.Fatalf("+1/+1 counters on the Bear: %d, want 1", n)
	}
	if !hasEffectiveKeyword(t, g, lt, "protection from multicolored") || !hasEffectiveKeyword(t, g, lt, "vigilance") {
		t.Error("missing vigilance or protection from multicolored")
	}
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	var knight *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Controller == me.ID && c.Name == "Knight" {
			knight = c
		}
	}
	if knight == nil {
		t.Fatal("no Knight token after a creature with a +1/+1 counter died")
	}
	if knight.Power != 2 || knight.Toughness != 2 || !game.HasKeyword(knight, "vigilance") {
		t.Errorf("Knight is %d/%d, keywords %v", knight.Power, knight.Toughness, knight.Keywords)
	}
}

func TestS82BnsB26LieutenantMakesNothingForACreatureWithoutACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Basri's Lieutenant", "Creature — Human Knight", s82BnsLieutenant, 3, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Knight"); n != 0 {
		t.Errorf("%d Knights, want none for a creature with no counter", n)
	}
}

func TestS82BnsB26LieutenantMakesAKnightWhenItselfDiesWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lt := b12Push(g, me.ID, "Basri's Lieutenant", "Creature — Human Knight", s82BnsLieutenant, 3, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == lt {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 1}
			}
		}
	})
	killCreature(t, g, me.ID, lt)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Knight"); n != 1 {
		t.Errorf("%d Knights, want 1", n)
	}
}

func TestS82BnsB26BurglarHasLifelinkOnlyDuringItsControllersTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	mine := b12Push(g, me.ID, "Blood Burglar", "Creature — Vampire Rogue", s82BnsBurglar, 2, 2)
	theirs := b12Push(g, opp.ID, "Blood Burglar", "Creature — Vampire Rogue", s82BnsBurglar, 2, 2)
	if !hasEffectiveKeyword(t, g, mine, "lifelink") {
		t.Error("no lifelink on its controller's turn")
	}
	if hasEffectiveKeyword(t, g, theirs, "lifelink") {
		t.Error("lifelink on an opponent's turn")
	}
}

func TestS82BnsB26ExpertHasProwess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	e := b12Push(g, me.ID, "Bloodfire Expert", "Creature — Efreet Monk", s82BnsExpert, 3, 1)
	if !hasEffectiveKeyword(t, g, e, "prowess") {
		t.Error("no prowess")
	}
}

func TestS82BnsB26InciterGivesHasteAndTaps(t *testing.T) {
	g, me, opp := spendTable(t)
	inciter := b12Push(g, me.ID, "Bloodlust Inciter", "Creature — Human Warrior", s82BnsInciter, 1, 1)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	p7Activate(t, g, me, inciter, 0, game.ActivateAbilityParams{Targets: rfCardTarget(bear)})
	if !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the target did not gain haste")
	}
	if !tappedNow(g, inciter) {
		t.Error("the Inciter did not tap")
	}
}

func TestS82BnsB26AerialistGrowsWhenYouGainLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := b12Push(g, me.ID, "Bloodthirsty Aerialist", "Creature — Vampire Rogue", s82BnsAerialist, 2, 3)
	gainLifeNow(t, g, a, opp, 2)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, a, "+1/+1"); n != 0 {
		t.Fatalf("an opponent's life gain gave %d counters", n)
	}
	gainLifeNow(t, g, a, me, 2)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, a, "+1/+1"); n != 1 {
		t.Errorf("+1/+1 counters %d, want 1", n)
	}
}

func TestS82BnsB26IntimidatorMakesCowardsAndWarriors(t *testing.T) {
	g, me, opp := spendTable(t)
	boldwyr := b12Push(g, me.ID, "Boldwyr Intimidator", "Creature — Giant Warrior", s82BnsIntimidator, 5, 5)
	theirs := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	mine := b12Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 1)

	if err := g.ActivateCatalogAbility(me.ID, boldwyr, 0, game.ActivateAbilityParams{Targets: rfCardTarget(theirs), Strict: true}); err == nil {
		t.Fatal("activated {R} with no mana")
	}
	apaMana(me, "R")
	p7Activate(t, g, me, boldwyr, 0, game.ActivateAbilityParams{Targets: rfCardTarget(theirs), Strict: true})
	apaMana(me, "R", "C", "C")
	p7Activate(t, g, me, boldwyr, 1, game.ActivateAbilityParams{Targets: rfCardTarget(mine), Strict: true})

	if !slices.Contains(effectiveOf(t, g, theirs).Subtypes, "Coward") || !slices.Contains(effectiveOf(t, g, theirs).Subtypes, "Elf") {
		t.Errorf("their Elf subtypes %v, want Elf and Coward", effectiveOf(t, g, theirs).Subtypes)
	}
	if !slices.Contains(effectiveOf(t, g, mine).Subtypes, "Warrior") {
		t.Errorf("my Elf subtypes %v, want Warrior", effectiveOf(t, g, mine).Subtypes)
	}
	var cowardCantBlock bool
	g.ReadSnapshot(func() {
		w, _ := battlefieldCard(g, mine)
		c, _ := battlefieldCard(g, theirs)
		cowardCantBlock = !g.CanBlockLocked(&w, &c)
	})
	if !cowardCantBlock {
		t.Error("a Coward blocked a Warrior")
	}
}

func TestS82BnsB26ClericSacrificesToReanimateAtSorcerySpeed(t *testing.T) {
	g, me, _ := spendTable(t)
	cleric := b12Push(g, me.ID, "Bonecaller Cleric", "Creature — Human Cleric", s82BnsCleric, 2, 1)
	dead := graveyardCreature(me, "Dead Bear", "Bear", 2, 2)
	apaMana(me, "B", "C")
	if err := g.ActivateCatalogAbility(me.ID, cleric, 0, game.ActivateAbilityParams{Targets: rfCardTarget(dead), Strict: true}); err == nil {
		t.Fatal("activated {3}{B} with two mana")
	}
	apaMana(me, "C", "C")
	p7Activate(t, g, me, cleric, 0, game.ActivateAbilityParams{Targets: rfCardTarget(dead), Strict: true})
	if _, ok := battlefieldCard(g, dead); !ok {
		t.Error("the creature card did not return")
	}
	if _, ok := battlefieldCard(g, cleric); ok {
		t.Error("the Cleric was not sacrificed")
	}
}

func TestS82BnsB26NecromancerExilesForAZombie(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	dead := pushGraveyardCreature(g, opp.ID, "Dead Elf", "{1}{G}")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Boneclad Necromancer", TypeLine: "Creature — Human Wizard",
		OracleID: s82BnsNecromancer, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, dead)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(dead) {
		t.Fatal("the creature card was not exiled")
	}
	var zombie *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Controller == me.ID && c.Name == "Zombie" {
			zombie = c
		}
	}
	if zombie == nil || zombie.Power != 2 || zombie.Toughness != 2 {
		t.Fatalf("no 2/2 Zombie token: %+v", zombie)
	}
}

func TestS82BnsB26NecromancerMakesNoZombieWhenYouDecline(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	dead := pushGraveyardCreature(g, opp.ID, "Dead Elf", "{1}{G}")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Boneclad Necromancer", TypeLine: "Creature — Human Wizard",
		OracleID: s82BnsNecromancer, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, dead)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(dead) {
		t.Error("exiled despite declining")
	}
	if n := countBattlefieldByName(g, "Zombie"); n != 0 {
		t.Errorf("%d Zombies, want none", n)
	}
}

func TestS82BnsB26HippocampUntapsAnotherCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Tapped Bear",
		TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Tapped: true, Owner: me.ID, Controller: me.ID})
	castCreatureCard(t, g, "Breaching Hippocamp", "Creature — Horse Fish", s82BnsHippocamp, 3, 2)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if tappedNow(g, bear) {
		t.Error("the Bear is still tapped")
	}
	if !hasEffectiveKeyword(t, g, findBattlefieldByName(g, "Breaching Hippocamp"), "flash") {
		t.Error("no flash")
	}
}
