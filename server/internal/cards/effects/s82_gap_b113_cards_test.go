package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b113_cards_test.go — S82 GAP batch 113 (Commander Masters, #5948).
// Every test plays the real catalog card through the engine.

const (
	gapB113Nahiri     = "a554aa2c-d947-4fc6-929d-d5891f467615"
	gapB113Wanderer   = "4c2196e5-6ed7-4cd2-b8e7-f01aff236569"
	gapB113Calamity   = "4bb3edc5-2616-4d0d-b9dc-32ed79f38f67"
	gapB113Feast      = "c3601f22-7b22-4116-8633-ba67841524bd"
	gapB113Magnetism  = "a1bdf7f1-e3f7-4a13-9cf0-a2394df6cc74"
	gapB113Courage    = "41b929ad-0c44-4e50-84d3-13b318035543"
	gapB113Rampage    = "ee6d271c-0728-4080-85bb-a6af05f8b213"
	gapB113Slice      = "463fc961-d34e-4f40-b383-5b78a0fcb5c8"
	gapB113Star       = "48220e6c-5752-46e0-9b7f-f0eef274d929"
	gapB113BoltOracle = "4457ed35-7c10-48c8-9776-456485fdf070"
)

func TestGapB113AllAreFull(t *testing.T) {
	for o, name := range map[string]string{
		gapB113Nahiri: "Nahiri, the Lithomancer", gapB113Wanderer: "The Wanderer",
		gapB113Calamity: "Calamity of the Titans", gapB113Feast: "Feast of Succession",
		gapB113Magnetism: "Animal Magnetism", gapB113Courage: "Courage in Crisis",
		gapB113Rampage: "Predatory Rampage", gapB113Slice: "Slice and Dice",
		gapB113Star: "Star of Extinction",
	} {
		spec, ok := Lookup(o)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v, name %q, completeness %v, caveats %v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestGapB113FeastOfSuccessionShrinksEveryCreatureAndCrownsYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 4, 4)
	big := b12Creature(g, opp.ID, "Big", "Creature — Bear", 5, 5)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Feast of Succession", "Sorcery", gapB113Feast, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || g.Battlefield.Contains(mine) {
		t.Error("a creature with toughness 4 or less survived")
	}
	if p, tough := ptOf(t, g, big); p != 1 || tough != 1 {
		t.Errorf("the 5/5 is %d/%d, want 1/1", p, tough)
	}
	if g.Monarch != me.ID {
		t.Errorf("monarch is %v, want the caster", g.Monarch)
	}
}

func TestGapB113CourageInCrisisProliferatesTheCounterItPlaced(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Courage in Crisis", "Sorcery", gapB113Courage, cardRefs(bear))
	passPriorityAroundTable(t, g)
	answerProliferate(t, g, bear)
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, bear); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2 (one placed, one proliferated)", got)
	}
}

func TestGapB113CourageInCrisisNeedsATarget(t *testing.T) {
	g := newCatalogGame(t)
	if err := castCatalogSpellErrStrict(t, g, "Courage in Crisis", "Sorcery", gapB113Courage, nil); err == nil {
		t.Error("cast without a target")
	}
}

func TestGapB113SliceAndDiceDealsFourToEachCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	four := b12Creature(g, opp.ID, "Four", "Creature — Bear", 2, 4)
	five := b12Creature(g, me.ID, "Five", "Creature — Bear", 2, 5)
	castCatalogSpell(t, g, "Slice and Dice", "Sorcery", gapB113Slice, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(four) {
		t.Error("the 4-toughness creature survived 4 damage")
	}
	if !g.Battlefield.Contains(five) {
		t.Error("the 5-toughness creature died to 4 damage")
	}
}

func TestGapB113SliceAndDiceCycleMayDealOneToEachCreature(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
		one := b12Creature(g, opp.ID, "One", "Creature — Bear", 1, 1)
		two := b12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
		seedLibrary(me, "A", "B")
		hand := len(me.Hand.Cards)
		cycleFromHand(t, g, "Slice and Dice", "Sorcery", gapB113Slice, "{C}{C}{R}")
		answerLatestTriggerPrompt(t, g, me.ID, yes)
		passPriorityAroundTable(t, g)
		if got := len(me.Hand.Cards); got != hand+1 {
			t.Errorf("yes=%v: hand %d -> %d, want the cycled card out and a card drawn", yes, hand, got)
		}
		if g.Battlefield.Contains(one) == yes {
			t.Errorf("yes=%v: the 1/1 is on the battlefield = %v", yes, g.Battlefield.Contains(one))
		}
		if !g.Battlefield.Contains(two) {
			t.Errorf("yes=%v: the 2/2 died to 1 damage", yes)
		}
	}
}

func TestGapB113StarOfExtinctionDestroysALandAndDealsTwentyToCreaturesAndWalkers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	other := b12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	big := b12Creature(g, opp.ID, "Big", "Creature — Wurm", 19, 19)
	huge := b12Creature(g, me.ID, "Huge", "Creature — Wurm", 21, 21)
	walker := pushWalkerForTest(g, opp.ID, "Walker", "", 20)
	castCatalogSpell(t, g, "Star of Extinction", "Sorcery", gapB113Star, cardRefs(land))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the target land survived")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("another land was destroyed")
	}
	if g.Battlefield.Contains(big) {
		t.Error("a 19/19 survived 20 damage")
	}
	if !g.Battlefield.Contains(huge) {
		t.Error("a 21/21 died to 20 damage")
	}
	if g.Battlefield.Contains(walker) {
		t.Error("a 20-loyalty planeswalker survived 20 damage")
	}
}

func TestGapB113StarOfExtinctionOnlyTargetsALand(t *testing.T) {
	g := newCatalogGame(t)
	bear := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErrStrict(t, g, "Star of Extinction", "Sorcery", gapB113Star, cardRefs(bear)); err == nil {
		t.Error("targeted a creature")
	}
}

func TestGapB113PredatoryRampageBuffsYoursAndForcesTheirBlocks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Predatory Rampage", "Sorcery", gapB113Rampage, nil)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, mine); p != 5 || tough != 5 {
		t.Errorf("mine is %d/%d, want 5/5", p, tough)
	}
	if p, tough := ptOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("theirs is %d/%d, want unchanged 2/2", p, tough)
	}
	if !mkmB1MustBlock(t, g, theirs) {
		t.Error("their creature has no blocks-if-able requirement")
	}
	if mkmB1MustBlock(t, g, mine) {
		t.Error("my creature must block")
	}
	late := b12Creature(g, opp.ID, "Late", "Creature — Bear", 2, 2)
	if mkmB1MustBlock(t, g, late) {
		t.Error("a creature that arrived after resolution must block")
	}
}

func TestGapB113TheWandererPreventsNoncombatDamageToYouAndOtherPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	wanderer := pushWalkerForTest(g, opp.ID, "The Wanderer", gapB113Wanderer, 5)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	life := opp.Life
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", gapB113BoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("life %d -> %d, the Bolt should have been prevented", life, opp.Life)
	}
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", gapB113BoltOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("damage to another permanent you control was not prevented")
	}
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", gapB113BoltOracle, cardRefs(mine))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("damage to the other player's creature was prevented")
	}
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", gapB113BoltOracle, cardRefs(wanderer))
	passPriorityAroundTable(t, g)
	if got := loyaltyOf(g, wanderer); got != 2 {
		t.Errorf("The Wanderer's loyalty is %d, want 2: the static protects only other permanents", got)
	}
}

func TestGapB113TheWandererMinusTwoExilesAPowerFourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	w := pushWalkerForTest(g, me.ID, "The Wanderer", gapB113Wanderer, 5)
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	big := b12Creature(g, opp.ID, "Big", "Creature — Wurm", 4, 4)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(small)}); err == nil {
		t.Fatal("targeted a creature with power 3")
	}
	b16Activate(t, g, me.ID, w, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(big)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the 4/4 was not exiled")
	}
	if got := loyaltyOf(g, w); got != 3 {
		t.Errorf("loyalty %d, want 3", got)
	}
	if g.Exile == nil || !g.Exile.Contains(big) {
		t.Error("the 4/4 is not in exile")
	}
}

func TestGapB113NahiriPlusTwoMakesAKorSoldierAndMayAttachAnEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	n := pushWalkerForTest(g, me.ID, "Nahiri, the Lithomancer", gapB113Nahiri, 3)
	sword := b12Permanent(g, me.ID, "Sword", "Artifact — Equipment")
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, n, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	soldier := findBattlefieldByName(g, "Kor Soldier")
	if soldier == uuid.Nil {
		t.Fatal("no Kor Soldier token")
	}
	if c, _ := battlefieldCard(g, soldier); !c.IsCreature() || c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("Kor Soldier is %+v, want a 1/1 creature", c)
	}
	answerChooseCards(t, g, me.ID, sword)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, sword); c.AttachedTo.ID != soldier {
		t.Errorf("the Equipment is attached to %v, want the Soldier", c.AttachedTo.ID)
	}
	if got := loyaltyOf(g, n); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestGapB113NahiriPlusTwoWithNoEquipmentAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	n := pushWalkerForTest(g, me.ID, "Nahiri, the Lithomancer", gapB113Nahiri, 3)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, n, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Kor Soldier") == uuid.Nil {
		t.Fatal("no Kor Soldier token")
	}
	if latestChooseCardsFor(g, me.ID) != nil {
		t.Error("asked to attach with no Equipment to attach")
	}
}

func TestGapB113NahiriMinusTwoPutsAnEquipmentFromHandOrGraveyard(t *testing.T) {
	for _, zone := range []string{"hand", "graveyard"} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		n := pushWalkerForTest(g, me.ID, "Nahiri, the Lithomancer", gapB113Nahiri, 3)
		sword := uuid.New()
		card := game.Card{InstanceID: sword, Name: "Sword", TypeLine: "Artifact — Equipment", Owner: me.ID, Controller: me.ID}
		if zone == "hand" {
			me.Hand.PushTop(card)
		} else {
			me.Graveyard.PushTop(card)
		}
		notEquip := handCardForTest(me, "Bear", "Creature — Bear", "")
		advanceToMain(t, g)
		b16Activate(t, g, me.ID, n, 1, game.ActivateAbilityParams{})
		passPriorityAroundTable(t, g)
		pick := latestChooseCardsFor(g, me.ID)
		if pick == nil {
			t.Fatalf("%s: no pick prompt", zone)
		}
		for _, cid := range pick.ChooseCards {
			if cid == notEquip {
				t.Errorf("%s: a non-Equipment card is offered", zone)
			}
		}
		answerChooseCards(t, g, me.ID, sword)
		passPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(sword) {
			t.Errorf("%s: the Equipment did not enter", zone)
		}
		if got := loyaltyOf(g, n); got != 1 {
			t.Errorf("%s: loyalty %d, want 1", zone, got)
		}
	}
}

func TestGapB113NahiriMinusTenMakesTheStoneforgedBlade(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	n := pushWalkerForTest(g, me.ID, "Nahiri, the Lithomancer", gapB113Nahiri, 10)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, n, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	blade := findBattlefieldByName(g, "Stoneforged Blade")
	if blade == uuid.Nil {
		t.Fatal("no Stoneforged Blade token")
	}
	c, _ := battlefieldCard(g, blade)
	if !c.IsArtifact() || !c.HasSubtype("Equipment") || !c.IsColorless() {
		t.Errorf("token %q is not a colorless Equipment artifact", c.TypeLine)
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c, _ = battlefieldCard(g, blade); !game.HasKeyword(&c, "indestructible") {
		t.Error("the Blade is not indestructible")
	}
	// Equip {0}.
	if err := g.ActivateCatalogAbility(me.ID, blade, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("equip {0}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 7 || tough != 7 {
		t.Errorf("the equipped Bear is %d/%d, want 7/7", p, tough)
	}
	b, _ := battlefieldCard(g, bear)
	if !game.HasKeyword(&b, "double strike") {
		t.Error("the equipped Bear lacks double strike")
	}
}

func TestGapB113AnimalMagnetismOpponentChoosesACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, opp := range g.Seats[1:] {
		_ = opp
	}
	chooser := g.Seats[1]
	lib := seedLibrary(me, "Spell A", "Bear", "Spell B", "Wurm", "Spell C")
	setCard := func(id uuid.UUID, typeLine string) {
		for i := range me.Library.Cards {
			if me.Library.Cards[i].InstanceID == id {
				me.Library.Cards[i].TypeLine = typeLine
			}
		}
	}
	setCard(lib[1], "Creature — Bear")
	setCard(lib[3], "Creature — Wurm")
	castCatalogSpell(t, g, "Animal Magnetism", "Sorcery", gapB113Magnetism, nil)
	passPriorityAroundTable(t, g)
	answerChoosePlayer(t, g, me.ID, chooser)
	passPriorityAroundTable(t, g)
	pick := latestRevealPickFor(g, chooser.ID)
	if pick == nil {
		t.Fatal("the opponent was not asked to choose")
	}
	if len(pick.ChooseCards) != 2 {
		t.Errorf("%d candidates, want the two creature cards", len(pick.ChooseCards))
	}
	if err := g.ResolveRevealPick(pick.ID, chooser.ID, []uuid.UUID{lib[3]}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, lib[3]); !ok || c.Controller != me.ID {
		t.Error("the chosen Wurm is not on my battlefield")
	}
	for _, id := range []uuid.UUID{lib[0], lib[1], lib[2], lib[4]} {
		if !me.Graveyard.Contains(id) {
			t.Errorf("card %v is not in my graveyard", id)
		}
	}
}

func TestGapB113AnimalMagnetismWithNoCreatureSendsEverythingToTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := seedLibrary(me, "A", "B", "C", "D", "E")
	castCatalogSpell(t, g, "Animal Magnetism", "Sorcery", gapB113Magnetism, nil)
	passPriorityAroundTable(t, g)
	for _, id := range lib {
		if !me.Graveyard.Contains(id) {
			t.Errorf("card %v is not in my graveyard", id)
		}
	}
}

func gapB113Golem(p *game.Player, name, cost string, colors ...string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Artifact Creature — Golem", ManaCost: cost,
		Colors: colors, Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

func TestGapB113CalamityExilesEverythingCheaperThanTheRevealedCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	revealed := gapB113Golem(me, "Golem Five", "{5}")
	cheap := b12Creature(g, opp.ID, "Cheap", "Creature — Bear", 2, 2)
	cheap2 := b12Creature(g, me.ID, "Cheap Mine", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{cheap, cheap2} {
		g.WithWriteLock(func() { findBattlefieldCardForTest(g, id).ManaCost = "{4}" })
	}
	pricey := b12Creature(g, opp.ID, "Pricey", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, pricey).ManaCost = "{5}" })
	walker := pushWalkerForTest(g, opp.ID, "Walker", "", 3)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, walker).ManaCost = "{2}{W}" })
	b12PlayFromHand(t, g, "Calamity of the Titans", "Sorcery", gapB113Calamity,
		game.CastSpellParams{Strict: true, RevealIDs: []uuid.UUID{revealed}})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{cheap, cheap2, walker} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%v with mana value below 5 survived", id)
		}
	}
	if !g.Battlefield.Contains(pricey) {
		t.Error("a mana value 5 creature was exiled")
	}
	if !me.Hand.Contains(revealed) {
		t.Error("the revealed card left the hand")
	}
}

func TestGapB113CalamityRefusesAColoredOrMissingReveal(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	colored := gapB113Golem(me, "Colored", "{3}{W}", "W")
	noncreature := handCardForTest(me, "Rock", "Artifact", "")
	for name, params := range map[string]game.CastSpellParams{
		"no reveal":   {Strict: true},
		"colored":     {Strict: true, RevealIDs: []uuid.UUID{colored}},
		"noncreature": {Strict: true, RevealIDs: []uuid.UUID{noncreature}},
		"not in hand": {Strict: true, RevealIDs: []uuid.UUID{uuid.New()}},
	} {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Calamity of the Titans", TypeLine: "Sorcery", OracleID: gapB113Calamity, Owner: me.ID, Controller: me.ID})
		advanceToMain(t, g)
		if err := g.CastSpell(me.ID, id, params); err == nil {
			t.Errorf("%s: the cast was accepted", name)
		}
	}
}
