package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestChainOfVaporCopiesItselfForTheSacrificingPlayer is the whole
// card, and the whole issue.
func TestChainOfVaporCopiesItselfForTheSacrificingPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushTribalCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.PushTribalCreature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := effectstest.PushLand(g, opp.ID, "Their Island")

	effectstest.CastCatalogSpell(t, g, "Chain of Vapor", "Instant", effectstest.ChainOfVaporOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)

	if !opp.Hand.Contains(theirs) {
		t.Fatalf("the targeted permanent went back to its owner's hand")
	}
	// The questions go to the BOUNCED permanent's controller.
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	sac := effectstest.LatestChooseCardsFor(g, opp.ID)
	if sac == nil {
		t.Fatalf("they pick which land to sacrifice: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(sac.ID, opp.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if g.Battlefield.Contains(land) {
		t.Error("the land was sacrificed")
	}

	// "If the player does, they may copy this spell."
	effectstest.AnswerMayChoice(t, g, opp.ID, true)

	// The copy's controller chooses its new target (CR 707.10c).
	prompt := effectstest.LatestPickTarget(g, opp.ID)
	if prompt == nil {
		t.Fatalf("the copy's controller re-targets it: %+v", g.PendingChoices)
	}
	if !effectstest.HasID(prompt.PickTargetCards, mine) {
		t.Fatalf("a new target is offered: %v", prompt.PickTargetCards)
	}
	if err := g.ResolvePickTarget(prompt.ID, opp.ID,
		game.TargetRef{Kind: game.TargetCard, ID: mine}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}

	// The copy is on the stack, under the sacrificing player's
	// control, before it resolves.
	copies := effectstest.CopiesOnStack(g)
	if len(copies) != 1 {
		t.Fatalf("one copy on the stack, got %d", len(copies))
	}
	if copies[0].Controller != opp.ID {
		t.Errorf("the copy is controlled by %s, want the player who made it (%s)",
			copies[0].Controller, opp.ID)
	}

	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(mine) {
		t.Error("the copy bounced its own new target")
	}
	// Chain of Vapor itself, and nothing else: a copy is not a card
	// (CR 707.10) and leaves no graveyard trace.
	if got := me.Graveyard.Size(); got != 1 {
		t.Errorf("the caster's graveyard holds %d cards, want 1 (the spell; a copy is not a card)", got)
	}
}

// TestChainOfVaporStopsWhenTheyDeclineTheSacrifice — the copy is
// gated on the sacrifice, and declining ends the chain.
func TestChainOfVaporStopsWhenTheyDeclineTheSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushTribalCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := effectstest.PushLand(g, opp.ID, "Their Island")

	effectstest.CastCatalogSpell(t, g, "Chain of Vapor", "Instant", effectstest.ChainOfVaporOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, opp.ID, false)

	if len(g.PendingChoices) != 0 {
		t.Errorf("declining ends the chain: %+v", g.PendingChoices)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("nothing was sacrificed")
	}
	if len(effectstest.CopiesOnStack(g)) != 0 {
		t.Error("no copy was made")
	}
	_ = me
}

// TestChainOfVaporAsksNobodyWhoControlsNoLand — "may sacrifice a
// land" with no land is not a question, and without the sacrifice
// there is no copy.
func TestChainOfVaporAsksNobodyWhoControlsNoLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.PushTribalCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Chain of Vapor", "Instant", effectstest.ChainOfVaporOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)

	if len(g.PendingChoices) != 0 {
		t.Errorf("a player with no land is asked nothing: %+v", g.PendingChoices)
	}
	if !opp.Hand.Contains(theirs) {
		t.Error("the bounce still happened")
	}
}

// TestChainOfVaporUndoesAcrossTheCopy — the resolving item rewinds
// with the prompt reading it, so answering the restored question
// still produces the copy.
func TestChainOfVaporUndoesAcrossTheCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushTribalCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// A second nonland permanent, so the copy has a new target to be
	// offered and the CR 707.10c prompt is actually asked.
	effectstest.PushTribalCreature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := effectstest.PushLand(g, opp.ID, "Their Island")

	effectstest.CastCatalogSpell(t, g, "Chain of Vapor", "Instant", effectstest.ChainOfVaporOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	sac := effectstest.LatestChooseCardsFor(g, opp.ID)
	if err := g.ResolveChooseCards(sac.ID, opp.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}

	snap := g.Clone()
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	if effectstest.LatestPickTarget(g, g.Seats[1].ID) == nil {
		t.Fatal("the copy's re-target prompt is open")
	}

	g.RestoreFrom(snap)
	opp = g.Seats[1]
	if effectstest.LatestPickTarget(g, opp.ID) != nil {
		t.Errorf("undo rewinds past the copy: %+v", g.PendingChoices)
	}
	if effectstest.LatestConfirmFor(g, opp.ID) == nil {
		t.Fatalf("undo puts the copy question back: %+v", g.PendingChoices)
	}
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	if effectstest.LatestPickTarget(g, opp.ID) == nil {
		t.Fatalf("answering the restored question still copies: %+v", g.PendingChoices)
	}
}

// TestChainOfSmogCopiesItselfForTheDiscardingPlayer — the other card
// in the cycle: no sacrifice gate, and the copy goes to the player who
// was hit.
func TestChainOfSmogCopiesItselfForTheDiscardingPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	before := len(opp.Hand.Cards)

	effectstest.CastCatalogSpell(t, g, "Chain of Smog", "Sorcery", effectstest.ChainOfSmogOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	// The discarding player chooses the two cards (not at random).
	hand := effectstest.DiscardChoiceFor(g, opp.ID)
	if hand == nil || hand.ChooseMin != 2 || hand.ChooseMax != 2 {
		t.Fatalf("the target chooses two cards to discard: %+v", hand)
	}
	picks := []uuid.UUID{opp.Hand.Cards[3].InstanceID, opp.Hand.Cards[5].InstanceID}
	effectstest.AnswerDiscard(t, g, opp.ID, picks...)
	for _, id := range picks {
		if !opp.Graveyard.Contains(id) {
			t.Error("the chosen card is in the graveyard")
		}
	}
	if got := len(opp.Hand.Cards); got != before-2 {
		t.Errorf("hand %d → %d, want two discarded", before, got)
	}
	effectstest.AnswerMayChoice(t, g, opp.ID, true)

	prompt := effectstest.LatestPickTarget(g, opp.ID)
	if prompt == nil {
		t.Fatalf("the copy's controller re-targets it: %+v", g.PendingChoices)
	}
	if !hasPlayerID(prompt.PickTargetPlayers, third.ID) {
		t.Fatalf("every player is a legal new target: %v", prompt.PickTargetPlayers)
	}
	thirdBefore := len(third.Hand.Cards)
	if err := g.ResolvePickTarget(prompt.ID, opp.ID,
		game.TargetRef{Kind: game.TargetPlayer, ID: third.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	copies := effectstest.CopiesOnStack(g)
	if len(copies) != 1 || copies[0].Controller != opp.ID {
		t.Fatalf("one copy controlled by the discarding player, got %+v", copies)
	}

	effectstest.PassPriorityAroundTable(t, g)
	// The copy's target chooses their own two cards too.
	effectstest.AnswerDiscard(t, g, third.ID, third.Hand.Cards[0].InstanceID, third.Hand.Cards[1].InstanceID)
	// The third player may also copy; decline so the chain stops.
	if c := effectstest.LatestConfirmFor(g, third.ID); c != nil {
		effectstest.AnswerMayChoice(t, g, third.ID, false)
	}
	if got := len(third.Hand.Cards); got != thirdBefore-2 {
		t.Errorf("the copy hit its new target: hand %d → %d, want two discarded", thirdBefore, got)
	}
	if got := me.Graveyard.Size(); got != 1 {
		t.Errorf("the caster's graveyard holds %d cards, want 1 (the spell only)", got)
	}
}

func hasPlayerID(ids []uuid.UUID, want uuid.UUID) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}
