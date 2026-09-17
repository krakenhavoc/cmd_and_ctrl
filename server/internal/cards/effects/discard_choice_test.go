package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// discard_choice_test.go — the shared vocabulary for "this card made
// somebody discard" (#651), plus the card-level tests for the two
// orders: a loot (draw, THEN discard) and a rummage (discard, THEN
// draw).
//
// Before #651 an effect discard was a counter in Game.DiscardPending —
// the cleanup step's hand-size map — so every test in this package
// asserted `g.DiscardPending[p.ID] == n` and then walked on. That
// assertion is what encoded the bug: nothing waited for the discard,
// and entering cleanup wiped the map. The discard is now a real
// PendingChoice, so the assertion is "a prompt is open for this seat,
// for this many cards", and the table is gated until it is answered.

const mindRotOracleID = "ad44cf74-b717-48fb-9fa2-77512024d76a"

// discardChoiceFor returns the discard prompt owed by a player right
// now, or nil. A discard is the choose-cards pick over the chooser's
// OWN hand; a test with two hand picks open at once tells them apart
// by Reason.
func discardChoiceFor(g *game.Game, player uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c == nil || c.Kind != game.PendingChoiceChooseCards {
			continue
		}
		if c.Chooser == player && c.FromPlayer == player {
			return c
		}
	}
	return nil
}

// discardOwed is how many cards a player has been asked to discard by
// an effect, 0 when no prompt is open for them. It replaces the old
// `g.DiscardPending[id]` read everywhere in this package.
func discardOwed(g *game.Game, player uuid.UUID) int {
	if c := discardChoiceFor(g, player); c != nil {
		return c.ChooseMax
	}
	return 0
}

// answerDiscard answers a player's discard prompt with named cards.
func answerDiscard(t *testing.T, g *game.Game, player uuid.UUID, ids ...uuid.UUID) {
	t.Helper()
	c := discardChoiceFor(g, player)
	if c == nil {
		t.Fatalf("no discard prompt for %s", player)
	}
	if err := g.ResolveChooseCards(c.ID, player, ids); err != nil {
		t.Fatalf("ResolveChooseCards (discard): %v", err)
	}
}

// discardFromHand answers a player's discard prompt off the top of
// their hand — what a test wants whenever WHICH card is pitched is not
// the thing under test.
func discardFromHand(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	c := discardChoiceFor(g, player)
	if c == nil {
		t.Fatalf("no discard prompt for %s", player)
	}
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Hand.Size() < c.ChooseMax {
		t.Fatalf("hand too small to answer a %d-card discard", c.ChooseMax)
	}
	ids := make([]uuid.UUID, 0, c.ChooseMax)
	for i := 0; i < c.ChooseMax; i++ {
		ids = append(ids, p.Hand.Cards[i].InstanceID)
	}
	if err := g.ResolveChooseCards(c.ID, player, ids); err != nil {
		t.Fatalf("ResolveChooseCards (discard): %v", err)
	}
}

// --- the two orders, through the real harness --------------------

// TestLootDrawsBeforeItAsksAndAsksBeforeItFinishes pins the looting
// order: Careful Study draws two, and the prompt it opens offers the
// POST-draw hand, so a card just drawn is a legal pitch. Nothing
// leaves the hand until the caster answers.
func TestLootDrawsBeforeItAsksAndAsksBeforeItFinishes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Careful Study", "Sorcery", b25CarefulStudyOracle, nil)
	passPriorityAroundTable(t, g)

	if me.Hand.Size() != before+2 {
		t.Fatalf("Careful Study draws two first: hand %d → %d", before, me.Hand.Size())
	}
	c := discardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("Careful Study then asks the caster to discard two")
	}
	if c.ChooseMin != 2 || c.ChooseMax != 2 {
		t.Errorf("prompt bounds = [%d,%d], want [2,2]", c.ChooseMin, c.ChooseMax)
	}
	if len(c.ChooseCards) != me.Hand.Size() {
		t.Errorf("the prompt offers the post-draw hand: %d candidates, hand %d",
			len(c.ChooseCards), me.Hand.Size())
	}
	// A card just drawn is among the candidates — the whole point of
	// drawing before asking.
	drawn := me.Hand.Cards[me.Hand.Size()-1].InstanceID
	found := false
	for _, id := range c.ChooseCards {
		if id == drawn {
			found = true
		}
	}
	if !found {
		t.Errorf("a card just drawn must be a legal discard")
	}
	// #791's gate: the loot is not over, so nobody moves.
	if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
		t.Errorf("PassPriority while the loot's discard is owed: %v, want ErrChoicePending", err)
	}
	discardFromHand(t, g, me.ID)
	if me.Hand.Size() != before {
		t.Errorf("hand %d → %d, want back to %d", before+2, me.Hand.Size(), before)
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("priority passes once the loot is paid: %v", err)
	}
}

// TestRummageDiscardsBeforeItDraws is the other order: Syphon Mind's
// "you draw a card for each card discarded this way" rides the
// prompt's continuation, so the caster draws nothing until an opponent
// has actually pitched, and draws exactly one per opponent who did.
func TestRummageDiscardsBeforeItDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	emptyHandToLibrary(g, c)
	before := me.Hand.Size()

	castCatalogSpell(t, g, "Syphon Mind", "Sorcery", b09SyphonMindOracle, nil)
	passPriorityAroundTable(t, g)

	if me.Hand.Size() != before {
		t.Errorf("the draw waits for the discard: hand %d → %d, want no change yet",
			before, me.Hand.Size())
	}
	if discardOwed(g, c.ID) != 0 {
		t.Error("an opponent with no hand was asked to discard")
	}
	if discardOwed(g, me.ID) != 0 {
		t.Error("the caster was asked to discard")
	}
	for i, opp := range []*game.Player{a, b} {
		if discardOwed(g, opp.ID) != 1 {
			t.Fatalf("%s owes one discard, got %d", opp.Name, discardOwed(g, opp.ID))
		}
		hand := opp.Hand.Size()
		discardFromHand(t, g, opp.ID)
		if opp.Hand.Size() != hand-1 {
			t.Errorf("%s pitched %d cards, want 1", opp.Name, hand-opp.Hand.Size())
		}
		if me.Hand.Size() != before+i+1 {
			t.Errorf("after %d discards the caster holds %d, want %d",
				i+1, me.Hand.Size(), before+i+1)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("both discards are paid, so priority passes: %v", err)
	}
}

// TestEffectDiscardWithAnEmptyHandQueuesNothing — CR 701.8a discards as
// many as you can, and a prompt with no candidates and a floor of one
// is a prompt nobody can answer, holding the whole table (#544).
func TestEffectDiscardWithAnEmptyHandQueuesNothing(t *testing.T) {
	g := newCatalogGame(t)
	target := g.Seats[1]
	emptyHandToLibrary(g, target)

	castCatalogSpell(t, g, "Mind Rot", "Sorcery", mindRotOracleID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}})
	passPriorityAroundTable(t, g)

	if c := discardChoiceFor(g, target.ID); c != nil {
		t.Fatalf("an empty hand queues no prompt, got one for %d cards", c.ChooseMax)
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("nothing is owed, so priority passes: %v", err)
	}
}
