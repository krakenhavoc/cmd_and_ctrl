package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pure_intentions_test.go — ADR 0153 PR 2 (#3168).

const pureIntentionsOracle = "77ef82e6-16e3-44fa-9542-252060fa6c96"

// An opponent's discard of Pure Intentions itself: it returns to its
// owner's hand at the next end step.
func TestPureIntentionsReturnsAtTheNextEndStepWhenAnOpponentMakesYouDiscardIt(t *testing.T) {
	d := newDBOTable(t)
	pi := d.handCard(d.victim, "Pure Intentions", "Instant — Arcane", "{W}", pureIntentionsOracle, 0, 0)
	filler := d.filler(d.victim)

	d.theyDiscard(t, pi, filler)

	if !d.victim.Graveyard.Contains(pi) {
		t.Fatalf("before the end step the card is in %q, want the graveyard", d.zoneOf(pi))
	}
	d.toEndStep(t)
	if !d.victim.Hand.Contains(pi) {
		t.Errorf("at the end step the card is in %q, want the hand", d.zoneOf(pi))
	}
	if !d.victim.Graveyard.Contains(filler) {
		t.Error("the other discarded card came back too")
	}
}

func TestPureIntentionsStaysInTheGraveyardWhenYouDiscardItYourself(t *testing.T) {
	d := newDBOTable(t)
	pi := d.handCard(d.caster, "Pure Intentions", "Instant — Arcane", "{W}", pureIntentionsOracle, 0, 0)
	filler := d.filler(d.caster)

	d.iDiscard(t, pi, filler)
	d.toEndStep(t)

	if !d.caster.Graveyard.Contains(pi) {
		t.Errorf("the card is in %q, want the graveyard", d.zoneOf(pi))
	}
}

// Cast first, the spell returns a card an opponent's ability makes you
// discard for the rest of the turn: an opponent's Painful Quandary asks
// the caster to discard, and the card comes straight back.
func TestPureIntentionsReturnsAnOpponentCausedDiscardThisTurn(t *testing.T) {
	d := newDBOTable(t)
	castCatalogSpell(t, d.g, "Pure Intentions", "Instant — Arcane", pureIntentionsOracle, nil)
	passPriorityAroundTable(t, d.g)
	b12Push(d.g, d.victim.ID, "Painful Quandary", "Enchantment", painfulQuandaryOracle, 0, 0)
	card := d.handCard(d.caster, "Card A", "Sorcery", "{1}", "", 0, 0)

	castCatalogSpell(t, d.g, "Filler Spell", "Instant", "not-in-catalog", nil)
	passPriorityAroundTable(t, d.g)
	answerMayChoice(t, d.g, d.caster.ID, true)
	pick := latestChooseCardsFor(d.g, d.caster.ID)
	if pick == nil {
		t.Fatalf("the discard branch asks which card: %+v", d.g.PendingChoices)
	}
	if err := d.g.ResolveChooseCards(pick.ID, d.caster.ID, []uuid.UUID{card}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, d.g)

	if !d.caster.Hand.Contains(card) {
		t.Errorf("the discarded card is in %q, want the hand", d.zoneOf(card))
	}
}

// Your own discard is not returned by it.
func TestPureIntentionsLeavesYourOwnDiscardsInTheGraveyard(t *testing.T) {
	d := newDBOTable(t)
	castCatalogSpell(t, d.g, "Pure Intentions", "Instant — Arcane", pureIntentionsOracle, nil)
	passPriorityAroundTable(t, d.g)
	a := d.handCard(d.caster, "Card A", "Sorcery", "{1}", "", 0, 0)
	b := d.handCard(d.caster, "Card B", "Sorcery", "{1}", "", 0, 0)

	d.iDiscard(t, a, b)

	if !d.caster.Graveyard.Contains(a) || !d.caster.Graveyard.Contains(b) {
		t.Error("your own Mind Rot's discards left the graveyard")
	}
}

// The delayed trigger's condition is an opponent-caused discard by the
// trigger's controller, and nothing else.
func TestPureIntentionsCondition(t *testing.T) {
	d := newDBOTable(t)
	me, them := d.caster.ID, d.victim.ID
	for _, tc := range []struct {
		name string
		ev   game.Event
		want bool
	}{
		{"an opponent's spell", game.Event{Kind: game.EventDiscardCard, Actor: me, Cause: game.MoveCauseEffect, CauseController: them}, true},
		{"your own spell", game.Event{Kind: game.EventDiscardCard, Actor: me, Cause: game.MoveCauseEffect, CauseController: me}, false},
		{"a cost", game.Event{Kind: game.EventDiscardCard, Actor: me, Cause: game.MoveCauseCost, CauseController: me}, false},
		{"cleanup", game.Event{Kind: game.EventDiscardCard, Actor: me, Cause: game.MoveCauseRule}, false},
		{"another player's discard", game.Event{Kind: game.EventDiscardCard, Actor: them, Cause: game.MoveCauseEffect, CauseController: me}, false},
	} {
		if got := game.DiscardCausedByOpponentOf(tc.ev, me); got != tc.want {
			t.Errorf("%s: %v, want %v", tc.name, got, tc.want)
		}
	}
}
