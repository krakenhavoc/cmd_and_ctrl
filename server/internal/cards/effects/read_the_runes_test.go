package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// read_the_runes_test.go — the X repetitions, and the fact that each
// one is its own choice. The interesting assertions are that the two
// repetitions can be answered DIFFERENTLY and that the second question
// is built from the board the first one left behind.

const readTheRunesOracle = "74316be1-041f-4a2e-a296-0b496f90ca25"

// castReadTheRunes casts the instant for the given X.
func castReadTheRunes(t *testing.T, g *game.Game, caster uuid.UUID, x int) {
	t.Helper()
	p := g.PlayerByIDForEffect(caster)
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Read the Runes", TypeLine: "Instant",
		OracleID: readTheRunesOracle, Owner: caster, Controller: caster,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(caster, id, game.CastSpellParams{XValue: x}); err != nil {
		t.Fatalf("CastSpell Read the Runes: %v", err)
	}
}

// TestReadTheRunesDrawsXThenAsksOncePerCard walks both branches in one
// resolution: repetition one is paid with a discard, repetition two
// with a sacrifice.
func TestReadTheRunesDrawsXThenAsksOncePerCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	hand := me.Hand.Size()

	castReadTheRunes(t, g, me.ID, 2)
	passPriorityAroundTable(t, g)
	// Taken after the spell has resolved: Read the Runes is itself in
	// the graveyard by the time the first question is asked.
	graves := me.Graveyard.Size()

	// "Draw X cards" happened before any question was asked.
	if got := me.Hand.Size(); got != hand+2 {
		t.Fatalf("hand %d → %d, want +2 drawn before the first question", hand, got)
	}
	first := latestOptionPickFor(g, me.ID)
	if first == nil {
		t.Fatalf("no question was asked: %+v", g.PendingChoices)
	}
	if len(first.PickOptions) != 2 {
		t.Fatalf("options = %d, want 2 (discard, sacrifice)", len(first.PickOptions))
	}

	// Repetition one: discard.
	answerOptionPick(t, g, me.ID, 0)
	discardFromHand(t, g, me.ID)
	if got := me.Graveyard.Size(); got != graves+1 {
		t.Errorf("graveyard %d → %d after the discard branch", graves, got)
	}

	// Repetition two: sacrifice. The question is only asked once the
	// first one has been paid.
	second := latestOptionPickFor(g, me.ID)
	if second == nil {
		t.Fatalf("the second question was not asked: %+v", g.PendingChoices)
	}
	answerOptionPick(t, g, me.ID, 1)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatalf("the sacrifice branch asked nothing: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{token}); err != nil {
		t.Fatalf("ResolveChooseCards (sacrifice): %v", err)
	}

	if g.Battlefield.Contains(token) {
		t.Error("the chosen permanent was not sacrificed")
	}
	if latestOptionPickFor(g, me.ID) != nil {
		t.Error("a third question was asked for X = 2")
	}
}

// TestReadTheRunesOffersOnlyTheDiscardWithNoPermanents is CR 608.2's
// "as much as possible": an option the player cannot take is never
// offered, and the discard — the printed default — is always first.
func TestReadTheRunesOffersOnlyTheDiscardWithNoPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	castReadTheRunes(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)

	c := latestOptionPickFor(g, me.ID)
	if c == nil {
		t.Fatalf("no question was asked: %+v", g.PendingChoices)
	}
	if len(c.PickOptions) != 1 {
		t.Fatalf("options = %d, want 1 — there is nothing to sacrifice", len(c.PickOptions))
	}
	if c.PickOptions[0].Label != "Discard a card" {
		t.Errorf("the only option is %q, want the discard", c.PickOptions[0].Label)
	}
}

// TestReadTheRunesForZeroDoesNothing: X = 0 draws nothing and asks
// nothing, which is a legal way to cast it.
func TestReadTheRunesForZeroDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()

	castReadTheRunes(t, g, me.ID, 0)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d → %d, want unchanged", hand, got)
	}
	if latestOptionPickFor(g, me.ID) != nil {
		t.Error("X = 0 asked a question")
	}
}
