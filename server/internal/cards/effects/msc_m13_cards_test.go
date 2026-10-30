package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_m13_cards_test.go — Marvel Super Heroes Commander, M–Z batch 13
// (#4455). Each test plays the real catalog entry and asserts what its text
// says.

const (
	mscM13MattMurdockOracle = "e64efd0d-3de3-42e2-949e-4a152d08def8"
	mscM13MsMarvelOracle    = "26a1b36c-3c27-4eb4-b85f-63459653b773"
	mscM13StickOracle       = "e7bd3882-a0f2-4706-80e7-1d3ec455801f"
	mscM13StarksOracle      = "913e6ff0-b522-45a3-9a1e-4a87a515cc1d"
	mscM13ImmortalOracle    = "f016ad56-4f73-4091-a1c2-937782642ecb"
	mscM13WhirlwindOracle   = "72eabfcc-61b6-4a07-9a9c-b40a9c05cf66"
	mscM13StormOracle       = "d8e585c5-1de3-4558-8a2f-0d9fb07edad6"
	mscM13FantasticOracle   = "a8011dce-17fb-46e1-8297-81a4acccbd05"
	mscM13ThingOracle       = "c8de07d4-9bc1-4bfe-b5a4-0a58e05af022"
	mscM13NicoOracle        = "5303b3f6-970c-47d8-977e-2d6679206ba5"
	mscM13WillieOracle      = "b08633e9-7b8f-4f69-a3b7-8d982b17574c"
	mscM13QuicksilverOracle = "cc20f910-03e9-4019-a62a-2e713ddb3c06"
	mscM13HeroType          = "Legendary Creature — Human Hero"
)

func TestMscM13EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mscM13MattMurdockOracle: "Matt Murdock, Justice Seeker",
		mscM13MsMarvelOracle:    "Ms. Marvel, Elastic Ally",
		mscM13StickOracle:       "Stick, Fearless Mentor",
		mscM13StarksOracle:      "Stark's Ingenuity",
		mscM13ImmortalOracle:    "Mister Immortal",
		mscM13WhirlwindOracle:   "Whirlwind, Killer Cyclone",
		mscM13StormOracle:       "Storm, Queen of Wakanda",
		mscM13FantasticOracle:   "Mister Fantastic",
		mscM13ThingOracle:       "The Thing",
		mscM13NicoOracle:        "Nico Minoru, Runaway",
		mscM13WillieOracle:      "Willie Lumpkin, Postman",
		mscM13QuicksilverOracle: "Quicksilver, Speedster",
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

// mscM13OpenPickTargetCards is the cards the chooser's newest pick_target
// prompt offers, nil when none is open.
func mscM13OpenPickTargetCards(g *game.Game, chooser uuid.UUID) []uuid.UUID {
	if p := latestPickTarget(g, chooser); p != nil {
		return p.PickTargetCards
	}
	return nil
}

// mscM13TriggersFrom counts the triggers of `source` that exist: on the
// stack, or waiting on a target pick.
func mscM13TriggersFrom(g *game.Game, source uuid.UUID) int {
	n := triggersOnStackFrom(g, source)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget && c.Source == source {
			n++
		}
	}
	return n
}

// mscM13MarkEntered records an entry for each permanent, as if it had just
// come onto the battlefield (the "entered this turn" cell is fed by the ETB
// event; a card pushed straight onto the battlefield has none).
func mscM13MarkEntered(g *game.Game, ids ...uuid.UUID) {
	g.WithWriteLock(func() {
		for _, id := range ids {
			g.EmitEvent(game.Event{Kind: game.EventETB, CardID: id})
		}
	})
}

// mscM13PassUntilPickTarget passes priority until the chooser owes a target.
func mscM13PassUntilPickTarget(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	for i := 0; i < 10 && g.Stack.Size() > 0; i++ {
		if latestPickTarget(g, chooser) != nil {
			return
		}
		if err := g.PassPriority(); err != nil {
			break
		}
	}
}

// --- Matt Murdock, Justice Seeker ---------------------------------------------

func mscM13MattAtCombat(t *testing.T, g *game.Game) (matt, bear uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	matt = pushCatalogPermanent(g, me.ID, "Matt Murdock, Justice Seeker", "Legendary Creature — Human Advisor Hero", mscM13MattMurdockOracle, false)
	bear = b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	// The pool empties at every step change, so the {1} is floated now.
	b06AddMana(me, "G")
	passPriorityAroundTable(t, g)
	return matt, bear
}

func TestMscM13MattMurdockPaysOneForACounterOnACreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	_, bear := mscM13MattAtCombat(t, g)

	answerPayUnless(t, g, me.ID, true)
	if cards := mscM13OpenPickTargetCards(g, me.ID); len(cards) > 0 {
		for _, id := range cards {
			if id == theirs {
				t.Fatalf("an opponent's creature was offered: %v", cards)
			}
		}
		answerPickTarget(t, g, bear)
	}
	passPriorityAroundTable(t, g)

	if n := hobA8Counters(t, g, bear); n != 1 {
		t.Errorf("Bear has %d counters, want 1", n)
	}
	if n := hobA8Counters(t, g, theirs); n != 0 {
		t.Errorf("their Bear has %d counters, want 0", n)
	}
	if len(poolColors(me)) != 0 {
		t.Errorf("pool = %v after paying {1}", poolColors(me))
	}
}

func TestMscM13MattMurdockDecliningThePaymentPutsNoCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	_, bear := mscM13MattAtCombat(t, g)

	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	if n := hobA8Counters(t, g, bear); n != 0 {
		t.Errorf("Bear has %d counters after declining", n)
	}
	if len(poolColors(me)) != 1 {
		t.Errorf("the {1} was spent anyway: pool = %v", poolColors(me))
	}
}

func TestMscM13MattMurdockGivesWardOnlyToYourCreaturesWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Matt Murdock, Justice Seeker", "Legendary Creature — Human Advisor Hero", mscM13MattMurdockOracle, false)
	marked := b12Creature(g, me.ID, "Marked Bear", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{marked, theirs} {
		if err := g.AddCounterBy(me.ID, id, game.CounterPlusOne, 1); err != nil {
			t.Fatalf("AddCounterBy: %v", err)
		}
	}
	advanceToMain(t, g)

	castAtWardedCreature(t, g, opp, plain)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID) && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if hasPayUnlessFor(g, opp.ID) {
		t.Fatal("a creature with no counter has ward")
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(plain) {
		t.Error("the unwarded creature survived its Doom Blade")
	}

	castAtWardedCreature(t, g, opp, marked)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("a creature with a counter on it has no ward {1}")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(marked) {
		t.Error("a declined ward did not counter the spell")
	}

	castAtWardedCreature(t, g, me, theirs)
	for i := 0; i < 8 && !hasPayUnlessFor(g, me.ID) && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if hasPayUnlessFor(g, me.ID) {
		t.Error("an opponent's creature with a counter has ward from Matt Murdock")
	}
}

// --- Ms. Marvel, Elastic Ally -------------------------------------------------

func TestMscM13MsMarvelPumpsTargetCreatureWhenShePlays(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Ms. Marvel, Elastic Ally", "Legendary Creature — Mutant Inhuman Hero", mscM13MsMarvelOracle, nil)
	mscM13PassUntilPickTarget(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("Bear power %d, want 4 (2 + 2)", got)
	}
	powerUpPassTo(t, g, 1)
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("Bear power %d the next turn, want 2: the pump is until end of turn", got)
	}
}

func TestMscM13MsMarvelDrawsOnceForAStrongerCreatureThatConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Ms. Marvel, Elastic Ally", "Legendary Creature — Mutant Inhuman Hero", mscM13MsMarvelOracle, false)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.AddCounterBy(me.ID, id, game.CounterPlusOne, 1); err != nil {
			t.Fatalf("AddCounterBy: %v", err)
		}
	}
	handBefore := me.Hand.Size()

	attackWith(t, g, opp.ID, plain, a, b)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("drew %d cards, want 1: two stronger creatures connected but the ability triggers once each turn", got)
	}
}

func TestMscM13MsMarvelDrawsNothingForAnUnboostedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Ms. Marvel, Elastic Ally", "Legendary Creature — Mutant Inhuman Hero", mscM13MsMarvelOracle, false)
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	handBefore := me.Hand.Size()

	attackWith(t, g, opp.ID, plain)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size() - handBefore; got != 0 {
		t.Errorf("drew %d cards for a creature at its base power", got)
	}
}

func TestMscM13MsMarvelDrawsNothingWhenTheStrongerCreatureIsBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Ms. Marvel, Elastic Ally", "Legendary Creature — Mutant Inhuman Hero", mscM13MsMarvelOracle, false)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	if err := g.AddCounterBy(me.ID, a, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounterBy: %v", err)
	}
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	handBefore := me.Hand.Size()

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, a); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size() - handBefore; got != 0 {
		t.Errorf("drew %d cards: a blocked creature dealt no damage to a player", got)
	}
}

// --- Stick, Fearless Mentor ---------------------------------------------------

func TestMscM13StickExilesTheTopCardWhenYourSourceDamagesYouOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Stick, Fearless Mentor", "Legendary Creature — Human Monk Hero", mscM13StickOracle, false)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	top := stackLibrary(me, "Impulse One", "Sorcery", "{1}")

	// A source an opponent controls does nothing.
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(theirs, me.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(top) {
		t.Fatal("damage from an opponent's source exiled a card")
	}

	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(mine, me.ID, 2); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Fatal("damage to you from your own source did not exile the top card")
	}
	if !b12ExiledPlayableBy(t, g, top, me.ID) {
		t.Error("the exiled card is not playable by Stick's controller")
	}

	second := stackLibrary(me, "Impulse Two", "Sorcery", "{1}")
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(mine, me.ID, 2); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(second) {
		t.Error("the ability triggered a second time in one turn")
	}
}

func TestMscM13StickTheCardIsPlayableUntilTheEndOfYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Stick, Fearless Mentor", "Legendary Creature — Human Monk Hero", mscM13StickOracle, false)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	top := seedLibrary(me, "Impulse One")[0]
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(mine, me.ID, 2); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)

	powerUpPassTo(t, g, 1)
	if !b12ExiledPlayableBy(t, g, top, me.ID) {
		t.Error("the card is not playable during an opponent's turn")
	}
	powerUpPassTo(t, g, 0)
	if !b12ExiledPlayableBy(t, g, top, me.ID) {
		t.Error("the card is not playable during your next turn")
	}
	powerUpPassTo(t, g, 1)
	if b12ExiledPlayableBy(t, g, top, me.ID) {
		t.Error("the card is still playable after the end of your next turn")
	}
}

// --- Stark's Ingenuity --------------------------------------------------------

func mscM13CastStarks(t *testing.T, g *game.Game, bear uuid.UUID, pool int) {
	t.Helper()
	castCatalogSpell(t, g, "Stark's Ingenuity", "Enchantment — Aura", mscM13StarksOracle, cardRefs(bear))
	fillPool(g.Seats[0], pool)
	passPriorityAroundTable(t, g)
}

func TestMscM13StarksIngenuityPaysXToDrawXAndEachDrawGrowsTheCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	handBefore := me.Hand.Size()

	mscM13CastStarks(t, g, bear, 3)
	pa := answerNumber(t, g, me.ID, 2)
	if pa.Min != 0 || pa.Max != 3 {
		t.Errorf("X prompt = %+v, want 0..3", pa)
	}
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("hand grew by %d, want 2 (two cards drawn)", got)
	}
	if n := hobA8Counters(t, g, bear); n != 2 {
		t.Errorf("Bear has %d +1/+1 counters, want 2: one per card drawn", n)
	}
	if left := me.ManaPool.Total(); left != 1 {
		t.Errorf("%d mana left, want 1 (paid {X} with X = 2 of 3)", left)
	}
}

func TestMscM13StarksIngenuityPayingNothingDrawsNothingButLaterDrawsStillCount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	mscM13CastStarks(t, g, bear, 2)
	answerNumber(t, g, me.ID, 0)
	passPriorityAroundTable(t, g)
	if n := hobA8Counters(t, g, bear); n != 0 {
		t.Fatalf("Bear has %d counters after X = 0", n)
	}

	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := hobA8Counters(t, g, bear); n != 0 {
		t.Errorf("an opponent's draw put %d counters on Bear", n)
	}
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := hobA8Counters(t, g, bear); n != 1 {
		t.Errorf("Bear has %d counters after your draw, want 1", n)
	}
}

// --- Mister Immortal ----------------------------------------------------------

func TestMscM13MisterImmortalReturnsTappedFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := pushCatalogGraveyardCard(me, "Mister Immortal", "Legendary Creature — Mutant Hero", mscM13ImmortalOracle, 2, 1)
	floatForTest(g, me, "CCG")

	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)

	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("Mister Immortal did not return to the battlefield")
	}
	if !c.Tapped {
		t.Error("Mister Immortal returned untapped")
	}
}

func TestMscM13MisterImmortalReturnsTappedFromExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := uuid.New()
	g.Exile.PushTop(game.Card{
		InstanceID: id, Name: "Mister Immortal", TypeLine: "Legendary Creature — Mutant Hero", OracleID: mscM13ImmortalOracle,
		Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	floatForTest(g, me, "CCG")

	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate from exile: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Exile.Contains(id) {
		t.Fatal("Mister Immortal is still in exile")
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Mister Immortal" && c.Controller == me.ID {
			found = true
			if !c.Tapped {
				t.Error("Mister Immortal returned untapped")
			}
		}
	}
	if !found {
		t.Fatal("Mister Immortal did not return to the battlefield")
	}
}

func TestMscM13MisterImmortalCostsThreeAndWorksOnlyFromThoseZones(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	grave := pushCatalogGraveyardCard(me, "Mister Immortal", "Legendary Creature — Mutant Hero", mscM13ImmortalOracle, 2, 1)
	floatForTest(g, me, "GG")
	if err := g.ActivateCatalogAbility(me.ID, grave, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated for {G}{G}: the ability costs {2}{G}")
	}
	if _, ok := battlefieldCard(g, grave); ok {
		t.Error("a refused activation returned the card")
	}

	onField := pushCatalogPermanent(g, me.ID, "Mister Immortal", "Legendary Creature — Mutant Hero", mscM13ImmortalOracle, false)
	floatForTest(g, me, "CCG")
	if err := g.ActivateCatalogAbility(me.ID, onField, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the return ability worked from the battlefield")
	}
	hand := handCardFull(me, "Mister Immortal", "Legendary Creature — Mutant Hero", "{1}{G}", mscM13ImmortalOracle, []string{"G"})
	if err := g.ActivateCatalogAbility(me.ID, hand, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the return ability worked from the hand")
	}
}

// --- Whirlwind, Killer Cyclone ------------------------------------------------

func TestMscM13WhirlwindStopsTheDefendersCreatureFromBlockingWhenANewCreatureAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	whirl := pushCatalogPermanent(g, me.ID, "Whirlwind, Killer Cyclone", "Legendary Creature — Mutant Villain", mscM13WhirlwindOracle, false)
	mscM13MarkEntered(g, whirl)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	theirsToo := b12Creature(g, opp.ID, "Their Other Bear", "Creature — Bear", 2, 2)
	elsewhere := b12Creature(g, third.ID, "Third Bear", "Creature — Bear", 2, 2)

	declareAttack(t, g, opp.ID, whirl)
	cards := mscM13OpenPickTargetCards(g, me.ID)
	if len(cards) == 0 {
		t.Fatal("no target was asked for")
	}
	for _, id := range cards {
		if id == elsewhere {
			t.Fatalf("a creature of a player who is not being attacked was offered: %v", cards)
		}
	}
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)

	if c := findBattlefieldCardByID(g, theirs); !game.Restricted(c, game.CantBlock) {
		t.Error("the chosen creature can still block")
	}
	if c := findBattlefieldCardByID(g, theirsToo); game.Restricted(c, game.CantBlock) {
		t.Error("a creature that was not chosen can't block")
	}
}

func TestMscM13WhirlwindDoesNothingWhenNoAttackerEnteredThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Pushed with no entry event: it was already on the battlefield.
	whirl := b12Push(g, me.ID, "Whirlwind, Killer Cyclone", "Legendary Creature — Mutant Villain", mscM13WhirlwindOracle, 2, 3)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	declareAttack(t, g, opp.ID, whirl)

	if n := mscM13TriggersFrom(g, whirl); n != 0 {
		t.Errorf("%d triggers for a creature that did not enter this turn", n)
	}
}

func TestMscM13WhirlwindTriggersOncePerPlayerAttackedNoMatterHowManyNewCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	whirl := pushCatalogPermanent(g, me.ID, "Whirlwind, Killer Cyclone", "Legendary Creature — Mutant Villain", mscM13WhirlwindOracle, false)
	oldBear := b12Creature(g, me.ID, "Old Bear", "Creature — Bear", 2, 2)
	newA := pushCatalogPermanent(g, me.ID, "New Bear A", "Creature — Bear", "", false)
	newB := pushCatalogPermanent(g, me.ID, "New Bear B", "Creature — Bear", "", false)
	mscM13MarkEntered(g, whirl, newA, newB)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b12Creature(g, third.ID, "Third Bear", "Creature — Bear", 2, 2)

	advanceTo(t, g, game.StepDeclareAttackers)
	// The old creature is declared first, then two new ones at one player and
	// one at another.
	for _, a := range []struct{ id, at uuid.UUID }{
		{oldBear, opp.ID}, {whirl, opp.ID}, {newA, opp.ID}, {newB, third.ID},
	} {
		if err := g.DeclareAttacker(a.id, a.at); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)

	n := 0
	for _, id := range []uuid.UUID{whirl} {
		n += mscM13TriggersFrom(g, id)
	}
	if n != 2 {
		t.Errorf("%d triggers, want 2: one for each player attacked by creatures that entered this turn", n)
	}
}

// --- Storm, Queen of Wakanda --------------------------------------------------

func TestMscM13StormGivesAnotherAttackerFlyingAndHerPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	storm := b12Push(g, me.ID, "Storm, Queen of Wakanda", "Legendary Creature — Mutant Noble Hero", mscM13StormOracle, 4, 5)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	idle := b12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)

	declareAttack(t, g, opp.ID, storm, bear)
	if cards := mscM13OpenPickTargetCards(g, me.ID); len(cards) > 0 {
		for _, id := range cards {
			if id == storm || id == idle {
				t.Fatalf("a legal target was wrong: %v", cards)
			}
		}
		answerPickTarget(t, g, bear)
	}
	passPriorityAroundTable(t, g)

	if got := effectivePower(t, g, bear); got != 6 {
		t.Errorf("Bear power %d, want 6 (2 + Storm's 4)", got)
	}
	if !containsString(effectiveAbilities(t, g, bear), "flying") {
		t.Error("Bear did not gain flying")
	}
	if got := effectivePower(t, g, idle); got != 2 {
		t.Errorf("a creature that is not attacking got %d power", got)
	}
}

func TestMscM13StormUsesHerLastKnownPowerWhenShesGone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	storm := b12Push(g, me.ID, "Storm, Queen of Wakanda", "Legendary Creature — Mutant Noble Hero", mscM13StormOracle, 4, 5)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	declareAttack(t, g, opp.ID, storm, bear)
	if openPickTarget(g) {
		answerPickTarget(t, g, bear)
	}
	killCreature(t, g, me.ID, storm)
	passPriorityAroundTable(t, g)

	if got := effectivePower(t, g, bear); got != 6 {
		t.Errorf("Bear power %d, want 6: X is Storm's last-known power", got)
	}
}

func TestMscM13StormHasNoTargetWhenSheAttacksAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	storm := b12Push(g, me.ID, "Storm, Queen of Wakanda", "Legendary Creature — Mutant Noble Hero", mscM13StormOracle, 4, 5)

	declareAttack(t, g, opp.ID, storm)

	if n := mscM13TriggersFrom(g, storm); n != 0 {
		t.Errorf("%d triggers: another target attacking creature is required", n)
	}
	if !g.Battlefield.Contains(storm) {
		t.Error("Storm left the battlefield")
	}
}

func mscM13FlyerAttack(t *testing.T, g *game.Game, flying bool, toughness int, at uuid.UUID) uuid.UUID {
	t.Helper()
	opp := g.Seats[1]
	var keywords []string
	if flying {
		keywords = []string{"flying"}
	}
	flyer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Raider", TypeLine: "Creature — Bird", Keywords: keywords,
		Power: 1, Toughness: toughness, Owner: opp.ID, Controller: opp.ID,
	})
	powerUpPassTo(t, g, 1)
	declareAttack(t, g, at, flyer)
	passPriorityAroundTable(t, g)
	return flyer
}

func TestMscM13StormDamagesAFlyingAttackerThatAttacksYou(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Storm, Queen of Wakanda", "Legendary Creature — Mutant Noble Hero", mscM13StormOracle, 4, 5)

	flyer := mscM13FlyerAttack(t, g, true, 4, me.ID)

	if g.Battlefield.Contains(flyer) {
		t.Error("a 1/4 flyer attacking Storm's controller survived Storm's 4 damage")
	}
}

func TestMscM13StormIgnoresGroundAttackersAndFlyersAttackingSomeoneElse(t *testing.T) {
	g := newCatalogGame(t)
	me, third := g.Seats[0], g.Seats[2]
	b12Push(g, me.ID, "Storm, Queen of Wakanda", "Legendary Creature — Mutant Noble Hero", mscM13StormOracle, 4, 5)

	ground := mscM13FlyerAttack(t, g, false, 4, me.ID)
	if !g.Battlefield.Contains(ground) {
		t.Error("Storm damaged a creature without flying")
	}

	g2 := newCatalogGame(t)
	b12Push(g2, g2.Seats[0].ID, "Storm, Queen of Wakanda", "Legendary Creature — Mutant Noble Hero", mscM13StormOracle, 4, 5)
	flyer := mscM13FlyerAttack(t, g2, true, 4, g2.Seats[2].ID)
	if !g2.Battlefield.Contains(flyer) {
		t.Error("Storm damaged a flyer that attacked another player")
	}
	_ = third
}

// --- Mister Fantastic ---------------------------------------------------------

func mscM13FantasticAtCombat(t *testing.T, g *game.Game, cast func()) {
	t.Helper()
	advanceToMain(t, g)
	cast()
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
}

func TestMscM13MisterFantasticDrawsAtCombatAfterANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Mister Fantastic", "Legendary Creature — Human Scientist Hero", mscM13FantasticOracle, false)
	handBefore := me.Hand.Size()

	mscM13FantasticAtCombat(t, g, func() {
		castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	})

	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("hand grew by %d, want 1 (the Bolt left the hand, one card drawn: net 1 over the pushed Bolt)", got)
	}
}

func TestMscM13MisterFantasticDrawsNothingWithoutANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Mister Fantastic", "Legendary Creature — Human Scientist Hero", mscM13FantasticOracle, false)
	advanceToMain(t, g)
	handBefore := me.Hand.Size()

	castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size() - handBefore; got != 0 {
		t.Errorf("hand changed by %d: a creature spell is not a noncreature spell", got)
	}
}

func TestMscM13MisterFantasticCopiesATriggeredAbilityTwice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fantastic := pushCatalogPermanent(g, me.ID, "Mister Fantastic", "Legendary Creature — Human Scientist Hero", mscM13FantasticOracle, false)
	other := pushCatalogPermanent(g, me.ID, "Other", "Artifact", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Theirs", "Artifact", "", false)
	if err := g.AnnounceTrigger(me.ID, other, game.AbilityParams{Label: "a trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	if err := g.AnnounceTrigger(opp.ID, theirs, game.AbilityParams{Label: "their trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	mine := acItemLabelled(g, "a trigger")
	notMine := acItemLabelled(g, "their trigger")

	offered := acLegalAbilityTargets(g, me.ID, mscM13FantasticOracle, 0)
	if !offered[mine] || offered[notMine] {
		t.Fatalf("offered %v: want only the trigger you control", offered)
	}

	floatForTest(g, me, "RGWU")
	if err := g.ActivateCatalogAbility(me.ID, fantastic, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	acPassUntilCopy(t, g)
	if n := acCopiesOnStack(g); n != 2 {
		t.Errorf("%d copies on the stack, want 2", n)
	}
	if c := b12Card(t, g, fantastic); !c.Tapped {
		t.Error("Mister Fantastic did not tap")
	}
}

func TestMscM13MisterFantasticNeedsAllFourColors(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fantastic := pushCatalogPermanent(g, me.ID, "Mister Fantastic", "Legendary Creature — Human Scientist Hero", mscM13FantasticOracle, false)
	other := pushCatalogPermanent(g, me.ID, "Other", "Artifact", "", false)
	if err := g.AnnounceTrigger(me.ID, other, game.AbilityParams{Label: "a trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	mine := acItemLabelled(g, "a trigger")

	floatForTest(g, me, "RGW")
	err := g.ActivateCatalogAbility(me.ID, fantastic, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}},
	})
	if err == nil {
		t.Error("activated without {U}")
	}
	if n := acCopiesOnStack(g); n != 0 {
		t.Errorf("%d copies after a refused activation", n)
	}
}

// --- The Thing ----------------------------------------------------------------

func TestMscM13TheThingGetsFourCountersAtCombatAfterANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thing := pushCatalogPermanent(g, me.ID, "The Thing", mscM13HeroType, mscM13ThingOracle, false)

	mscM13FantasticAtCombat(t, g, func() {
		castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	})

	if n := hobA8Counters(t, g, thing); n != 4 {
		t.Errorf("The Thing has %d counters, want 4", n)
	}
}

func TestMscM13TheThingGetsNothingWithoutANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	thing := pushCatalogPermanent(g, me.ID, "The Thing", mscM13HeroType, mscM13ThingOracle, false)

	mscM13FantasticAtCombat(t, g, func() {
		castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	})

	if n := hobA8Counters(t, g, thing); n != 0 {
		t.Errorf("The Thing has %d counters after only a creature spell", n)
	}
}

func mscM13ThingAttacks(t *testing.T, g *game.Game, mana string) (thing, bear uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	thing = pushCatalogPermanent(g, me.ID, "The Thing", mscM13HeroType, mscM13ThingOracle, false)
	bear = b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, thing)
	if mana != "" {
		floatForTest(g, me, mana)
	}
	passPriorityAroundTable(t, g)
	return thing, bear
}

func TestMscM13TheThingPaysFourColorsToDoubleEachKindOfCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	thing := pushCatalogPermanent(g, me.ID, "The Thing", mscM13HeroType, mscM13ThingOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for id, counters := range map[uuid.UUID]map[string]int{
		thing:  {game.CounterPlusOne: 3},
		bear:   {game.CounterPlusOne: 2, "charge": 1},
		theirs: {game.CounterPlusOne: 2},
	} {
		for kind, n := range counters {
			if err := g.AddCounterBy(me.ID, id, kind, n); err != nil {
				t.Fatalf("AddCounterBy: %v", err)
			}
		}
	}
	declareAttack(t, g, opp.ID, thing)
	floatForTest(g, me, "RGWU")
	passPriorityAroundTable(t, g)

	answerPayUnless(t, g, me.ID, true)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no targets were asked for")
	}
	for _, id := range prompt.PickTargetCards {
		if id == theirs {
			t.Fatalf("an opponent's permanent was offered: %v", prompt.PickTargetCards)
		}
	}
	if err := g.ResolvePickTargets(prompt.ID, me.ID, cardRefs(thing, bear)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)

	if n := hobA8Counters(t, g, thing); n != 6 {
		t.Errorf("The Thing has %d +1/+1 counters, want 6", n)
	}
	if n := hobA8Counters(t, g, bear); n != 4 {
		t.Errorf("Bear has %d +1/+1 counters, want 4", n)
	}
	if n := b12Counter(t, g, bear, "charge"); n != 2 {
		t.Errorf("Bear has %d charge counters, want 2: each kind doubles", n)
	}
	if n := hobA8Counters(t, g, theirs); n != 2 {
		t.Errorf("their Bear has %d counters, want 2: it was not chosen", n)
	}
	if left := me.ManaPool.Total(); left != 0 {
		t.Errorf("%d mana left, want 0", left)
	}
}

func TestMscM13TheThingDecliningThePaymentDoublesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	thing, bear := mscM13ThingAttacks(t, g, "RGWU")
	_ = thing
	if err := g.AddCounterBy(me.ID, bear, game.CounterPlusOne, 2); err != nil {
		t.Fatalf("AddCounterBy: %v", err)
	}
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	if n := hobA8Counters(t, g, bear); n != 2 {
		t.Errorf("Bear has %d counters after declining, want 2", n)
	}
	if left := me.ManaPool.Total(); left != 4 {
		t.Errorf("%d mana left, want 4: nothing was paid", left)
	}
}

// --- Nico Minoru, Runaway -----------------------------------------------------

func TestMscM13NicoDamagesEachOpponentForASpellCastFromExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	nico := pushCatalogPermanent(g, me.ID, "Nico Minoru, Runaway", "Legendary Creature — Human Warlock Hero", mscM13NicoOracle, false)
	hit := uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: hit, Name: "Two Drop", TypeLine: "Creature — Test", ManaCost: "{1}{G}",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Exiled Land", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	})
	discard := handCardOf(g, me, "Spare", "Sorcery", "{2}", "")
	floatForTest(g, me, "CCR")

	if err := g.ActivateCatalogAbility(me.ID, nico, 0, game.ActivateAbilityParams{
		Strict: true, DiscardIDs: []uuid.UUID{discard},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, hit)
	if err := castThroughResolutionCast(t, g, me.ID, hit, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the exiled card for free: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(hit) {
		t.Error("the exiled card was not cast")
	}
	for _, p := range g.Seats {
		want := 40
		if p.ID != me.ID {
			want = 38
		}
		if p.Life != want {
			t.Errorf("%s has %d life, want %d", p.Name, p.Life, want)
		}
	}
	_ = opp
}

func TestMscM13NicoDoesNothingForASpellCastFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Nico Minoru, Runaway", "Legendary Creature — Human Warlock Hero", mscM13NicoOracle, false)

	castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)

	if opp.Life != 40 {
		t.Errorf("opponent at %d life: a cast from hand does not trigger Nico", opp.Life)
	}
}

func TestMscM13NicoActivationNeedsADiscardAndAnUntappedNico(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	sick := pushCatalogPermanent(g, me.ID, "Nico Minoru, Runaway", "Legendary Creature — Human Warlock Hero", mscM13NicoOracle, true)
	discard := handCardOf(g, me, "Spare", "Sorcery", "{2}", "")
	floatForTest(g, me, "CCR")
	if err := g.ActivateCatalogAbility(me.ID, sick, 0, game.ActivateAbilityParams{
		Strict: true, DiscardIDs: []uuid.UUID{discard},
	}); err == nil {
		t.Error("activated while summoning sick")
	}

	ready := pushCatalogPermanent(g, me.ID, "Nico Minoru, Runaway", "Legendary Creature — Human Warlock Hero", mscM13NicoOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, ready, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated without discarding a card")
	}
	if !me.Hand.Contains(discard) {
		t.Error("a refused activation discarded the card")
	}
}

func TestMscM13NicoDecliningTheFreeCastLeavesTheCardInExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	nico := pushCatalogPermanent(g, me.ID, "Nico Minoru, Runaway", "Legendary Creature — Human Warlock Hero", mscM13NicoOracle, false)
	hit := uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: hit, Name: "Two Drop", TypeLine: "Creature — Test", ManaCost: "{1}{G}",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	discard := handCardOf(g, me, "Spare", "Sorcery", "{2}", "")
	floatForTest(g, me, "CCR")

	if err := g.ActivateCatalogAbility(me.ID, nico, 0, game.ActivateAbilityParams{
		Strict: true, DiscardIDs: []uuid.UUID{discard},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, hit)
	declineResolutionCast(t, g, me.ID)
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(hit) {
		t.Error("the declined card left exile")
	}
	if g.Battlefield.Contains(hit) {
		t.Error("the declined card was cast")
	}
}

// --- Willie Lumpkin, Postman --------------------------------------------------

func TestMscM13WillieCannotBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	willie := b12Push(g, me.ID, "Willie Lumpkin, Postman", "Legendary Creature — Human Citizen", mscM13WillieOracle, 1, 3)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(willie, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, willie); err == nil {
		t.Error("a creature blocked Willie Lumpkin")
	}
}

func mscM13WillieHits(t *testing.T, g *game.Game) (willie uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	willie = b12Push(g, me.ID, "Willie Lumpkin, Postman", "Legendary Creature — Human Citizen", mscM13WillieOracle, 1, 3)
	attackWith(t, g, opp.ID, willie)
	passPriorityAroundTable(t, g)
	return willie
}

func TestMscM13WillieYouDrawAndTheOpponentMayDrawAndIsThenStopped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myHand, theirHand := me.Hand.Size(), opp.Hand.Size()

	mscM13WillieHits(t, g)
	if got := me.Hand.Size() - myHand; got != 1 {
		t.Errorf("you drew %d cards, want 1", got)
	}
	answerMayChoice(t, g, opp.ID, true)
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - theirHand; got != 1 {
		t.Errorf("the opponent drew %d cards, want 1", got)
	}

	// Their next turn: they can't attack the controller.
	attacker := b12Creature(g, opp.ID, "Raider", "Creature — Bear", 2, 2)
	powerUpPassTo(t, g, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, me.ID); err == nil {
		t.Error("the opponent attacked the controller during their next turn")
	}
	if err := g.DeclareAttacker(attacker, g.Seats[2].ID); err != nil {
		t.Errorf("the opponent could not attack another player: %v", err)
	}
}

func TestMscM13WillieDecliningTheDrawLeavesTheOpponentFree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirHand := opp.Hand.Size()

	mscM13WillieHits(t, g)
	answerMayChoice(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - theirHand; got != 0 {
		t.Errorf("the opponent drew %d cards after declining", got)
	}

	attacker := b12Creature(g, opp.ID, "Raider", "Creature — Bear", 2, 2)
	powerUpPassTo(t, g, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, me.ID); err != nil {
		t.Errorf("the opponent who declined could not attack the controller: %v", err)
	}
}

// --- Quicksilver, Speedster ---------------------------------------------------

func mscM13SetTapped(t *testing.T, g *game.Game, id uuid.UUID, tapped bool) {
	t.Helper()
	g.WithWriteLock(func() { battlefieldCardPtr(t, g, id).Tapped = tapped })
}

func TestMscM13QuicksilverHasFlashDoubleStrikeAndHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	qs := pushCatalogPermanent(g, me.ID, "Quicksilver, Speedster", "Legendary Creature — Mutant Hero", mscM13QuicksilverOracle, false)
	abilities := effectiveAbilities(t, g, qs)
	for _, kw := range []string{"flash", "double strike", "haste"} {
		if !containsString(abilities, kw) {
			t.Errorf("Quicksilver has no %s: %v", kw, abilities)
		}
	}
}

func TestMscM13QuicksilverGivesFlashOnlyWhileTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	qs := pushCatalogPermanent(g, me.ID, "Quicksilver, Speedster", "Legendary Creature — Mutant Hero", mscM13QuicksilverOracle, false)
	advanceTo(t, g, game.StepBeginCombat)

	if marVellTimingOpen(g, me, "Sorcery") || marVellTimingOpen(g, me, "Creature — Bear") {
		t.Fatal("an untapped Quicksilver let a sorcery-speed spell begin at the beginning of combat")
	}
	mscM13SetTapped(t, g, qs, true)
	if !marVellTimingOpen(g, me, "Sorcery") || !marVellTimingOpen(g, me, "Creature — Bear") {
		t.Fatal("a tapped Quicksilver did not open the instant-speed window")
	}
	if marVellTimingOpen(g, opp, "Sorcery") {
		t.Error("Quicksilver gave its controller's opponent the window")
	}
	id := handCardFull(me, "Mystery Sorcery", "Sorcery", "", "", nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Errorf("a sorcery was refused at the beginning of combat while Quicksilver is tapped: %v", err)
	}
	mscM13SetTapped(t, g, qs, false)
	if marVellTimingOpen(g, me, "Sorcery") {
		t.Error("the window stayed open after Quicksilver untapped")
	}
}

func TestMscM13QuicksilverAnAttackerOpensTheWindowOnTheOpponentsTurnToo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	qs := pushCatalogPermanent(g, me.ID, "Quicksilver, Speedster", "Legendary Creature — Mutant Hero", mscM13QuicksilverOracle, false)

	declareAttack(t, g, opp.ID, qs)
	if !marVellTimingOpen(g, me, "Creature — Bear") {
		t.Error("an attacking (tapped) Quicksilver did not open the window")
	}
	// The next turn Quicksilver untaps for its controller only on their own
	// untap step: during an opponent's turn it is still tapped, so the
	// window is open there.
	powerUpPassTo(t, g, 1)
	if c := b12Card(t, g, qs); !c.Tapped {
		t.Fatal("Quicksilver untapped during an opponent's turn")
	}
	if !marVellTimingOpen(g, me, "Sorcery") {
		t.Error("the window was closed on an opponent's turn while Quicksilver stayed tapped")
	}
}
