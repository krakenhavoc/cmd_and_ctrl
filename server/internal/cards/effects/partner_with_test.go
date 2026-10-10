package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// partner_with_test.go — #2142, the entry half of "Partner with
// [name]" (CR 702.124j): "When this permanent enters, target player
// may search their library for a card named [name], reveal it, put it
// into their hand, then shuffle." The deck half is internal/deck's
// partner_test.go.

const (
	pwFrodoOracle     = "58ee8c10-7be8-4889-944d-13cae0a4166d"
	pwSamOracle       = "2837be10-23ad-49d2-943e-b1e00edbbbd9"
	pwPirOracle       = "7683c2b2-a06f-4691-9cc5-1968dc032885"
	pwToothyOracle    = "41d6cce0-b852-4d0e-aee2-081df13dd9b8"
	pwLeyWeaverOracle = "24f1f0e9-8c9b-4f32-95ec-7af883bbeef4"
	pwLoreWeaverOrcl  = "040f6f11-7fbe-4635-b935-442e41f1e704"
	pwCaptainOracle   = "1675476c-25b1-46c2-81ce-fef9a2360bcc"
	pwRecruiterOracle = "33e9db5a-986d-4731-88c9-87de6f967db4"
)

// pwCastSam casts Sam, Loyal Attendant for the active seat, lets it
// resolve, and answers the trigger's target prompt with `target`. The
// trigger is then waiting on the stack.
func pwCastSam(t *testing.T, g *game.Game, target uuid.UUID) uuid.UUID {
	t.Helper()
	sam := castCatalogSpell(t, g, "Sam, Loyal Attendant", "Legendary Creature — Halfling Peasant", pwSamOracle, nil)
	passPriorityAroundTable(t, g)
	answerPickTargetPlayer(t, g, target)
	if triggerOnStack(g, sam) == nil {
		t.Fatal("Sam's partner-with trigger is not on the stack")
	}
	return sam
}

// pwSeedFrodo puts a Frodo, Adventurous Hobbit into a player's library
// among four other cards.
func pwSeedFrodo(p *game.Player) uuid.UUID {
	for i := 0; i < 2; i++ {
		pushLibraryCardForTest(p, game.Card{Name: "Filler", TypeLine: "Instant"})
	}
	frodo := pushLibraryCardForTest(p, game.Card{Name: "Frodo, Adventurous Hobbit",
		TypeLine: "Legendary Creature — Halfling Scout", OracleID: pwFrodoOracle})
	for i := 0; i < 2; i++ {
		pushLibraryCardForTest(p, game.Card{Name: "Filler", TypeLine: "Instant"})
	}
	return frodo
}

func pwSearchEvents(g *game.Game, actor uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventSearchLibrary && ev.Actor == actor {
			n++
		}
	}
	return n
}

func pwLibraryOrder(p *game.Player) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(p.Library.Cards))
	for _, c := range p.Library.Cards {
		out = append(out, c.InstanceID)
	}
	return out
}

// The trigger may target another player, and it is that player who is
// asked and who searches their own library. The card goes to their
// hand, revealed.
func TestPartnerWithTargetsAnotherPlayerWhoFindsTheCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	frodo := pwSeedFrodo(opp)
	pwCastSam(t, g, opp.ID)
	passPriorityAroundTable(t, g)

	if latestConfirmFor(g, me.ID) != nil {
		t.Fatal("the trigger's controller was asked; the target player decides")
	}
	answerMayChoice(t, g, opp.ID, true)
	answerSearchNamed(t, g, opp.ID, "Frodo, Adventurous Hobbit")
	passPriorityAroundTable(t, g)

	if !opp.Hand.Contains(frodo) {
		t.Fatal("Frodo did not reach the target player's hand")
	}
	if pwSearchEvents(g, opp.ID) != 1 {
		t.Errorf("search events by the target %d, want 1", pwSearchEvents(g, opp.ID))
	}
	if pwSearchEvents(g, me.ID) != 0 {
		t.Error("the trigger's controller searched")
	}
	var c game.Card
	for _, h := range opp.Hand.Cards {
		if h.InstanceID == frodo {
			c = h
		}
	}
	if !c.IsKnownTo(me.ID) {
		t.Error("the card found was not revealed to the table")
	}
}

// "May": a target who declines does not search, so nothing is
// revealed and the library is not shuffled.
func TestPartnerWithDeclineNeitherSearchesNorShuffles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	frodo := pwSeedFrodo(me)
	before := pwLibraryOrder(me)
	pwCastSam(t, g, me.ID)
	passPriorityAroundTable(t, g)

	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	if me.Hand.Contains(frodo) || !me.Library.Contains(frodo) {
		t.Fatal("a declined search still moved Frodo")
	}
	if pwSearchEvents(g, me.ID) != 0 {
		t.Error("a declined search still searched")
	}
	after := pwLibraryOrder(me)
	if len(after) != len(before) {
		t.Fatalf("library %d cards, want %d", len(after), len(before))
	}
	for i := range before {
		if before[i] != after[i] {
			t.Fatal("a declined search shuffled the library")
		}
	}
}

// With no card of that name in the library the search finds nothing,
// and the library is still searched and shuffled.
func TestPartnerWithNoCardFound(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushLibraryCardForTest(me, game.Card{Name: "Frodo Baggins", TypeLine: "Legendary Creature — Halfling Scout"})
	hand := me.Hand.Size()
	pwCastSam(t, g, me.ID)
	passPriorityAroundTable(t, g)

	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	if searchChoiceFor(g, me.ID) != nil {
		t.Fatal("a search with nothing to find opened a prompt")
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d: a card with a different name was found", me.Hand.Size(), hand)
	}
	if pwSearchEvents(g, me.ID) != 1 {
		t.Errorf("search events %d, want 1", pwSearchEvents(g, me.ID))
	}
}

// A card named the partner may still be left where it is (CR
// 701.23b): the search shuffles and nothing moves.
func TestPartnerWithFailToFind(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	frodo := pwSeedFrodo(me)
	pwCastSam(t, g, me.ID)
	passPriorityAroundTable(t, g)

	answerMayChoice(t, g, me.ID, true)
	answerSearchFailToFind(t, g, me.ID)
	passPriorityAroundTable(t, g)

	if !me.Library.Contains(frodo) {
		t.Fatal("failing to find still took Frodo")
	}
	if pwSearchEvents(g, me.ID) != 1 {
		t.Errorf("search events %d, want 1", pwSearchEvents(g, me.ID))
	}
}

// A partner-with trigger waiting on the stack is a restore point, and
// the restored trigger asks its target and finds the card.
func TestPartnerWithTriggerSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	frodo := pwSeedFrodo(opp)
	sam := pwCastSam(t, g, opp.ID)
	corpusRequireTriggeredStamp(t, triggerOnStack(g, sam), "own:")

	restored := restoreThroughJSON(t, g)
	var rOpp *game.Player
	for _, p := range restored.Seats {
		if p.ID == opp.ID {
			rOpp = p
		}
	}
	passPriorityAroundTable(t, restored)
	answerMayChoice(t, restored, rOpp.ID, true)
	answerSearchNamed(t, restored, rOpp.ID, "Frodo, Adventurous Hobbit")
	passPriorityAroundTable(t, restored)
	if !rOpp.Hand.Contains(frodo) {
		t.Fatal("the restored trigger did not find Frodo")
	}
}

// Every partner-with card in the catalog carries the trigger, naming
// its partner.
func TestPartnerWithCardsCarryTheTrigger(t *testing.T) {
	pairs := map[string]string{
		pwFrodoOracle:     "Sam, Loyal Attendant",
		pwSamOracle:       "Frodo, Adventurous Hobbit",
		pwPirOracle:       "Toothy, Imaginary Friend",
		pwToothyOracle:    "Pir, Imaginative Rascal",
		pwLeyWeaverOracle: "Lore Weaver",
		pwLoreWeaverOrcl:  "Ley Weaver",
		pwCaptainOracle:   "Blaring Recruiter",
		pwRecruiterOracle: "Blaring Captain",
	}
	for oracle, partner := range pairs {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		found := 0
		for _, tr := range spec.Triggered {
			if IsPartnerWithRow(tr) {
				found++
				if want := spec.Name + partnerWithKeyMark + partner; tr.Key != want {
					t.Errorf("%s: row %q, want %q", spec.Name, tr.Key, want)
				}
				if tr.Targets == nil {
					t.Errorf("%s: the trigger has no target", spec.Name)
				}
			}
		}
		if found != 1 {
			t.Errorf("%s: %d partner-with rows, want 1", spec.Name, found)
		}
	}
}

// --- the cards' other abilities ---------------------------------

// Frodo: with 3 life gained this turn the Ring tempts on attack, and
// on the second tempt, with Frodo the Ring-bearer, it draws.
func TestFrodoAdventurousHobbitTemptsAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	frodo := b11Push(g, me.ID, "Frodo, Adventurous Hobbit", "Legendary Creature — Halfling Scout", pwFrodoOracle, 1, 3)
	g.WithWriteLock(func() {
		_ = g.RingTemptsForEffect(me.ID, uuid.Nil, nil)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3)
	})
	if game.RingTemptCount(g, me.ID) != 1 || game.RingBearerOf(g, me.ID) != frodo {
		t.Fatalf("setup: tempts %d, bearer %v", game.RingTemptCount(g, me.ID), game.RingBearerOf(g, me.ID))
	}
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, frodo)
	passPriorityAroundTable(t, g)
	if game.RingTemptCount(g, me.ID) != 2 {
		t.Fatalf("tempts %d, want 2", game.RingTemptCount(g, me.ID))
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

// Frodo: without 3 life gained this turn, attacking does nothing.
func TestFrodoAdventurousHobbitNeedsThreeLifeGained(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	frodo := b11Push(g, me.ID, "Frodo, Adventurous Hobbit", "Legendary Creature — Halfling Scout", pwFrodoOracle, 1, 3)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2) })
	declareAttack(t, g, opp.ID, frodo)
	passPriorityAroundTable(t, g)
	if game.RingTemptCount(g, me.ID) != 0 {
		t.Errorf("tempts %d, want 0", game.RingTemptCount(g, me.ID))
	}
}

// Sam: a Food you control costs {1} less to activate. From a pool of
// two, one is left with Sam and none without.
func TestSamLoyalAttendantDiscountsFoods(t *testing.T) {
	for _, withSam := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		if withSam {
			b11Push(g, me.ID, "Sam, Loyal Attendant", "Legendary Creature — Halfling Peasant", pwSamOracle, 2, 4)
		}
		g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, FoodToken(), 1) })
		var food uuid.UUID
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.HasSubtype("Food") {
				food = c.InstanceID
			}
		}
		advanceTo(t, g, game.StepPrecombatMain)
		floatForTest(g, me, "CC")
		if err := g.ActivateCatalogAbility(me.ID, food, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate the Food: %v", err)
		}
		want := 0
		if withSam {
			want = 1
		}
		if len(me.ManaPool) != want {
			t.Errorf("Sam %v: %d mana left, want %d", withSam, len(me.ManaPool), want)
		}
	}
}

// Toothy grows with each card drawn and draws that many as it leaves.
func TestToothyDrawsForItsCountersAsItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toothy := b11Push(g, me.ID, "Toothy, Imaginary Friend", "Legendary Creature — Illusion", pwToothyOracle, 1, 1)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 2) })
	passPriorityAroundTable(t, g)
	if n := countersOn(g, toothy, game.CounterPlusOne); n != 2 {
		t.Fatalf("Toothy has %d counters, want 2", n)
	}
	hand := me.Hand.Size()
	b25Destroy(g, toothy)
	passPriorityAroundTable(t, g)
	// The two draws trigger Toothy again only while it is on the
	// battlefield; it is gone, so the count is exactly two.
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+2)
	}
}

// Ley Weaver untaps two target lands.
func TestLeyWeaverUntapsTwoLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	weaver := b11Push(g, me.ID, "Ley Weaver", "Creature — Human Druid", pwLeyWeaverOracle, 2, 2)
	a := b11Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	b := b11Push(g, me.ID, "Island", "Basic Land — Island", "", 0, 0)
	for i := range g.Battlefield.Cards {
		if id := g.Battlefield.Cards[i].InstanceID; id == a || id == b {
			g.Battlefield.Cards[i].Tapped = true
		}
	}
	advanceTo(t, g, game.StepPrecombatMain)
	poolActivate(t, g, me, weaver, "", cardRef(a), cardRef(b))
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if (c.InstanceID == a || c.InstanceID == b) && c.Tapped {
			t.Errorf("%s is still tapped", c.Name)
		}
	}
}

// Lore Weaver: target player draws two.
func TestLoreWeaverTargetPlayerDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	weaver := b11Push(g, me.ID, "Lore Weaver", "Creature — Human Wizard", pwLoreWeaverOrcl, 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	hand := opp.Hand.Size()
	poolActivate(t, g, me, weaver, "CCCCCUU", game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d", opp.Hand.Size(), hand+2)
	}
}

// Blaring Recruiter makes a 1/1 Warrior; Blaring Captain's attack
// gives every attacking Warrior +1/+1.
func TestBlaringPairMakesAndPumpsWarriors(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	captain := b11Push(g, me.ID, "Blaring Captain", "Creature — Azra Warrior", pwCaptainOracle, 2, 2)
	recruiter := b11Push(g, me.ID, "Blaring Recruiter", "Creature — Elf Warrior", pwRecruiterOracle, 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	poolActivate(t, g, me, recruiter, "CCW")
	passPriorityAroundTable(t, g)
	warriors := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() && c.HasSubtype("Warrior") {
			warriors++
		}
	}
	if warriors != 1 {
		t.Fatalf("%d Warrior tokens, want 1", warriors)
	}
	declareAttack(t, g, opp.ID, captain, recruiter)
	passPriorityAroundTable(t, g)
	if p := powerOf(t, g, captain); p != 3 {
		t.Errorf("Captain power %d, want 3", p)
	}
	if p := powerOf(t, g, recruiter); p != 3 {
		t.Errorf("Recruiter power %d, want 3", p)
	}
}
