package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cool_but_rude_test.go — the three levels of a Class, each asserted
// through the state it produces rather than through the registry:
// level 1 rummages once per ATTACK, level 2's ping does not exist
// until the Class is level 2, and level 3 fires exactly once as the
// level is set.

const coolButRudeOracle = "1f076303-d160-4e02-aa1e-9ed6040c3735"

// pushCoolButRudeAttacker seeds a ready 2/2 for the controller.
func pushCoolButRudeAttacker(g *game.Game, controller uuid.UUID, name string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: controller, Controller: controller,
	})
}

// TestCoolButRudeRummagesOncePerAttackNotPerAttacker is the level-1
// line and the OncePerBatch half of it: three attackers, one trigger,
// one discard prompt, one card drawn.
func TestCoolButRudeRummagesOncePerAttackNotPerAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	class := effectstest.PushClass(g, me.ID, "Cool but Rude", "Enchantment — Class", coolButRudeOracle)
	one := pushCoolButRudeAttacker(g, me.ID, "Bear One")
	two := pushCoolButRudeAttacker(g, me.ID, "Bear Two")
	three := pushCoolButRudeAttacker(g, me.ID, "Bear Three")

	effectstest.DeclareAttack(t, g, opp.ID, one, two, three)
	if got := effectstest.TriggersOnStackFrom(g, class); got != 1 {
		t.Fatalf("triggers from a three-creature attack = %d, want 1 — \"whenever you attack\" is one per declaration", got)
	}

	hand, graves := me.Hand.Size(), me.Graveyard.Size()
	pitch := me.Hand.Cards[0].InstanceID
	effectstest.PassPriorityAroundTable(t, g)

	// "You may discard a card" — the prompt is over the pre-draw hand.
	effectstest.AnswerDiscard(t, g, me.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(pitch) {
		t.Error("the pitched card is not in the graveyard")
	}
	if got := me.Graveyard.Size(); got != graves+1 {
		t.Errorf("graveyard %d → %d, want exactly one card discarded", graves, got)
	}
	// One out, one in: the hand is the same size it started at.
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d → %d, want %d — discard one, draw one", hand, got, hand)
	}
}

// TestCoolButRudeDeclinedRummageDrawsNothing is the "if you do": the
// draw is conditional on the discard, so declining costs and gains
// nothing.
func TestCoolButRudeDeclinedRummageDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushClass(g, me.ID, "Cool but Rude", "Enchantment — Class", coolButRudeOracle)
	bear := pushCoolButRudeAttacker(g, me.ID, "Bear")

	effectstest.DeclareAttack(t, g, opp.ID, bear)
	hand, graves := me.Hand.Size(), me.Graveyard.Size()
	effectstest.PassPriorityAroundTable(t, g)

	// Answering the "up to one" prompt with nothing is the decline.
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d → %d, want unchanged — no discard, no draw", hand, got)
	}
	if got := me.Graveyard.Size(); got != graves {
		t.Errorf("graveyard %d → %d, want unchanged", graves, got)
	}
}

// TestCoolButRudeLevelTwoPingsOnADiscardAndNotBefore is the gate. The
// level-1 line is already discarding at level 1; nobody is pinged for
// it until the Class is level 2.
func TestCoolButRudeLevelTwoPingsOnADiscardAndNotBefore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	class := effectstest.PushClass(g, me.ID, "Cool but Rude", "Enchantment — Class", coolButRudeOracle)
	effectstest.AdvanceToMain(t, g)

	// At level 1 the ping is not an ability the Class has.
	lives := effectstest.OpponentLives(g, me.ID)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	for id, before := range lives {
		if now := g.PlayerByIDForEffect(id).Life; now != before {
			t.Fatalf("a level-1 Class pinged for %d — the level-2 line must not exist yet", before-now)
		}
	}

	if err := effectstest.LevelUpTo(t, g, me.ID, class, 0, "{1}{R}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	if got := effectstest.LevelOf(g, class); got != 2 {
		t.Fatalf("level = %d, want 2", got)
	}

	lives = effectstest.OpponentLives(g, me.ID)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	for id, before := range lives {
		if now := g.PlayerByIDForEffect(id).Life; now != before-2 {
			t.Errorf("opponent life %d → %d, want %d (2 damage)", before, now, before-2)
		}
	}
	if opp.Life == 40 {
		t.Error("the named opponent took no damage")
	}
}

// TestCoolButRudeLevelThreeTutorsThenDiscardsAtRandom is the "becomes
// level 3" trigger: a search for anything, into hand, and a random
// discard chained behind the search's prompt.
func TestCoolButRudeLevelThreeTutorsThenDiscardsAtRandom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	class := effectstest.PushClass(g, me.ID, "Cool but Rude", "Enchantment — Class", coolButRudeOracle)
	effectstest.AdvanceToMain(t, g)
	needle := effectstest.PushLibraryCardForTest(me, game.Card{
		Name: "Needle in the Deck", TypeLine: "Sorcery",
	})

	if err := effectstest.LevelUpTo(t, g, me.ID, class, 0, "{1}{R}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	// Level 2's line does nothing on its own; no prompt should be open.
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil {
		t.Fatal("levelling to 2 opened a search prompt")
	}

	graves := me.Graveyard.Size()
	if err := effectstest.LevelUpTo(t, g, me.ID, class, 1, "{1}{R}"); err != nil {
		t.Fatalf("level 3: %v", err)
	}
	if got := effectstest.LevelOf(g, class); got != 3 {
		t.Fatalf("level = %d, want 3", got)
	}

	effectstest.AnswerSearchByID(t, g, me.ID, needle)
	effectstest.PassPriorityAroundTable(t, g)

	if !me.Hand.Contains(needle) && !me.Graveyard.Contains(needle) {
		t.Error("the tutored card reached neither hand nor graveyard")
	}
	// "then discard a card at random" — one card left the hand, and
	// the tutored card was in it when the random pick was made.
	if got := me.Graveyard.Size(); got != graves+1 {
		t.Errorf("graveyard %d → %d, want exactly one random discard", graves, got)
	}
	if me.Library.Contains(needle) {
		t.Error("the tutored card is still in the library")
	}
}
