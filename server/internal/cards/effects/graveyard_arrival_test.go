package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tests for the "put into your graveyard from anywhere" batch reader
// (graveyard_arrival.go, #4721) and the three Marvel Commander cards built
// on it.

const (
	voraciousBroodOracle        = "aeee8308-76aa-4ed5-b7b2-04fb9e29c676"
	ultronsAuxiliaryOracle      = "ff1ddf25-df25-4733-b9f1-236ad5a08782"
	ultronTheAnnihilatorOracle  = "33f2c6cb-f864-4bee-b365-aedf33fbedf5"
	graveyardArrivalBroodType   = "Creature — Alien Insect"
	graveyardArrivalRobotType   = "Artifact Creature — Robot Villain"
	graveyardArrivalPlusOneKind = game.CounterPlusOne
)

// triggersQueued is how many triggered abilities wait to go on the stack or
// already sit on it.
func triggersQueued(g *game.Game) int { return len(g.PendingTriggers) + len(g.StackMeta) }

func pushHandCardOfType(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

func TestVoraciousBroodTwoCreatureCardsDiscardedTogetherAreOneTriggerForTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	brood := pushCatalogPermanent(g, me.ID, "Voracious Brood", graveyardArrivalBroodType, voraciousBroodOracle, false)
	emptyHand(g, me)
	pushHandCardOfType(me, "Bear A", "Creature — Bear")
	pushHandCardOfType(me, "Bear B", "Creature — Bear")

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 2) })
	if n := triggersQueued(g); n != 1 {
		t.Fatalf("two creature cards discarded at once: %d triggers, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, brood, graveyardArrivalPlusOneKind); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2 (that many)", got)
	}
}

func TestVoraciousBroodCountsADeathAndAMillInTheSameBatchTogether(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	brood := pushCatalogPermanent(g, me.ID, "Voracious Brood", graveyardArrivalBroodType, voraciousBroodOracle, false)
	bear := seedCreature(g, "Bear", me.ID)
	b02TopOfLibrary(me, "Milled Bear", "Creature — Bear")
	b02TopOfLibrary(me, "Milled Rock", "Artifact")

	// One resolution's worth of events: a creature dies, a creature card and
	// a noncreature card are milled. All one event batch (event_batch.go).
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(bear)
		_ = g.MillNForEffect(me.ID, 2)
	})
	if n := triggersQueued(g); n != 1 {
		t.Fatalf("one batch: %d triggers, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, brood, graveyardArrivalPlusOneKind); got != 2 {
		t.Errorf("a death and a mill in one batch: %d counters, want 2 (the Rock is no creature card)", got)
	}

	// The next batch is a new occurrence: one more creature card, one more
	// trigger, one more counter.
	b02TopOfLibrary(me, "Next Bear", "Creature — Bear")
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	if got := countersOn(g, brood, graveyardArrivalPlusOneKind); got != 3 {
		t.Errorf("a later batch: %d counters, want 3", got)
	}
}

func TestVoraciousBroodIgnoresAnOpponentsGraveyardAndTokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brood := pushCatalogPermanent(g, me.ID, "Voracious Brood", graveyardArrivalBroodType, voraciousBroodOracle, false)

	// An opponent's mill and an opponent's creature dying.
	b02TopOfLibrary(opp, "Their Milled Bear", "Creature — Bear")
	theirs := seedCreature(g, "Their Bear", opp.ID)
	g.WithWriteLock(func() {
		_ = g.MillNForEffect(opp.ID, 1)
		_ = g.DestroyPermanentForEffect(theirs)
	})
	if n := triggersQueued(g); n != 0 {
		t.Fatalf("an opponent's graveyard fired %d triggers, want 0", n)
	}

	// My creature token dying is not a creature CARD (CR 108.2b).
	g.BeginEventBatchForTest()
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1}, 1)
	})
	tok := findBattlefieldByName(g, "Goblin")
	if tok == uuid.Nil {
		t.Fatal("token not created")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tok) })
	if n := triggersQueued(g); n != 0 {
		t.Fatalf("a dying creature token fired %d triggers, want 0", n)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, brood, graveyardArrivalPlusOneKind); got != 0 {
		t.Errorf("counters = %d, want 0", got)
	}
}

func TestVoraciousBroodCountsACardThatLeftTheGraveyardInResponse(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	brood := pushCatalogPermanent(g, me.ID, "Voracious Brood", graveyardArrivalBroodType, voraciousBroodOracle, false)
	b02TopOfLibrary(me, "Milled Bear", "Creature — Bear")
	b02TopOfLibrary(me, "Milled Elf", "Creature — Elf")
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 2) })
	// One of them is exiled before the trigger resolves: it was still put
	// into the graveyard, so "that many" counts it.
	elf := uuid.Nil
	for _, c := range me.Graveyard.Cards {
		if c.Name == "Milled Elf" {
			elf = c.InstanceID
		}
	}
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(elf) })
	passPriorityAroundTable(t, g)
	if got := countersOn(g, brood, graveyardArrivalPlusOneKind); got != 2 {
		t.Errorf("counters = %d, want 2", got)
	}
}

func TestVoraciousBroodEntersWithACounterPerCreatureCardInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushGraveyardCardForTest(me, "Old Bear")
	pushGraveyardCardForTest(me, "Old Elf")
	brood := castCatalogSpell(t, g, "Voracious Brood", graveyardArrivalBroodType, voraciousBroodOracle, nil)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, brood, graveyardArrivalPlusOneKind); got != 2 {
		t.Errorf("entered with %d counters, want 2", got)
	}
}

func TestUltronsAuxiliaryCountsEachArtifactWhereverItCameFrom(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aux := pushCatalogPermanent(g, me.ID, "Ultron's Auxiliary", graveyardArrivalRobotType, ultronsAuxiliaryOracle, false)
	want := 0
	check := func(step string, add int) {
		t.Helper()
		passPriorityAroundTable(t, g)
		want += add
		if got := countersOn(g, aux, graveyardArrivalPlusOneKind); got != want {
			t.Errorf("%s: %d counters, want %d", step, got, want)
		}
	}

	// Another artifact dies.
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rock) })
	check("an artifact dies", 1)

	// An artifact creature of an opponent's dying is not put into MY graveyard.
	theirs := b12Permanent(g, opp.ID, "Their Golem", "Artifact Creature — Golem")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	check("an opponent's artifact dies", 0)

	// A nonartifact creature dying does nothing.
	bear := seedCreature(g, "Bear", me.ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	check("a creature dies", 0)

	// An artifact TOKEN dying counts: it was put into the graveyard.
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, TokenCard("2/2 colorless Robot Villain artifact"), 1)
	})
	tok := findBattlefieldByName(g, "Robot Villain")
	if tok == uuid.Nil {
		t.Fatal("token not created")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tok) })
	check("an artifact token dies", 1)

	// An artifact CARD milled, and one discarded.
	b02TopOfLibrary(me, "Milled Rock", "Artifact")
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 1) })
	check("an artifact card is milled", 1)

	emptyHand(g, me)
	pushHandCardOfType(me, "Discarded Rock", "Artifact — Equipment")
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	check("an artifact card is discarded", 1)

	// A nonartifact card milled, and an opponent's artifact card milled.
	b02TopOfLibrary(me, "Milled Bear", "Creature — Bear")
	b02TopOfLibrary(opp, "Their Milled Rock", "Artifact")
	g.WithWriteLock(func() {
		_ = g.MillNForEffect(me.ID, 1)
		_ = g.MillNForEffect(opp.ID, 1)
	})
	check("a creature card and an opponent's artifact card are milled", 0)
}

func TestUltronsAuxiliaryTwoArtifactsAtOnceAreTwoTriggers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	aux := pushCatalogPermanent(g, me.ID, "Ultron's Auxiliary", graveyardArrivalRobotType, ultronsAuxiliaryOracle, false)
	emptyHand(g, me)
	pushHandCardOfType(me, "Rock A", "Artifact")
	pushHandCardOfType(me, "Rock B", "Artifact")
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 2) })
	if n := triggersQueued(g); n != 2 {
		t.Fatalf("two artifact cards discarded together: %d triggers, want 2 (each is an occurrence, CR 603.2c)", n)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, aux, graveyardArrivalPlusOneKind); got != 2 {
		t.Errorf("counters = %d, want 2", got)
	}
}

func TestUltronTheAnnihilatorDrainsEachOpponentPerArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Ultron the Annihilator", "Legendary Artifact Creature — Robot Villain", ultronTheAnnihilatorOracle, false)
	lives := func() []int {
		out := make([]int, len(g.Seats))
		for i, s := range g.Seats {
			out[i] = s.Life
		}
		return out
	}
	before := lives()
	assertDrained := func(step string, n int) {
		t.Helper()
		for i, l := range lives() {
			want := before[i]
			if i != 0 {
				want -= n
			}
			if l != want {
				t.Errorf("%s: seat %d at %d, want %d", step, i, l, want)
			}
		}
	}

	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, TokenCard("2/2 colorless Robot Villain artifact"), 1)
	})
	tok := findBattlefieldByName(g, "Robot Villain")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tok) })
	passPriorityAroundTable(t, g)
	assertDrained("an artifact token dies", 1)

	emptyHand(g, me)
	pushHandCardOfType(me, "Discarded Rock", "Artifact")
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	assertDrained("an artifact card is discarded", 2)

	bear := seedCreature(g, "Bear", me.ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	assertDrained("a creature dies", 2)
}

// The reader itself: a dying artifact card is read by the battlefield half
// (WhenAnotherArtifactIsPutIntoYourGraveyard), never by the "anywhere other
// than the battlefield" half, and a token is never a card at all.
func TestUltronArtifactCardsFromElsewhereNeverCountADeathOrAToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, TokenCard("2/2 colorless Robot Villain artifact"), 1)
	})
	tok := findBattlefieldByName(g, "Robot Villain")
	b02TopOfLibrary(me, "Milled Rock", "Artifact")
	emptyHand(g, me)
	pushHandCardOfType(me, "Discarded Rock", "Artifact")
	var batch uint64
	g.WithWriteLock(func() {
		batch = g.CurrentEventBatchForEffect()
		_ = g.DestroyPermanentForEffect(rock)
		_ = g.DestroyPermanentForEffect(tok)
		_ = g.MillNForEffect(me.ID, 1)
		_ = g.DiscardRandomForEffect(me.ID, 1)
	})
	var elsewhere, anywhere []uuid.UUID
	g.WithWriteLock(func() {
		elsewhere = artifactCardsToYourGraveyardNotFromTheBattlefield.InBatch(g, batch, me.ID)
		anywhere = graveyardArrivals{Match: isArtifactCard}.InBatch(g, batch, me.ID)
	})
	if len(elsewhere) != 2 {
		t.Errorf("artifact cards from elsewhere = %d, want 2 (the mill and the discard)", len(elsewhere))
	}
	// "From anywhere" adds the Rock's death, but still not the token.
	if len(anywhere) != 3 {
		t.Errorf("artifact cards from anywhere = %d, want 3 (token excluded)", len(anywhere))
	}
}
