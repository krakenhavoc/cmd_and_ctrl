package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// celes_rune_knight_test.go — Celes, Rune Knight (#3770, deck #3655).
//
// The trigger asks one question of each creature that enters: did it
// come OUT of a graveyard (EventETB.EnteredFrom), or was the spell that
// became it CAST from one (Card.Provenance.FromZone, CR 400.7d)? Undying
// answers the first, an escaped Pharika's Spawn the second, and a hard
// cast from hand neither.

const celesOracle = "aea8345a-191d-4c62-8d15-a716a9934b66"

const celesTypeLine = "Legendary Creature — Human Wizard Knight"

func celesInPlay(g *game.Game, owner uuid.UUID) uuid.UUID {
	return poolPush(g, owner, "Celes, Rune Knight", celesOracle, 4, 4)
}

func TestCelesRegisteredFull(t *testing.T) {
	spec, ok := Lookup(celesOracle)
	if !ok || spec.Completeness != CompletenessFull {
		t.Fatalf("registered=%v completeness=%v", ok, spec.Completeness)
	}
}

// The enters trigger: discard any number, then draw that many plus one.
func TestCelesEntersDiscardsAnyNumberThenDrawsThatManyPlusOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	a := ssCard(me.Hand, me.ID, "Card A", "Sorcery", 0, 0)
	b := ssCard(me.Hand, me.ID, "Card B", "Instant", 0, 0)
	ssCard(me.Hand, me.ID, "Card C", "Instant", 0, 0)

	castCatalogSpell(t, g, "Celes, Rune Knight", celesTypeLine, celesOracle, nil)
	passPriorityAroundTable(t, g)

	c := discardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enters trigger asks which cards to discard")
	}
	if c.ChooseMin != 0 || c.ChooseMax != 3 {
		t.Errorf("any number: bounds [%d,%d], want [0,3]", c.ChooseMin, c.ChooseMax)
	}
	answerDiscard(t, g, me.ID, a, b)
	if got := me.Hand.Size(); got != 1+3 {
		t.Errorf("hand %d, want 4 (one kept, three drawn for two discarded)", got)
	}
}

func TestCelesDiscardingNothingStillDrawsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	ssCard(me.Hand, me.ID, "Card A", "Sorcery", 0, 0)

	castCatalogSpell(t, g, "Celes, Rune Knight", celesTypeLine, celesOracle, nil)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID)
	if got := me.Hand.Size(); got != 2 {
		t.Errorf("hand %d, want 2 (kept one, drew one)", got)
	}
}

// A creature that enters from a graveyard (undying) puts a counter on
// each creature Celes' controller has, Celes included.
func TestCelesCountersWhenACreatureEntersFromAGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	celes := celesInPlay(g, me.ID)
	bear := poolPush(g, me.ID, "Bear", "", 2, 2)
	ghoul := poolPush(g, me.ID, "Sightless Ghoul", poolSightlessGhoul, 2, 2)

	b25Destroy(g, ghoul)
	settleProwess(t, g)

	if !onBattlefield(g, ghoul) {
		t.Fatal("undying did not return the Ghoul")
	}
	// The Ghoul's own undying counter is a +1/+1; Celes' trigger adds one.
	if got := countersOn(g, ghoul, game.CounterPlusOne); got != 2 {
		t.Errorf("Ghoul +1/+1 counters %d, want 2 (undying and Celes)", got)
	}
	for name, id := range map[string]uuid.UUID{"Celes": celes, "Bear": bear} {
		if got := countersOn(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("%s has %d +1/+1 counters, want 1", name, got)
		}
	}
}

// A creature cast from a graveyard enters from the STACK, so only the
// provenance reading catches it.
func TestCelesCountersWhenACreatureWasCastFromAGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	celes := celesInPlay(g, me.ID)

	id := escapeIntoPlay(t, g, "Pharika's Spawn", "Creature — Gorgon", pharikasSpawnOracle, 3)
	settleProwess(t, g)

	if !g.Battlefield.Contains(id) {
		t.Fatal("Pharika's Spawn never landed")
	}
	if ev := efgLastETB(t, g, id); ev.EnteredFrom == game.ZoneGraveyard {
		t.Fatalf("a cast spell enters from the stack, got %q", ev.EnteredFrom)
	}
	if got := countersOn(g, celes, game.CounterPlusOne); got != 1 {
		t.Errorf("Celes has %d +1/+1 counters, want 1 (the Spawn was cast from a graveyard)", got)
	}
}

// The same creature hard-cast from hand is neither.
func TestCelesIgnoresACreatureCastFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	celes := celesInPlay(g, me.ID)

	id := handCastIntoPlay(t, g, "Pharika's Spawn", "Creature — Gorgon", pharikasSpawnOracle)
	settleProwess(t, g)

	if !g.Battlefield.Contains(id) {
		t.Fatal("Pharika's Spawn never landed")
	}
	if got := countersOn(g, celes, game.CounterPlusOne); got != 0 {
		t.Errorf("Celes has %d counters, want 0 for a hand cast", got)
	}
}

// "Other": Celes returning from the graveyard does not trigger itself.
func TestCelesDoesNotCountItselfEnteringFromAGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	celes := celesInPlay(g, me.ID)
	g.WithWriteLock(func() {
		if _, err := game.MoveCard(g.Battlefield, me.Graveyard, celes); err != nil {
			t.Fatalf("to the graveyard: %v", err)
		}
		if _, err := game.MoveCard(me.Graveyard, g.Battlefield, celes); err != nil {
			t.Fatalf("back: %v", err)
		}
	})
	settleProwess(t, g)
	if got := countersOn(g, celes, game.CounterPlusOne); got != 0 {
		t.Errorf("Celes has %d counters, want 0 — it is not 'another' creature", got)
	}
}

// "You control": an opponent's creature returning from their graveyard
// does nothing.
func TestCelesIgnoresAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	celes := celesInPlay(g, me.ID)
	ghoul := poolPush(g, opp.ID, "Sightless Ghoul", poolSightlessGhoul, 2, 2)

	b25Destroy(g, ghoul)
	settleProwess(t, g)

	if !onBattlefield(g, ghoul) {
		t.Fatal("undying did not return the opponent's Ghoul")
	}
	if got := countersOn(g, celes, game.CounterPlusOne); got != 0 {
		t.Errorf("Celes has %d counters, want 0", got)
	}
}

// A creature that did not come from a graveyard (a hard cast Bear)
// does nothing.
func TestCelesIgnoresAnOrdinaryCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	celes := celesInPlay(g, me.ID)

	castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)

	if got := countersOn(g, celes, game.CounterPlusOne); got != 0 {
		t.Errorf("Celes has %d counters, want 0", got)
	}
}

// "One or more": Rise of the Dark Realms puts two creatures onto the
// battlefield from graveyards in ONE event, so it is one trigger and
// every creature Celes' controller has gets ONE counter, not two.
func TestCelesTwoGraveyardEntriesInOneBatchAreOneTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	celes := celesInPlay(g, me.ID)
	a, b := gtWarden(me), gtWarden(opp)

	castCatalogSpell(t, g, "Rise of the Dark Realms", "Sorcery", b04RiseOfTheDarkRealmsOracle, nil)
	gtSettle(t, g)

	for name, id := range map[string]uuid.UUID{"Celes": celes, "first": a, "second": b} {
		if !onBattlefield(g, id) {
			t.Fatalf("%s is not on the battlefield", name)
		}
		if got := countersOn(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("%s has %d +1/+1 counters, want 1 (one batch, one trigger)", name, got)
		}
	}
}
