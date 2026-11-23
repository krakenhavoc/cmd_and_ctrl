package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Brightblade Stoat -------------------------------------------------------

func TestBlbBCBrightbladeStoatHasFirstStrikeAndLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	stoat := effectstest.BlbB1Push(g, me.ID, "Brightblade Stoat", "Creature — Weasel Soldier", blbBCStoatOracle, 2, 2)
	for _, kw := range []string{"first strike", "lifelink"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, stoat, kw) {
			t.Errorf("the Stoat lacks %s", kw)
		}
	}
}

// --- Charmed Sleep -----------------------------------------------------------

func TestBlbBCCharmedSleepTapsAndHoldsItsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	host := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	aura := effectstest.CastCatalogSpell(t, g, "Charmed Sleep", effectstest.AuraTypeLine, blbBCSleepOracle, effectstest.CardTarget(host))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.AttachmentHostOf(t, g, aura); got.Kind != game.TargetCard || got.ID != host {
		t.Fatalf("Charmed Sleep attached to %+v, want the bear", got)
	}
	if c, _ := effectstest.BattlefieldCard(g, host); !c.Tapped {
		t.Fatal("the entry trigger did not tap the enchanted creature")
	}
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats)) // opp's seat
	if c, _ := effectstest.BattlefieldCard(g, host); !c.Tapped {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
}

// --- Burrowguard Mentor ------------------------------------------------------

func TestBlbBCBurrowguardMentorCountsTheCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	mentor := effectstest.BlbB1Push(g, me.ID, "Burrowguard Mentor", "Creature — Rabbit Soldier", blbBCMentorOracle, 0, 0)
	if p, tough := effectstest.PtOf(t, g, mentor); p != 1 || tough != 1 {
		t.Fatalf("alone, the Mentor is %d/%d, want 1/1 (it counts itself)", p, tough)
	}
	effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.BlbB1Push(g, me.ID, "Other Bear", "Creature — Bear", "", 2, 2)
	effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	if p, tough := effectstest.PtOf(t, g, mentor); p != 3 || tough != 3 {
		t.Errorf("with two more creatures of mine, the Mentor is %d/%d, want 3/3", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, mentor, "trample") {
		t.Error("the Mentor lacks trample")
	}
}

// --- Brambleguard Veteran ----------------------------------------------------

func TestBlbBCBrambleguardVeteranPumpsRaccoonsOnExpendFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	vet := effectstest.BlbB1Push(g, me.ID, "Brambleguard Veteran", "Creature — Raccoon Warrior", blbBCVeteranOracle, 3, 4)
	raccoon := effectstest.BlbB1Push(g, me.ID, "Raccoon", "Creature — Raccoon", "", 2, 2)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.ToMainForCost(t, g)
	effectstest.CastPaid(t, g, me, "Three", "{2}{G}", "CCG")
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, vet); p != 3 {
		t.Fatalf("three mana spent: the Veteran's power is %d, want 3", p)
	}
	effectstest.CastPaid(t, g, me, "One", "{G}", "G")
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{vet, raccoon} {
		if !effectstest.HasKeywordOnBattlefield(t, g, id, "vigilance") {
			t.Errorf("%s: a Raccoon lacks vigilance after expend 4", id)
		}
	}
	if p, tough := effectstest.PtOf(t, g, vet); p != 4 || tough != 5 {
		t.Errorf("the Veteran is %d/%d, want 4/5", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, raccoon); p != 3 || tough != 3 {
		t.Errorf("the Raccoon is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 2 || effectstest.HasKeywordOnBattlefield(t, g, bear, "vigilance") {
		t.Errorf("a non-Raccoon got the bonus: %d/%d", p, tough)
	}
}

// --- Byway Barterer ----------------------------------------------------------

func TestBlbBCBywayBartererMayDiscardTheHandToDrawTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Byway Barterer", "Creature — Raccoon Rogue", blbBCBartererOracle, 3, 3)
	effectstest.ToMainForCost(t, g)
	effectstest.CastPaid(t, g, me, "Four", "{3}{R}", "CCCR")
	effectstest.PassPriorityAroundTable(t, g)
	hand, yard := me.Hand.Size(), me.Graveyard.Size()
	if hand == 0 {
		t.Fatal("the test needs cards in hand")
	}
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 2 {
		t.Errorf("hand %d, want 2", got)
	}
	if got := me.Graveyard.Size() - yard; got < hand {
		t.Errorf("%d cards reached the graveyard, want at least the hand of %d", got, hand)
	}
}

func TestBlbBCBywayBartererKeepsTheHandOnNo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Byway Barterer", "Creature — Raccoon Rogue", blbBCBartererOracle, 3, 3)
	effectstest.ToMainForCost(t, g)
	effectstest.CastPaid(t, g, me, "Four", "{3}{R}", "CCCR")
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d after a no, want %d", got, hand)
	}
}

// --- Bumbleflower's Sharepot -------------------------------------------------

func TestBlbBCSharepotMakesAFoodAndDestroysANonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	pot := effectstest.CastHoldCreature(t, g, "Bumbleflower's Sharepot", "Artifact", blbBCSharepotOracle, 0, 0)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Food"); n != 1 {
		t.Fatalf("%d Food tokens, want 1", n)
	}
	relic := effectstest.B12Permanent(g, opp.ID, "Their Relic", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Their Swamp", "Basic Land — Swamp")
	effectstest.FloatForTest(g, me, "CCCCC")
	if err := g.ActivateCatalogAbility(me.ID, pot, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(land)}); err == nil {
		t.Fatal("a land was accepted; the target is a nonland permanent")
	}
	effectstest.B16Activate(t, g, me.ID, pot, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(relic)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) || g.Battlefield.Contains(pot) {
		t.Error("the target is destroyed and the Sharepot sacrificed")
	}
}

func TestBlbBCSharepotActivatesOnlyAsASorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	pot := effectstest.BlbB1Push(g, me.ID, "Bumbleflower's Sharepot", "Artifact", blbBCSharepotOracle, 0, 0)
	relic := effectstest.B12Permanent(g, opp.ID, "Their Relic", "Artifact")
	// The harness parks on the draw step: not a main phase.
	effectstest.FloatForTest(g, me, "CCCCC")
	if err := g.ActivateCatalogAbility(me.ID, pot, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(relic)}); err == nil {
		t.Error("the ability was activated in the draw step; it is sorcery speed")
	}
}

// --- Carrot Cake -------------------------------------------------------------

func TestBlbBCCarrotCakeMakesARabbitOnEntryAndOnSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	cake := effectstest.CastHoldCreature(t, g, "Carrot Cake", "Artifact — Food", blbBCCakeOracle, 0, 0)
	effectstest.AnswerScryIfAsked(t, g, me)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Rabbit"); n != 1 {
		t.Fatalf("%d Rabbits after the Cake entered, want 1", n)
	}
	life := effectstest.LifeOf(g, me.ID)
	effectstest.FloatForTest(g, me, "CC")
	effectstest.B16Activate(t, g, me.ID, cake, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.AnswerScryIfAsked(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cake) {
		t.Fatal("the Cake was not sacrificed")
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Rabbit"); n != 2 {
		t.Errorf("%d Rabbits after the Cake was sacrificed, want 2", n)
	}
	if got := effectstest.LifeOf(g, me.ID); got != life+3 {
		t.Errorf("life %d, want %d", got, life+3)
	}
}

// --- Cindering Cutthroat -----------------------------------------------------

func TestBlbBCCinderingCutthroatEntersPlainWhenNoOpponentLostLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	cut := effectstest.CastHoldCreature(t, g, "Cindering Cutthroat", blbBCCutthroatTypeLn, blbBCCutthroatOracle, 3, 2)
	if p, tough := effectstest.PtOf(t, g, cut); p != 3 || tough != 2 {
		t.Errorf("the Cutthroat is %d/%d, want 3/2", p, tough)
	}
}

func TestBlbBCCinderingCutthroatEntersWithACounterAfterAnOpponentLostLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1); err != nil {
			t.Fatalf("lose life: %v", err)
		}
	})
	cut := effectstest.CastHoldCreature(t, g, "Cindering Cutthroat", blbBCCutthroatTypeLn, blbBCCutthroatOracle, 3, 2)
	if p, tough := effectstest.PtOf(t, g, cut); p != 4 || tough != 3 {
		t.Errorf("the Cutthroat is %d/%d, want 4/3", p, tough)
	}
}

func TestBlbBCCinderingCutthroatOnlyCountsAnOpponentsLoss(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -1); err != nil {
			t.Fatalf("lose life: %v", err)
		}
	})
	cut := effectstest.CastHoldCreature(t, g, "Cindering Cutthroat", blbBCCutthroatTypeLn, blbBCCutthroatOracle, 3, 2)
	if p, _ := effectstest.PtOf(t, g, cut); p != 3 {
		t.Errorf("my own life loss gave the Cutthroat a counter: power %d", p)
	}
}

func TestBlbBCCinderingCutthroatGainsMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	cut := effectstest.BlbB1Push(g, me.ID, "Cindering Cutthroat", blbBCCutthroatTypeLn, blbBCCutthroatOracle, 3, 2)
	effectstest.FloatForTest(g, me, "CR")
	effectstest.B16Activate(t, g, me.ID, cut, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, cut, "menace") {
		t.Error("the Cutthroat lacks menace after the activation")
	}
}

// --- Clifftop Lookout --------------------------------------------------------

func TestBlbBCClifftopLookoutPutsTheFirstLandOntoTheBattlefieldTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	land := effectstest.PlTop(me, "Forest", "Basic Land — Forest", "")
	a := effectstest.PlTop(me, "A", "Instant", "")
	b := effectstest.PlTop(me, "B", "Sorcery", "")
	lookout := effectstest.CastHoldCreature(t, g, "Clifftop Lookout", "Creature — Frog Scout", blbBCLookoutOracle, 1, 2)
	if !g.Battlefield.Contains(lookout) {
		t.Fatal("the Lookout is not on the battlefield")
	}
	if !g.Battlefield.Contains(land) || !effectstest.TappedOnBattlefield(t, g, land) {
		t.Fatal("the revealed land should enter tapped")
	}
	bottom := effectstest.PlBottomIDs(me, 2)
	if !bottom[a] || !bottom[b] {
		t.Error("the cards revealed above the land go to the bottom")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, lookout, "reach") {
		t.Error("the Lookout lacks reach")
	}
}

// --- Bonebind Orator ---------------------------------------------------------

func TestBlbBCBonebindOratorExilesItselfToReturnAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.ToMainForCost(t, g)
	orator := effectstest.PushCatalogGraveyardCard(me, "Bonebind Orator", "Creature — Squirrel Warlock Bard", blbBCOratorOracle, 2, 2)
	bear := effectstest.PushCatalogGraveyardCard(me, "Gy Bear", "Creature — Bear", "", 2, 2)
	spell := effectstest.PushCatalogGraveyardCard(me, "Gy Bolt", "Instant", "", 0, 0)
	effectstest.FloatForTest(g, me, "CCCB")
	for _, bad := range []uuid.UUID{orator, spell} {
		if err := g.ActivateCatalogAbility(me.ID, orator, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bad)}); err == nil {
			t.Fatalf("%s was accepted; the target is another creature card", bad)
		}
	}
	effectstest.B16Activate(t, g, me.ID, orator, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the other creature card did not return to hand")
	}
	if !g.Exile.Contains(orator) {
		t.Error("the Orator was not exiled")
	}
}

// --- Brazen Collector --------------------------------------------------------

func TestBlbBCBrazenCollectorAddsAKeptRedOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	collector := effectstest.BlbB1Push(g, me.ID, "Brazen Collector", "Creature — Raccoon Rogue", blbBCCollectorOracle, 2, 1)
	if !effectstest.HasKeywordOnBattlefield(t, g, collector, "first strike") {
		t.Error("the Collector lacks first strike")
	}
	effectstest.DeclareAttack(t, g, opp.ID, collector)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.ManaPool); got != 1 || me.ManaPool[0].Color != "R" {
		t.Fatalf("mana pool %+v after the attack, want one {R}", me.ManaPool)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	if got := len(me.ManaPool); got != 1 {
		t.Errorf("mana pool = %d after the combat steps, want the kept {R}", got)
	}
}

// --- Bonecache Overseer ------------------------------------------------------

func TestBlbBCBonecacheOverseerNeedsThreeCardsToLeaveTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	overseer := effectstest.BlbB1Push(g, me.ID, "Bonecache Overseer", blbBCOverseerTypeLine, blbBCOverseerOracle, 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, overseer, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the ability was activated with no card gone from the graveyard and no Food sacrificed")
	}
	for i := 0; i < 3; i++ {
		id := effectstest.PushCatalogGraveyardCard(me, "Dead", "Creature — Bear", "", 2, 2)
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
	hand, life := me.Hand.Size(), effectstest.LifeOf(g, me.ID)
	effectstest.B16Activate(t, g, me.ID, overseer, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	if got := effectstest.LifeOf(g, me.ID); got != life-1 {
		t.Errorf("life %d, want %d", got, life-1)
	}
}

func TestBlbBCBonecacheOverseerWorksAfterAFoodWasSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	overseer := effectstest.BlbB1Push(g, me.ID, "Bonecache Overseer", blbBCOverseerTypeLine, blbBCOverseerOracle, 1, 1)
	food := effectstest.PushToken(g, me.ID, effects.FoodToken())
	effectstest.FloatForTest(g, me, "CC")
	effectstest.B16Activate(t, g, me.ID, food, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, overseer, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

// --- Consumed by Greed -------------------------------------------------------

func TestBlbBCConsumedByGreedPromisedEdictsAndReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := effectstest.BlbB1Seats(g)
	big := effectstest.BlbB1Push(g, opp.ID, "Their Big", "Creature — Beast", "", 5, 5)
	small := effectstest.BlbB1Push(g, opp.ID, "Their Small", "Creature — Bear", "", 2, 2)
	dead := effectstest.PushCatalogGraveyardCard(me, "Gy Bear", "Creature — Bear", "", 2, 2)
	otherHand := other.Hand.Size()
	effectstest.MustCastWithGift(t, g, "Consumed by Greed", "Instant", blbBCGreedOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
		{Kind: game.TargetCard, ID: dead, Slot: 1},
	}, other.ID)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	a := effectstest.BlbB1Push(g, opp.ID, "Their Big", "Creature — Beast", "", 5, 5)
	b := effectstest.BlbB1Push(g, opp.ID, "Their Other Big", "Creature — Beast", "", 5, 5)
	small := effectstest.BlbB1Push(g, opp.ID, "Their Small", "Creature — Bear", "", 2, 2)
	dead := effectstest.PushCatalogGraveyardCard(me, "Gy Bear", "Creature — Bear", "", 2, 2)
	if _, err := effectstest.CastWithGift(t, g, "Consumed by Greed", "Instant", blbBCGreedOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: me.ID, Slot: 0},
	}, uuid.Nil); err == nil {
		t.Fatal("I was accepted as the target; it is a target opponent")
	}
	effectstest.MustCastWithGift(t, g, "Consumed by Greed", "Instant", blbBCGreedOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
	}, uuid.Nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Two creatures tie for the greatest power: the opponent chooses.
	effectstest.AnswerSacrifice(t, g, opp.ID, b)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) || g.Battlefield.Contains(b) || !g.Battlefield.Contains(small) {
		t.Error("the opponent's chosen 5-power creature should be the only one sacrificed")
	}
	if me.Hand.Contains(dead) {
		t.Error("without the gift nothing returns")
	}
}
