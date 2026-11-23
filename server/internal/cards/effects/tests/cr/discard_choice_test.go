package testscr

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- the two orders, through the real harness --------------------

// TestLootDrawsBeforeItAsksAndAsksBeforeItFinishes pins the looting
// order: Careful Study draws two, and the prompt it opens offers the
// POST-draw hand, so a card just drawn is a legal pitch. Nothing
// leaves the hand until the caster answers.
func TestLootDrawsBeforeItAsksAndAsksBeforeItFinishes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Careful Study", "Sorcery", effectstest.B25CarefulStudyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if me.Hand.Size() != before+2 {
		t.Fatalf("Careful Study draws two first: hand %d → %d", before, me.Hand.Size())
	}
	c := effectstest.DiscardChoiceFor(g, me.ID)
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
	effectstest.DiscardFromHand(t, g, me.ID)
	if me.Hand.Size() != before {
		t.Errorf("hand %d → %d, want back to %d", before+2, me.Hand.Size(), before)
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("priority passes once the loot is paid: %v", err)
	}
}

// TestRummageDiscardsBeforeItDraws is the other order: Syphon Mind's
// "you draw a card for each card discarded this way" is the prompted-
// discard RUN's continuation (#1027), so the caster draws nothing
// until the WHOLE table has pitched, and then draws exactly one per
// card that was really discarded.
//
// It used to ride each prompt's own Then and draw one card per answer,
// spread across the table's decisions. That was this card's declared
// cosmetic simplification and it is gone.
func TestRummageDiscardsBeforeItDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	effectstest.EmptyHandToLibrary(g, c)
	before := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Syphon Mind", "Sorcery", effectstest.B09SyphonMindOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if me.Hand.Size() != before {
		t.Errorf("the draw waits for the discard: hand %d → %d, want no change yet",
			before, me.Hand.Size())
	}
	if effectstest.DiscardOwed(g, c.ID) != 0 {
		t.Error("an opponent with no hand was asked to discard")
	}
	if effectstest.DiscardOwed(g, me.ID) != 0 {
		t.Error("the caster was asked to discard")
	}
	for i, opp := range []*game.Player{a, b} {
		if effectstest.DiscardOwed(g, opp.ID) != 1 {
			t.Fatalf("%s owes one discard, got %d", opp.Name, effectstest.DiscardOwed(g, opp.ID))
		}
		hand := opp.Hand.Size()
		effectstest.DiscardFromHand(t, g, opp.ID)
		if opp.Hand.Size() != hand-1 {
			t.Errorf("%s pitched %d cards, want 1", opp.Name, hand-opp.Hand.Size())
		}
		// The draw is the RUN's, so it lands once — after the LAST
		// opponent, not one card per answer.
		want := before
		if i == 1 {
			want = before + 2
		}
		if me.Hand.Size() != want {
			t.Errorf("after %d of two discards the caster holds %d, want %d",
				i+1, me.Hand.Size(), want)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("both discards are paid, so priority passes: %v", err)
	}
}

// TestEffectDiscardWithAnEmptyHandQueuesNothing — CR 701.9a discards as
// many as you can, and a prompt with no candidates and a floor of one
// is a prompt nobody can answer, holding the whole table (#544).
func TestEffectDiscardWithAnEmptyHandQueuesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	target := g.Seats[1]
	effectstest.EmptyHandToLibrary(g, target)

	effectstest.CastCatalogSpell(t, g, "Mind Rot", "Sorcery", effectstest.MindRotOracleID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	if c := effectstest.DiscardChoiceFor(g, target.ID); c != nil {
		t.Fatalf("an empty hand queues no prompt, got one for %d cards", c.ChooseMax)
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("nothing is owed, so priority passes: %v", err)
	}
}
