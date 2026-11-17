package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b08_cards_test.go — March of the Machine, batch 8 (S82, #5477).
// Each test plays the real catalog entry through the engine.

const (
	momB08Arachnoid = "042cc9d9-ad13-401f-b2f5-c2f1558d7e44"
	momB08Change    = "aba7e9e2-8df8-454e-a332-9abae1d95b6f"
	momB08Hot       = "9666729d-3a06-43f9-a957-d5cc4d25b974"
	momB08Derision  = "4b74d2d9-d75b-496e-bbb4-2b95b3d3ba55"
	momB08Flourish  = "63a85361-56bc-4209-aea8-b49319bc7b92"
	momB08Avenged   = "ec83f398-31af-484d-b6d2-8c7add142260"
	momB08Truth     = "ef61fb5c-2ea6-4f82-92c5-c4713dee45fa"
	momB08Scrollsh  = "9f762b9f-7525-4ff8-bc1d-fbab66fd4013"
	momB08Seed      = "863d6607-912b-4186-9874-87d5473ec481"
	momB08Unseal    = "b3ea71d2-6298-452a-9c54-7c880b7e9cc0"
	momB08Shapecraf = "cf459f47-e565-4bbb-89e4-ed54da8adae1"
)

func TestMomB08EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		momB08Arachnoid: "Arachnoid Adaptation",
		momB08Change:    "Change the Equation",
		momB08Hot:       "Coming In Hot",
		momB08Derision:  "Deadly Derision",
		momB08Flourish:  "Final Flourish",
		momB08Avenged:   "Mirrodin Avenged",
		momB08Truth:     "Moment of Truth",
		momB08Scrollsh:  "Scrollshift",
		momB08Seed:      "Seed of Hope",
		momB08Unseal:    "Unseal the Necropolis",
		momB08Shapecraf: "Zhalfirin Shapecraft",
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

// momB08Cast puts the named instant in seat 0's hand and casts it strictly.
func momB08Cast(t *testing.T, g *game.Game, name, cost, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	advanceToMain(t, g)
	params.Strict = true
	return castFromHandForTest(t, g, me, name, "Instant", cost, oracle, params)
}

func momB08Refused(g *game.Game, name, cost, oracle string, params game.CastSpellParams) error {
	me := g.Seats[0]
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", ManaCost: cost,
			OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	params.Strict = true
	return g.CastSpell(me.ID, id, params)
}

func momB08Library(p *game.Player, names ...string) []uuid.UUID {
	// names are pushed so that names[0] ends up on top.
	ids := make([]uuid.UUID, len(names))
	for i := len(names) - 1; i >= 0; i-- {
		typeLine := "Creature — Test"
		if names[i][0] == '!' {
			typeLine = "Instant"
		}
		id := uuid.New()
		p.Library.PushTop(game.Card{InstanceID: id, Name: names[i], TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
		ids[i] = id
	}
	return ids
}

func TestMomB08ArachnoidAdaptation(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	land := b12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	tapCard(g, bear)
	advanceToMain(t, g)
	b06AddMana(me, "G")
	if err := momB08Refused(g, "Arachnoid Adaptation", "{G}", momB08Arachnoid, game.CastSpellParams{Targets: cardRefs(land)}); err == nil {
		t.Error("cast at a land")
	}
	momB08Cast(t, g, "Arachnoid Adaptation", "{G}", momB08Arachnoid, game.CastSpellParams{Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
	c := b12Card(t, g, bear)
	if !effectiveHasKeyword(&c, "reach") {
		t.Error("the Bear did not gain reach")
	}
	if c.Tapped {
		t.Error("the Bear was not untapped")
	}
}

func TestMomB08ChangeTheEquation(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cheap := b13OpponentCasts(t, g, opp, "Cheap Bolt", "Instant", "", "{R}", nil)
	b06AddMana(me, "U", "C")
	if err := momB08Refused(g, "Change the Equation", "{1}{U}", momB08Change,
		game.CastSpellParams{Modes: []int{1}, Targets: b16TargetCard(cheap)}); err != nil {
		// mode 2 takes red spells up to 6: a {R} spell is legal there too.
		t.Fatalf("mode two at a red {R} spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Stack.Size() != 0 || !opp.Graveyard.Contains(cheap) {
		t.Error("the red spell was not countered")
	}

	// Mode one: any colour, mana value 2 or less; refuses mana value 3.
	blue2 := b13OpponentCasts(t, g, opp, "Blue Two", "Instant", "", "{1}{U}", nil)
	three := b13OpponentCasts(t, g, opp, "Blue Three", "Instant", "", "{2}{U}", nil)
	b06AddMana(me, "U", "C")
	if err := momB08Refused(g, "Change the Equation", "{1}{U}", momB08Change,
		game.CastSpellParams{Modes: []int{0}, Targets: b16TargetCard(three)}); err == nil {
		t.Error("mode one at a mana value 3 spell")
	}
	b06AddMana(me, "U", "C")
	if err := momB08Refused(g, "Change the Equation", "{1}{U}", momB08Change,
		game.CastSpellParams{Modes: []int{1}, Targets: b16TargetCard(three)}); err == nil {
		t.Error("mode two at a blue spell")
	}
	b06AddMana(me, "U", "C")
	if err := momB08Refused(g, "Change the Equation", "{1}{U}", momB08Change,
		game.CastSpellParams{Modes: []int{0}, Targets: b16TargetCard(blue2)}); err != nil {
		t.Fatalf("mode one at a mana value 2 spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(blue2) {
		t.Error("the mana value 2 spell was not countered")
	}

	// Mode two refuses a red spell with mana value 7.
	big := b13OpponentCasts(t, g, opp, "Big Red", "Instant", "", "{6}{R}", nil)
	b06AddMana(me, "U", "C")
	if err := momB08Refused(g, "Change the Equation", "{1}{U}", momB08Change,
		game.CastSpellParams{Modes: []int{1}, Targets: b16TargetCard(big)}); err == nil {
		t.Error("mode two at a mana value 7 spell")
	}
}

func TestMomB08ComingInHot(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	land := b12Push(g, me.ID, "Mountain", "Basic Land — Mountain", "", 0, 0)
	advanceToMain(t, g)
	b06AddMana(me, "R")
	if err := momB08Refused(g, "Coming In Hot", "{R}", momB08Hot, game.CastSpellParams{Targets: cardRefs(land)}); err == nil {
		t.Error("cast at a land")
	}
	b06AddMana(me, "R")
	momB08Cast(t, g, "Coming In Hot", "{R}", momB08Hot, game.CastSpellParams{Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 3/2", p, tough)
	}
	c := b12Card(t, g, bear)
	if !effectiveHasKeyword(&c, "first strike") {
		t.Error("no first strike")
	}
	if scryChoiceFor(g, me.ID) == nil {
		t.Error("no scry 1 prompt")
	}
}

func TestMomB08DeadlyDerision(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 4}, Owner: opp.ID, Controller: opp.ID,
	})
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	advanceToMain(t, g)
	b06AddMana(me, "B", "B", "C", "C")
	if err := momB08Refused(g, "Deadly Derision", "{2}{B}{B}", momB08Derision, game.CastSpellParams{Targets: cardRefs(rock)}); err == nil {
		t.Error("cast at an artifact")
	}
	before := b27CountNamed(g, me.ID, "Treasure")
	momB08Cast(t, g, "Deadly Derision", "{2}{B}{B}", momB08Derision, game.CastSpellParams{Targets: cardRefs(ogre)})
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, ogre) != nil {
		t.Error("the creature survived")
	}
	if n := b27CountNamed(g, me.ID, "Treasure"); n != before+1 {
		t.Errorf("Treasures %d → %d, want one more", before, n)
	}
	b06AddMana(me, "B", "B", "C", "C")
	momB08Cast(t, g, "Deadly Derision", "{2}{B}{B}", momB08Derision, game.CastSpellParams{Targets: cardRefs(walker)})
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, walker) != nil {
		t.Error("the planeswalker survived")
	}
}

func TestMomB08FinalFlourish(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Giant", "Creature — Giant", 7, 7)
	rock := b12Push(g, me.ID, "Mind Stone", "Artifact", "", 0, 0)
	land := b12Push(g, me.ID, "Swamp", "Basic Land — Swamp", "", 0, 0)
	advanceToMain(t, g)
	b06AddMana(me, "B", "C")
	if err := momB08Refused(g, "Final Flourish", "{1}{B}", momB08Flourish,
		game.CastSpellParams{Targets: cardRefs(victim), OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Error("kicked by sacrificing a land")
	}
	b06AddMana(me, "B", "C")
	momB08Cast(t, g, "Final Flourish", "{1}{B}", momB08Flourish, game.CastSpellParams{Targets: cardRefs(victim)})
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, victim); p != 5 || tough != 5 {
		t.Errorf("unkicked: %d/%d, want 5/5", p, tough)
	}
	b06AddMana(me, "B", "C")
	momB08Cast(t, g, "Final Flourish", "{1}{B}", momB08Flourish,
		game.CastSpellParams{Targets: cardRefs(victim), OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{rock}})
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, rock) != nil {
		t.Error("the artifact was not sacrificed")
	}
	// -2/-2 then -6/-6 (kicked, instead of -2/-2 for that cast): 7-2-6 = -1.
	g.RunStateChecksForTest()
	if findBattlefieldCardForTest(g, victim) != nil {
		t.Error("the Giant survived -8/-8 over two casts")
	}
}

func TestMomB08FinalFlourishKickedIsSixNotEight(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Wurm", "Creature — Wurm", 7, 7)
	rock := b12Push(g, me.ID, "Mind Stone", "Artifact", "", 0, 0)
	advanceToMain(t, g)
	b06AddMana(me, "B", "C")
	momB08Cast(t, g, "Final Flourish", "{1}{B}", momB08Flourish,
		game.CastSpellParams{Targets: cardRefs(victim), OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{rock}})
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, victim); p != 1 || tough != 1 {
		t.Errorf("kicked: %d/%d, want 1/1", p, tough)
	}
}

func TestMomB08MirrodinAvenged(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hurt := b12Creature(g, opp.ID, "Hurt Ogre", "Creature — Ogre", 4, 4)
	fresh := b12Creature(g, opp.ID, "Fresh Ogre", "Creature — Ogre", 4, 4)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, hurt, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	advanceToMain(t, g)
	b06AddMana(me, "B")
	if err := momB08Refused(g, "Mirrodin Avenged", "{B}", momB08Avenged, game.CastSpellParams{Targets: cardRefs(fresh)}); err == nil {
		t.Error("cast at an undamaged creature")
	}
	b06AddMana(me, "B")
	hand := me.Hand.Size()
	momB08Cast(t, g, "Mirrodin Avenged", "{B}", momB08Avenged, game.CastSpellParams{Targets: cardRefs(hurt)})
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, hurt) != nil {
		t.Error("the damaged creature survived")
	}
	if findBattlefieldCardForTest(g, fresh) == nil {
		t.Error("the undamaged creature died")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one more (the draw)", hand, me.Hand.Size())
	}
}

func TestMomB08Scrollshift(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, mine, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	advanceToMain(t, g)
	b06AddMana(me, "W", "C", "C")
	if err := momB08Refused(g, "Scrollshift", "{2}{W}", momB08Scrollsh, game.CastSpellParams{Targets: cardRefs(theirs)}); err == nil {
		t.Error("cast at an opponent's creature")
	}
	b06AddMana(me, "W", "C", "C")
	hand := me.Hand.Size()
	momB08Cast(t, g, "Scrollshift", "{2}{W}", momB08Scrollsh, game.CastSpellParams{Targets: cardRefs(mine)})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one more (the draw)", hand, me.Hand.Size())
	}
	if findBattlefieldCardForTest(g, mine) != nil {
		t.Error("the Bear kept its object identity (a blink returns a new object, CR 400.7)")
	}
	if n := b27CountNamed(g, me.ID, "Bear"); n != 1 {
		t.Errorf("%d Bears on the battlefield after the blink, want 1", n)
	}
}

func TestMomB08ScrollshiftWithNoTargetStillDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b06AddMana(me, "W", "C", "C")
	hand := me.Hand.Size()
	momB08Cast(t, g, "Scrollshift", "{2}{W}", momB08Scrollsh, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one more (the draw, with no target chosen)", hand, me.Hand.Size())
	}
}

func TestMomB08ZhalfirinShapecraft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	giant := b12Creature(g, me.ID, "Giant", "Creature — Giant", 6, 6)
	mouse := b12Creature(g, me.ID, "Mouse", "Creature — Mouse", 1, 1)
	land := b12Push(g, me.ID, "Island", "Basic Land — Island", "", 0, 0)
	advanceToMain(t, g)
	b06AddMana(me, "U", "C")
	if err := momB08Refused(g, "Zhalfirin Shapecraft", "{1}{U}", momB08Shapecraf, game.CastSpellParams{Targets: cardRefs(land)}); err == nil {
		t.Error("cast at a land")
	}
	b06AddMana(me, "U", "C")
	momB08Cast(t, g, "Zhalfirin Shapecraft", "{1}{U}", momB08Shapecraf, game.CastSpellParams{Targets: cardRefs(giant)})
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, giant); p != 4 || tough != 3 {
		t.Errorf("the Giant is %d/%d, want base 4/3", p, tough)
	}
	b06AddMana(me, "U", "C")
	momB08Cast(t, g, "Zhalfirin Shapecraft", "{1}{U}", momB08Shapecraf, game.CastSpellParams{Targets: cardRefs(mouse)})
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, mouse); p != 4 || tough != 3 {
		t.Errorf("the Mouse is %d/%d, want base 4/3", p, tough)
	}
}

func TestMomB08SeedOfHope(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	ids := momB08Library(me, "Milled Bear", "!Milled Bolt")
	b06AddMana(me, "G")
	life := me.Life
	momB08Cast(t, g, "Seed of Hope", "{G}", momB08Seed, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("Seed of Hope offered no permanent card")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != ids[0] || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("offered %v (%d..%d), want only the milled creature, up to one", pick.ChooseCards, pick.ChooseMin, pick.ChooseMax)
	}
	answerChooseCards(t, g, me.ID, ids[0])
	if !me.Hand.Contains(ids[0]) {
		t.Error("the permanent did not reach the hand")
	}
	if !me.Graveyard.Contains(ids[1]) {
		t.Error("the instant stayed out of the graveyard")
	}
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
}

func TestMomB08SeedOfHopeWithNoPermanentStillGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	momB08Library(me, "!One", "!Two")
	b06AddMana(me, "G")
	life := me.Life
	momB08Cast(t, g, "Seed of Hope", "{G}", momB08Seed, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("offered a card when none of the milled cards was a permanent")
	}
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
}

func TestMomB08MomentOfTruth(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	ids := momB08Library(me, "Card A", "Card B", "Card C")
	b06AddMana(me, "U", "C")
	size := me.Library.Size()
	momB08Cast(t, g, "Moment of Truth", "{1}{U}", momB08Truth, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, ids[1]) // Card B into the hand
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, ids[0]) // Card A into the graveyard
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) || !me.Graveyard.Contains(ids[0]) {
		t.Errorf("hand has B %v, graveyard has A %v", me.Hand.Contains(ids[1]), me.Graveyard.Contains(ids[0]))
	}
	if me.Library.Size() != size-2 {
		t.Errorf("library %d → %d, want two fewer", size, me.Library.Size())
	}
	cards := me.Library.Cards
	if len(cards) == 0 || cards[0].InstanceID != ids[2] {
		t.Error("Card C is not on the bottom of the library")
	}
	if me.Library.Contains(ids[0]) || me.Library.Contains(ids[1]) {
		t.Error("a chosen card is still in the library")
	}
}

func TestMomB08UnsealTheNecropolis(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mine := momB08Library(me, "My Bear", "My Wolf", "!My Bolt")
	theirs := momB08Library(opp, "Their Bear", "!Their Bolt", "!Their Zap")
	old := b17GraveyardCard(me, "Old Ogre", "Creature — Ogre", "{3}")
	b06AddMana(me, "B", "C", "C")
	momB08Cast(t, g, "Unseal the Necropolis", "{2}{B}", momB08Unseal, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	for _, id := range append(append([]uuid.UUID{}, mine...), theirs...) {
		if me.Library.Contains(id) || opp.Library.Contains(id) {
			t.Fatalf("a milled card %s is still in a library", id)
		}
	}
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no return prompt")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 2 {
		t.Errorf("bounds %d..%d, want 0..2", pick.ChooseMin, pick.ChooseMax)
	}
	for _, id := range pick.ChooseCards {
		if id == theirs[0] {
			t.Error("an opponent's milled creature was offered")
		}
		if id == mine[2] {
			t.Error("a noncreature card was offered")
		}
	}
	answerChooseCards(t, g, me.ID, mine[0], old)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(mine[0]) || !me.Hand.Contains(old) {
		t.Error("the chosen creature cards are not in hand")
	}
	if !opp.Graveyard.Contains(theirs[0]) {
		t.Error("the opponent's creature was not left in their graveyard")
	}
}
