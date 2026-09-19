package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// chain_cycle_test.go — #920. A spell copying ITSELF at resolution
// (CR 707.10), which the engine could not do until the resolving
// item's metadata outlived its own resolution function.
//
// The assertions that matter are the three the bug hid: a copy is
// created at all, it is controlled by the player who chose to make it
// rather than by the caster (CR 707.10b), and its target is the one
// that player picked.

const (
	chainOfVaporOracle = "2167ac25-d042-4b48-b770-8b94acc1a965"
	chainOfSmogOracle  = "ea14c26b-bf2f-48b4-b879-6e63069ded1f"
)

// pushLand puts a land on the battlefield under a player's control.
func pushLand(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Basic Land — Island",
		Owner: owner, Controller: owner,
	})
	return id
}

// TestChainOfVaporCopiesItselfForTheSacrificingPlayer is the whole
// card, and the whole issue.
func TestChainOfVaporCopiesItselfForTheSacrificingPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushTribalCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := pushTribalCreature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := pushLand(g, opp.ID, "Their Island")

	castCatalogSpell(t, g, "Chain of Vapor", "Instant", chainOfVaporOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)

	if !opp.Hand.Contains(theirs) {
		t.Fatalf("the targeted permanent went back to its owner's hand")
	}
	// The questions go to the BOUNCED permanent's controller.
	answerMayChoice(t, g, opp.ID, true)
	sac := latestChooseCardsFor(g, opp.ID)
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
	answerMayChoice(t, g, opp.ID, true)

	// The copy's controller chooses its new target (CR 707.10c).
	prompt := latestPickTarget(g, opp.ID)
	if prompt == nil {
		t.Fatalf("the copy's controller re-targets it: %+v", g.PendingChoices)
	}
	if !hasID(prompt.PickTargetCards, mine) {
		t.Fatalf("a new target is offered: %v", prompt.PickTargetCards)
	}
	if err := g.ResolvePickTarget(prompt.ID, opp.ID,
		game.TargetRef{Kind: game.TargetCard, ID: mine}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}

	// The copy is on the stack, under the sacrificing player's
	// control, before it resolves.
	copies := copiesOnStack(g)
	if len(copies) != 1 {
		t.Fatalf("one copy on the stack, got %d", len(copies))
	}
	if copies[0].Controller != opp.ID {
		t.Errorf("the copy is controlled by %s, want the player who made it (%s)",
			copies[0].Controller, opp.ID)
	}

	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushTribalCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := pushLand(g, opp.ID, "Their Island")

	castCatalogSpell(t, g, "Chain of Vapor", "Instant", chainOfVaporOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, opp.ID, false)

	if len(g.PendingChoices) != 0 {
		t.Errorf("declining ends the chain: %+v", g.PendingChoices)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("nothing was sacrificed")
	}
	if len(copiesOnStack(g)) != 0 {
		t.Error("no copy was made")
	}
	_ = me
}

// TestChainOfVaporAsksNobodyWhoControlsNoLand — "may sacrifice a
// land" with no land is not a question, and without the sacrifice
// there is no copy.
func TestChainOfVaporAsksNobodyWhoControlsNoLand(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := pushTribalCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Chain of Vapor", "Instant", chainOfVaporOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushTribalCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// A second nonland permanent, so the copy has a new target to be
	// offered and the CR 707.10c prompt is actually asked.
	pushTribalCreature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := pushLand(g, opp.ID, "Their Island")

	castCatalogSpell(t, g, "Chain of Vapor", "Instant", chainOfVaporOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, opp.ID, true)
	sac := latestChooseCardsFor(g, opp.ID)
	if err := g.ResolveChooseCards(sac.ID, opp.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}

	snap := g.Clone()
	answerMayChoice(t, g, opp.ID, true)
	if latestPickTarget(g, g.Seats[1].ID) == nil {
		t.Fatal("the copy's re-target prompt is open")
	}

	g.RestoreFrom(snap)
	opp = g.Seats[1]
	if latestPickTarget(g, opp.ID) != nil {
		t.Errorf("undo rewinds past the copy: %+v", g.PendingChoices)
	}
	if latestConfirmFor(g, opp.ID) == nil {
		t.Fatalf("undo puts the copy question back: %+v", g.PendingChoices)
	}
	answerMayChoice(t, g, opp.ID, true)
	if latestPickTarget(g, opp.ID) == nil {
		t.Fatalf("answering the restored question still copies: %+v", g.PendingChoices)
	}
}

// TestChainOfSmogCopiesItselfForTheDiscardingPlayer — the other card
// in the cycle: no sacrifice gate, and the copy goes to the player who
// was hit.
func TestChainOfSmogCopiesItselfForTheDiscardingPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	before := len(opp.Hand.Cards)

	castCatalogSpell(t, g, "Chain of Smog", "Sorcery", chainOfSmogOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	if got := len(opp.Hand.Cards); got != before-2 {
		t.Errorf("hand %d → %d, want two discarded", before, got)
	}
	answerMayChoice(t, g, opp.ID, true)

	prompt := latestPickTarget(g, opp.ID)
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
	copies := copiesOnStack(g)
	if len(copies) != 1 || copies[0].Controller != opp.ID {
		t.Fatalf("one copy controlled by the discarding player, got %+v", copies)
	}

	passPriorityAroundTable(t, g)
	// The third player may also copy; decline so the chain stops.
	if c := latestConfirmFor(g, third.ID); c != nil {
		answerMayChoice(t, g, third.ID, false)
	}
	if got := len(third.Hand.Cards); got != thirdBefore-2 {
		t.Errorf("the copy hit its new target: hand %d → %d, want two discarded", thirdBefore, got)
	}
	if got := me.Graveyard.Size(); got != 1 {
		t.Errorf("the caster's graveyard holds %d cards, want 1 (the spell only)", got)
	}
}

// copiesOnStack returns the stack items flagged IsCopy.
func copiesOnStack(g *game.Game) []game.StackItem {
	var out []game.StackItem
	for _, c := range g.Stack.Cards {
		item := g.StackItemForEffect(c.InstanceID)
		if item == nil || !item.IsCopy {
			continue
		}
		out = append(out, *item)
	}
	return out
}

func hasPlayerID(ids []uuid.UUID, want uuid.UUID) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}
