package effects

import "testing"

const jackdawOracle = "ac161100-46b7-4f0f-a72b-84aaa45980ad"

// TestJackdawDiscardsHandAndDrawsForArtifacts is the whole "you may
// discard your hand. If you do, draw a card for each artifact you
// control" reflexive shape: a yes answer discards the whole hand and
// then draws once per artifact controlled AFTER the discard (Jackdaw
// itself plus a second artifact, so the count is observably more than
// one and not merely "the hand size").
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

// TestJackdawDeclinedDiscardChangesNothing — CR 603.4/603.12: a "no"
// to the parent's optional prompt means the discard never happens, so
// the reflexive "if you do" trigger is never created either.
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
