package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const jackdawOracle = "ac161100-46b7-4f0f-a72b-84aaa45980ad"

// TestJackdawDiscardsHandAndDrawsForArtifacts is the whole "you may
// discard your hand. If you do, draw a card for each artifact you
// control" shape: a yes answer discards the whole hand and then draws
// once per artifact controlled AFTER the discard (Jackdaw itself plus
// a second artifact, so the count is observably more than one and not
// merely "the hand size").
func TestJackdawDiscardsHandAndDrawsForArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	jackdaw := castCatalogSpell(t, g, "Jackdaw", "Legendary Artifact — Vehicle", jackdawOracle, nil)
	passPriorityAroundTable(t, g)
	if isCreatureNow(g, jackdaw) {
		t.Fatal("a Vehicle is not a creature until it is crewed")
	}
	crewer := pushCrewerForTest(g, me.ID, "Crewer", 3)
	crewForTest(t, g, me.ID, jackdaw, crewer)

	// A second artifact besides Jackdaw itself.
	pushTypedCard(g, me.ID, "Sol Ring", "Artifact", "")

	fillHandTo(t, g, me, 4)

	dealCombatDamageToPlayer(g, jackdaw, opp.ID, 3)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != 2 {
		t.Fatalf("hand after discard+draw = %d, want 2 (one per artifact: Jackdaw + Sol Ring)", got)
	}
}

// TestJackdawIfYouDoIsSameResolutionNotAReflexiveTrigger pins the
// difference that made the first draft of this card wrong: "if you
// do" (CR 608.2c, same-resolution sequencing) is NOT "when you do"
// (CR 603.12, a reflexive triggered ability with its own stack object
// and response window). A reflexive trigger shares the parent's
// SOURCE card, so it can't be told apart from the parent by
// SourceCardID — what's unique is the STACK ITEM's own id in
// StackMeta. From the moment the parent's "you may" is answered yes
// to the moment the stack empties, the only key StackMeta ever holds
// must be the parent's own id: a second (or replacement) id appearing
// is exactly what a reflexive trigger going on the stack above the
// resolved parent would look like, and it's what would open a window
// for an opponent to act between the discard and the draw and change
// the artifact count — a window this card does not print.
func TestJackdawIfYouDoIsSameResolutionNotAReflexiveTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	jackdaw := castCatalogSpell(t, g, "Jackdaw", "Legendary Artifact — Vehicle", jackdawOracle, nil)
	passPriorityAroundTable(t, g)
	crewer := pushCrewerForTest(g, me.ID, "Crewer", 3)
	crewForTest(t, g, me.ID, jackdaw, crewer)
	pushTypedCard(g, me.ID, "Sol Ring", "Artifact", "")
	fillHandTo(t, g, me, 4)

	dealCombatDamageToPlayer(g, jackdaw, opp.ID, 3)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)

	var parentID uuid.UUID
	for id := range g.StackMeta {
		parentID = id
	}
	if parentID == uuid.Nil {
		t.Fatal("no stack item after answering yes")
	}

	for i := 0; i < 32 && !stackFullyEmpty(g); i++ {
		if n := len(g.StackMeta); n > 1 {
			t.Fatalf("StackMeta held %d items mid-resolution (iter %d); want at most 1", n, i)
		}
		for id := range g.StackMeta {
			if id != parentID {
				t.Fatalf("a different stack item (%s) appeared in place of the parent (%s) — \"if you do\" must not create a new triggered ability", id, parentID)
			}
		}
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				t.Fatalf("an unexpected prompt appeared mid-resolution: %v", err)
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !stackFullyEmpty(g) {
		t.Fatal("stack never emptied")
	}
	if got := me.Hand.Size(); got != 2 {
		t.Fatalf("hand after discard+draw = %d, want 2 (one per artifact)", got)
	}
}

// TestJackdawDeclinedDiscardChangesNothing — CR 603.4/603.5: a "no" to
// the parent's optional prompt means the discard never happens, so
// there is nothing for the "if you do" to have happened.
func TestJackdawDeclinedDiscardChangesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	jackdaw := castCatalogSpell(t, g, "Jackdaw", "Legendary Artifact — Vehicle", jackdawOracle, nil)
	passPriorityAroundTable(t, g)
	crewer := pushCrewerForTest(g, me.ID, "Crewer", 3)
	crewForTest(t, g, me.ID, jackdaw, crewer)

	fillHandTo(t, g, me, 4)
	handBefore := me.Hand.Size()

	dealCombatDamageToPlayer(g, jackdaw, opp.ID, 3)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != handBefore {
		t.Errorf("a declined discard changed hand size: %d -> %d", handBefore, got)
	}
}
