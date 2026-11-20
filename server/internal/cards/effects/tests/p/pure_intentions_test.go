package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// An opponent's discard of Pure Intentions itself: it returns to its
// owner's hand at the next end step.
func TestPureIntentionsReturnsAtTheNextEndStepWhenAnOpponentMakesYouDiscardIt(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	pi := d.HandCard(d.Victim, "Pure Intentions", "Instant — Arcane", "{W}", effectstest.PureIntentionsOracle, 0, 0)
	filler := d.Filler(d.Victim)

	d.TheyDiscard(t, pi, filler)

	if !d.Victim.Graveyard.Contains(pi) {
		t.Fatalf("before the end step the card is in %q, want the graveyard", d.ZoneOf(pi))
	}
	d.ToEndStep(t)
	if !d.Victim.Hand.Contains(pi) {
		t.Errorf("at the end step the card is in %q, want the hand", d.ZoneOf(pi))
	}
	if !d.Victim.Graveyard.Contains(filler) {
		t.Error("the other discarded card came back too")
	}
}

func TestPureIntentionsStaysInTheGraveyardWhenYouDiscardItYourself(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	pi := d.HandCard(d.Caster, "Pure Intentions", "Instant — Arcane", "{W}", effectstest.PureIntentionsOracle, 0, 0)
	filler := d.Filler(d.Caster)

	d.IDiscard(t, pi, filler)
	d.ToEndStep(t)

	if !d.Caster.Graveyard.Contains(pi) {
		t.Errorf("the card is in %q, want the graveyard", d.ZoneOf(pi))
	}
}

// Cast first, the spell returns a card an opponent's ability makes you
// discard for the rest of the turn: an opponent's Painful Quandary asks
// the caster to discard, and the card comes straight back.
func TestPureIntentionsReturnsAnOpponentCausedDiscardThisTurn(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	effectstest.CastCatalogSpell(t, d.G, "Pure Intentions", "Instant — Arcane", effectstest.PureIntentionsOracle, nil)
	effectstest.PassPriorityAroundTable(t, d.G)
	effectstest.B12Push(d.G, d.Victim.ID, "Painful Quandary", "Enchantment", effectstest.PainfulQuandaryOracle, 0, 0)
	card := d.HandCard(d.Caster, "Card A", "Sorcery", "{1}", "", 0, 0)

	effectstest.CastCatalogSpell(t, d.G, "Filler Spell", "Instant", "not-in-catalog", nil)
	effectstest.PassPriorityAroundTable(t, d.G)
	effectstest.AnswerMayChoice(t, d.G, d.Caster.ID, true)
	pick := effectstest.LatestChooseCardsFor(d.G, d.Caster.ID)
	if pick == nil {
		t.Fatalf("the discard branch asks which card: %+v", d.G.PendingChoices)
	}
	if err := d.G.ResolveChooseCards(pick.ID, d.Caster.ID, []uuid.UUID{card}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, d.G)

	if !d.Caster.Hand.Contains(card) {
		t.Errorf("the discarded card is in %q, want the hand", d.ZoneOf(card))
	}
}

// Your own discard is not returned by it.
func TestPureIntentionsLeavesYourOwnDiscardsInTheGraveyard(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	effectstest.CastCatalogSpell(t, d.G, "Pure Intentions", "Instant — Arcane", effectstest.PureIntentionsOracle, nil)
	effectstest.PassPriorityAroundTable(t, d.G)
	a := d.HandCard(d.Caster, "Card A", "Sorcery", "{1}", "", 0, 0)
	b := d.HandCard(d.Caster, "Card B", "Sorcery", "{1}", "", 0, 0)

	d.IDiscard(t, a, b)

	if !d.Caster.Graveyard.Contains(a) || !d.Caster.Graveyard.Contains(b) {
		t.Error("your own Mind Rot's discards left the graveyard")
	}
}

// The delayed trigger's condition is an opponent-caused discard by the
// trigger's controller, and nothing else.
func TestPureIntentionsCondition(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	me, them := d.Caster.ID, d.Victim.ID
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
