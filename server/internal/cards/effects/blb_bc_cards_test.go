package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blb_bc_cards_test.go — Bloomburrow, B–C (#4467). Each test plays the
// real catalog entry and asserts what its text says.

const (
	blbBCStoatOracle      = "f1d2f437-d474-4617-aea6-102daa5e5a15"
	blbBCSleepOracle      = "cb33edaf-b604-48b8-8b9c-d6e462a8bebb"
	blbBCMentorOracle     = "f293da27-e42d-4e87-8b61-d3c27e4c6d71"
	blbBCVeteranOracle    = "bc9c7e1a-5155-431a-be21-fcbd18718980"
	blbBCBartererOracle   = "a864b032-6e1a-4211-ad6e-54a5a5c7321b"
	blbBCSharepotOracle   = "6e152529-64b4-472c-8e50-99095af6af99"
	blbBCCakeOracle       = "5fe3aabd-3271-41bf-ae50-f6efaba7c4ae"
	blbBCCutthroatOracle  = "46e8bec0-1a09-4e6e-b78a-7ba161686ca9"
	blbBCLookoutOracle    = "446db5a4-21b0-44c1-b12b-d365313dd67a"
	blbBCOratorOracle     = "c79cf0e5-52dd-4af3-ab06-6cdaab526c23"
	blbBCCollectorOracle  = "f93dc9cc-3dd2-4e2b-b972-217ab9d89dfe"
	blbBCOverseerOracle   = "be3c7e4c-81fd-49e7-8fb5-5eb83994dafd"
	blbBCGreedOracle      = "b548578c-4910-402c-8e78-498924e0a9be"
	blbBCCutthroatTypeLn  = "Creature — Lizard Assassin"
	blbBCOverseerTypeLine = "Creature — Squirrel Warlock"
)

func TestBlbBCEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		blbBCStoatOracle:     "Brightblade Stoat",
		blbBCSleepOracle:     "Charmed Sleep",
		blbBCMentorOracle:    "Burrowguard Mentor",
		blbBCVeteranOracle:   "Brambleguard Veteran",
		blbBCBartererOracle:  "Byway Barterer",
		blbBCSharepotOracle:  "Bumbleflower's Sharepot",
		blbBCCakeOracle:      "Carrot Cake",
		blbBCCutthroatOracle: "Cindering Cutthroat",
		blbBCLookoutOracle:   "Clifftop Lookout",
		blbBCOratorOracle:    "Bonebind Orator",
		blbBCCollectorOracle: "Brazen Collector",
		blbBCOverseerOracle:  "Bonecache Overseer",
		blbBCGreedOracle:     "Consumed by Greed",
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

// --- Brightblade Stoat -------------------------------------------------------

func TestBlbBCBrightbladeStoatHasFirstStrikeAndLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	stoat := blbB1Push(g, me.ID, "Brightblade Stoat", "Creature — Weasel Soldier", blbBCStoatOracle, 2, 2)
	for _, kw := range []string{"first strike", "lifelink"} {
		if !hasKeywordOnBattlefield(t, g, stoat, kw) {
			t.Errorf("the Stoat lacks %s", kw)
		}
	}
}

// --- Charmed Sleep -----------------------------------------------------------

func TestBlbBCCharmedSleepTapsAndHoldsItsCreature(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	host := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	aura := castCatalogSpell(t, g, "Charmed Sleep", auraTypeLine, blbBCSleepOracle, cardTarget(host))
	passPriorityAroundTable(t, g)
	if got := attachmentHostOf(t, g, aura); got.Kind != game.TargetCard || got.ID != host {
		t.Fatalf("Charmed Sleep attached to %+v, want the bear", got)
	}
	if c, _ := battlefieldCard(g, host); !c.Tapped {
		t.Fatal("the entry trigger did not tap the enchanted creature")
	}
	advanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats)) // opp's seat
	if c, _ := battlefieldCard(g, host); !c.Tapped {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
}

// --- Burrowguard Mentor ------------------------------------------------------

func TestBlbBCBurrowguardMentorCountsTheCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	mentor := blbB1Push(g, me.ID, "Burrowguard Mentor", "Creature — Rabbit Soldier", blbBCMentorOracle, 0, 0)
	if p, tough := ptOf(t, g, mentor); p != 1 || tough != 1 {
		t.Fatalf("alone, the Mentor is %d/%d, want 1/1 (it counts itself)", p, tough)
	}
	blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	blbB1Push(g, me.ID, "Other Bear", "Creature — Bear", "", 2, 2)
	blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	if p, tough := ptOf(t, g, mentor); p != 3 || tough != 3 {
		t.Errorf("with two more creatures of mine, the Mentor is %d/%d, want 3/3", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, mentor, "trample") {
		t.Error("the Mentor lacks trample")
	}
}

// --- Brambleguard Veteran ----------------------------------------------------

func TestBlbBCBrambleguardVeteranPumpsRaccoonsOnExpendFour(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	vet := blbB1Push(g, me.ID, "Brambleguard Veteran", "Creature — Raccoon Warrior", blbBCVeteranOracle, 3, 4)
	raccoon := blbB1Push(g, me.ID, "Raccoon", "Creature — Raccoon", "", 2, 2)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	toMainForCost(t, g)
	castPaid(t, g, me, "Three", "{2}{G}", "CCG")
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, vet); p != 3 {
		t.Fatalf("three mana spent: the Veteran's power is %d, want 3", p)
	}
	castPaid(t, g, me, "One", "{G}", "G")
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{vet, raccoon} {
		if !hasKeywordOnBattlefield(t, g, id, "vigilance") {
			t.Errorf("%s: a Raccoon lacks vigilance after expend 4", id)
		}
	}
	if p, tough := ptOf(t, g, vet); p != 4 || tough != 5 {
		t.Errorf("the Veteran is %d/%d, want 4/5", p, tough)
	}
	if p, tough := ptOf(t, g, raccoon); p != 3 || tough != 3 {
		t.Errorf("the Raccoon is %d/%d, want 3/3", p, tough)
	}
	if p, tough := ptOf(t, g, bear); p != 2 || tough != 2 || hasKeywordOnBattlefield(t, g, bear, "vigilance") {
		t.Errorf("a non-Raccoon got the bonus: %d/%d", p, tough)
	}
}

// --- Byway Barterer ----------------------------------------------------------

func TestBlbBCBywayBartererMayDiscardTheHandToDrawTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Byway Barterer", "Creature — Raccoon Rogue", blbBCBartererOracle, 3, 3)
	toMainForCost(t, g)
	castPaid(t, g, me, "Four", "{3}{R}", "CCCR")
	passPriorityAroundTable(t, g)
	hand, yard := me.Hand.Size(), me.Graveyard.Size()
	if hand == 0 {
		t.Fatal("the test needs cards in hand")
	}
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 2 {
		t.Errorf("hand %d, want 2", got)
	}
	if got := me.Graveyard.Size() - yard; got < hand {
		t.Errorf("%d cards reached the graveyard, want at least the hand of %d", got, hand)
	}
}

func TestBlbBCBywayBartererKeepsTheHandOnNo(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Byway Barterer", "Creature — Raccoon Rogue", blbBCBartererOracle, 3, 3)
	toMainForCost(t, g)
	castPaid(t, g, me, "Four", "{3}{R}", "CCCR")
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d after a no, want %d", got, hand)
	}
}

// --- Bumbleflower's Sharepot -------------------------------------------------

func TestBlbBCSharepotMakesAFoodAndDestroysANonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	pot := castHoldCreature(t, g, "Bumbleflower's Sharepot", "Artifact", blbBCSharepotOracle, 0, 0)
	if n, _ := countTokensNamed(g, me.ID, "Food"); n != 1 {
		t.Fatalf("%d Food tokens, want 1", n)
	}
	relic := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	land := b12Permanent(g, opp.ID, "Their Swamp", "Basic Land — Swamp")
	floatForTest(g, me, "CCCCC")
	if err := g.ActivateCatalogAbility(me.ID, pot, 0, game.ActivateAbilityParams{Targets: cardTarget(land)}); err == nil {
		t.Fatal("a land was accepted; the target is a nonland permanent")
	}
	b16Activate(t, g, me.ID, pot, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(relic)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) || g.Battlefield.Contains(pot) {
		t.Error("the target is destroyed and the Sharepot sacrificed")
	}
}

func TestBlbBCSharepotActivatesOnlyAsASorcery(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	pot := blbB1Push(g, me.ID, "Bumbleflower's Sharepot", "Artifact", blbBCSharepotOracle, 0, 0)
	relic := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	// The harness parks on the draw step: not a main phase.
	floatForTest(g, me, "CCCCC")
	if err := g.ActivateCatalogAbility(me.ID, pot, 0, game.ActivateAbilityParams{Targets: cardTarget(relic)}); err == nil {
		t.Error("the ability was activated in the draw step; it is sorcery speed")
	}
}

// --- Carrot Cake -------------------------------------------------------------

func TestBlbBCCarrotCakeMakesARabbitOnEntryAndOnSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	cake := castHoldCreature(t, g, "Carrot Cake", "Artifact — Food", blbBCCakeOracle, 0, 0)
	answerScryIfAsked(t, g, me)
	if n, _ := countTokensNamed(g, me.ID, "Rabbit"); n != 1 {
		t.Fatalf("%d Rabbits after the Cake entered, want 1", n)
	}
	life := lifeOf(g, me.ID)
	floatForTest(g, me, "CC")
	b16Activate(t, g, me.ID, cake, 0, game.ActivateAbilityParams{Strict: true})
	answerScryIfAsked(t, g, me)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cake) {
		t.Fatal("the Cake was not sacrificed")
	}
	if n, _ := countTokensNamed(g, me.ID, "Rabbit"); n != 2 {
		t.Errorf("%d Rabbits after the Cake was sacrificed, want 2", n)
	}
	if got := lifeOf(g, me.ID); got != life+3 {
		t.Errorf("life %d, want %d", got, life+3)
	}
}

// --- Cindering Cutthroat -----------------------------------------------------

func TestBlbBCCinderingCutthroatEntersPlainWhenNoOpponentLostLife(t *testing.T) {
	g := newCatalogGame(t)
	cut := castHoldCreature(t, g, "Cindering Cutthroat", blbBCCutthroatTypeLn, blbBCCutthroatOracle, 3, 2)
	if p, tough := ptOf(t, g, cut); p != 3 || tough != 2 {
		t.Errorf("the Cutthroat is %d/%d, want 3/2", p, tough)
	}
}

func TestBlbBCCinderingCutthroatEntersWithACounterAfterAnOpponentLostLife(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1); err != nil {
			t.Fatalf("lose life: %v", err)
		}
	})
	cut := castHoldCreature(t, g, "Cindering Cutthroat", blbBCCutthroatTypeLn, blbBCCutthroatOracle, 3, 2)
	if p, tough := ptOf(t, g, cut); p != 4 || tough != 3 {
		t.Errorf("the Cutthroat is %d/%d, want 4/3", p, tough)
	}
}

func TestBlbBCCinderingCutthroatOnlyCountsAnOpponentsLoss(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -1); err != nil {
			t.Fatalf("lose life: %v", err)
		}
	})
	cut := castHoldCreature(t, g, "Cindering Cutthroat", blbBCCutthroatTypeLn, blbBCCutthroatOracle, 3, 2)
	if p, _ := ptOf(t, g, cut); p != 3 {
		t.Errorf("my own life loss gave the Cutthroat a counter: power %d", p)
	}
}

func TestBlbBCCinderingCutthroatGainsMenace(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	cut := blbB1Push(g, me.ID, "Cindering Cutthroat", blbBCCutthroatTypeLn, blbBCCutthroatOracle, 3, 2)
	floatForTest(g, me, "CR")
	b16Activate(t, g, me.ID, cut, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, cut, "menace") {
		t.Error("the Cutthroat lacks menace after the activation")
	}
}

// --- Clifftop Lookout --------------------------------------------------------

func TestBlbBCClifftopLookoutPutsTheFirstLandOntoTheBattlefieldTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	land := plTop(me, "Forest", "Basic Land — Forest", "")
	a := plTop(me, "A", "Instant", "")
	b := plTop(me, "B", "Sorcery", "")
	lookout := castHoldCreature(t, g, "Clifftop Lookout", "Creature — Frog Scout", blbBCLookoutOracle, 1, 2)
	if !g.Battlefield.Contains(lookout) {
		t.Fatal("the Lookout is not on the battlefield")
	}
	if !g.Battlefield.Contains(land) || !tappedOnBattlefield(t, g, land) {
		t.Fatal("the revealed land should enter tapped")
	}
	bottom := plBottomIDs(me, 2)
	if !bottom[a] || !bottom[b] {
		t.Error("the cards revealed above the land go to the bottom")
	}
	if !hasKeywordOnBattlefield(t, g, lookout, "reach") {
		t.Error("the Lookout lacks reach")
	}
}

// --- Bonebind Orator ---------------------------------------------------------

func TestBlbBCBonebindOratorExilesItselfToReturnAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	toMainForCost(t, g)
	orator := pushCatalogGraveyardCard(me, "Bonebind Orator", "Creature — Squirrel Warlock Bard", blbBCOratorOracle, 2, 2)
	bear := pushCatalogGraveyardCard(me, "Gy Bear", "Creature — Bear", "", 2, 2)
	spell := pushCatalogGraveyardCard(me, "Gy Bolt", "Instant", "", 0, 0)
	floatForTest(g, me, "CCCB")
	for _, bad := range []uuid.UUID{orator, spell} {
		if err := g.ActivateCatalogAbility(me.ID, orator, 0, game.ActivateAbilityParams{Targets: cardTarget(bad)}); err == nil {
			t.Fatalf("%s was accepted; the target is another creature card", bad)
		}
	}
	b16Activate(t, g, me.ID, orator, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(bear)})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the other creature card did not return to hand")
	}
	if !g.Exile.Contains(orator) {
		t.Error("the Orator was not exiled")
	}
}

// --- Brazen Collector --------------------------------------------------------

func TestBlbBCBrazenCollectorAddsAKeptRedOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	collector := blbB1Push(g, me.ID, "Brazen Collector", "Creature — Raccoon Rogue", blbBCCollectorOracle, 2, 1)
	if !hasKeywordOnBattlefield(t, g, collector, "first strike") {
		t.Error("the Collector lacks first strike")
	}
	declareAttack(t, g, opp.ID, collector)
	passPriorityAroundTable(t, g)
	if got := len(me.ManaPool); got != 1 || me.ManaPool[0].Color != "R" {
		t.Fatalf("mana pool %+v after the attack, want one {R}", me.ManaPool)
	}
	advanceTo(t, g, game.StepCombatDamage)
	if got := len(me.ManaPool); got != 1 {
		t.Errorf("mana pool = %d after the combat steps, want the kept {R}", got)
	}
}

// --- Bonecache Overseer ------------------------------------------------------

func TestBlbBCBonecacheOverseerNeedsThreeCardsToLeaveTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	overseer := blbB1Push(g, me.ID, "Bonecache Overseer", blbBCOverseerTypeLine, blbBCOverseerOracle, 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, overseer, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the ability was activated with no card gone from the graveyard and no Food sacrificed")
	}
	for i := 0; i < 3; i++ {
		id := pushCatalogGraveyardCard(me, "Dead", "Creature — Bear", "", 2, 2)
		g.WithWriteLock(func() {
			if err := g.ExileCardForEffect(id); err != nil {
				t.Fatalf("exile: %v", err)
			}
		})
		if i == 1 {
			if err := g.ActivateCatalogAbility(me.ID, overseer, 0, game.ActivateAbilityParams{}); err == nil {
				t.Fatal("the ability was activated after two cards left the graveyard")
			}
		}
	}
	hand, life := me.Hand.Size(), lifeOf(g, me.ID)
	b16Activate(t, g, me.ID, overseer, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	if got := lifeOf(g, me.ID); got != life-1 {
		t.Errorf("life %d, want %d", got, life-1)
	}
}

func TestBlbBCBonecacheOverseerWorksAfterAFoodWasSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	overseer := blbB1Push(g, me.ID, "Bonecache Overseer", blbBCOverseerTypeLine, blbBCOverseerOracle, 1, 1)
	food := pushToken(g, me.ID, FoodToken())
	floatForTest(g, me, "CC")
	b16Activate(t, g, me.ID, food, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, overseer, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

// --- Consumed by Greed -------------------------------------------------------

func TestBlbBCConsumedByGreedPromisedEdictsAndReturns(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := blbB1Seats(g)
	big := blbB1Push(g, opp.ID, "Their Big", "Creature — Beast", "", 5, 5)
	small := blbB1Push(g, opp.ID, "Their Small", "Creature — Bear", "", 2, 2)
	dead := pushCatalogGraveyardCard(me, "Gy Bear", "Creature — Bear", "", 2, 2)
	otherHand := other.Hand.Size()
	mustCastWithGift(t, g, "Consumed by Greed", "Instant", blbBCGreedOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
		{Kind: game.TargetCard, ID: dead, Slot: 1},
	}, other.ID)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) || !g.Battlefield.Contains(small) {
		t.Error("the target opponent should sacrifice the creature with the greatest power")
	}
	if !me.Hand.Contains(dead) {
		t.Error("with the gift promised, the creature card returns to hand")
	}
	if got := other.Hand.Size(); got != otherHand+1 {
		t.Errorf("the gift's opponent has %d cards, want %d", got, otherHand+1)
	}
}

func TestBlbBCConsumedByGreedUnpromisedOnlyEdicts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	a := blbB1Push(g, opp.ID, "Their Big", "Creature — Beast", "", 5, 5)
	b := blbB1Push(g, opp.ID, "Their Other Big", "Creature — Beast", "", 5, 5)
	small := blbB1Push(g, opp.ID, "Their Small", "Creature — Bear", "", 2, 2)
	dead := pushCatalogGraveyardCard(me, "Gy Bear", "Creature — Bear", "", 2, 2)
	if _, err := castWithGift(t, g, "Consumed by Greed", "Instant", blbBCGreedOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: me.ID, Slot: 0},
	}, uuid.Nil); err == nil {
		t.Fatal("I was accepted as the target; it is a target opponent")
	}
	mustCastWithGift(t, g, "Consumed by Greed", "Instant", blbBCGreedOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
	}, uuid.Nil)
	passPriorityAroundTable(t, g)
	// Two creatures tie for the greatest power: the opponent chooses.
	answerSacrifice(t, g, opp.ID, b)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) || g.Battlefield.Contains(b) || !g.Battlefield.Contains(small) {
		t.Error("the opponent's chosen 5-power creature should be the only one sacrificed")
	}
	if me.Hand.Contains(dead) {
		t.Error("without the gift nothing returns")
	}
}
