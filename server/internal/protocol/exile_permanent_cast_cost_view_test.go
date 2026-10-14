package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_permanent_cast_cost_view_test.go — #3402: the hand card's
// additional_cost carries the "exile a creature you control" picker, off
// the same walk the engine validates against, present-and-empty when the
// seat has no creature.

func TestSoulExchangeCarriesItsExilePicker(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Soul Exchange", TypeLine: "Sorcery", OracleID: "366ea3b4-5e7e-4693-9f4b-1e549972609e",
			ManaCost: "{B}{B}", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})

	c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	if c.AdditionalCost == nil || c.AdditionalCost.ExilePermanentOptions == nil {
		t.Fatalf("additional_cost = %+v, want an exile picker even with no creature", c.AdditionalCost)
	}
	if got := c.AdditionalCost.ExilePermanentOptions; len(got.Cards) != 0 || got.Min != 1 || got.Max != 1 {
		t.Errorf("with no creature the picker = %+v, want 1/1 and empty", got)
	}

	mine := pushViewCreature(g, me.ID, 2, false)
	tapped := pushViewCreature(g, me.ID, 2, true)
	pushViewCreature(g, opp.ID, 2, false)
	c = handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	got := map[string]bool{}
	for _, s := range c.AdditionalCost.ExilePermanentOptions.Cards {
		got[s] = true
	}
	if len(got) != 2 || !got[mine.String()] || !got[tapped.String()] {
		t.Errorf("picker offers %v, want my two creatures (a tapped one can be exiled) and not the opponent's", got)
	}
	if c.AdditionalCost.ExilePermanentLabel != "a creature you control" {
		t.Errorf("label = %q", c.AdditionalCost.ExilePermanentLabel)
	}
}
