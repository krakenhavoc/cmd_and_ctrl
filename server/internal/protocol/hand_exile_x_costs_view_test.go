package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hand_exile_x_costs_view_test.go — #3175: what the client is told about
// Shining Shoal's X-bound pitch and March of Swirling Mist's hand-exile
// discount.

func pushViewHandCard(g *game.Game, p *game.Player, name, typeLine, oracle, cost string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost,
			Owner: p.ID, Controller: p.ID, KnownBy: map[uuid.UUID]bool{p.ID: true},
		})
	})
	return id
}

// The pitch offer leaves X open (CR 107.3b does not lock it: the
// alternative cost includes X), and its picker carries each white card's
// mana value under mana_value_equals_x, so the pick is the X announced.
func TestShiningShoalPitchOfferLeavesXOpenAndShipsManaValues(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	shoal := pushViewHandCard(g, me, "Shining Shoal", "Instant — Arcane", "9937dae8-e639-46ad-849b-7e7be93bbbad", "{X}{W}{W}")
	angel := pushViewHandCard(g, me, "Serra Angel", "Creature — Angel", "", "{3}{W}{W}")
	pushViewHandCard(g, me, "Blue Drake", "Creature — Drake", "", "{1}{U}")

	c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), shoal.String())
	var pitch *AlternativeCostView
	for i := range c.AlternativeCosts {
		if c.AlternativeCosts[i].Key == "pitch" {
			pitch = &c.AlternativeCosts[i]
		}
	}
	if pitch == nil {
		t.Fatalf("alternative_costs = %+v, want the pitch offered with a white card in hand", c.AlternativeCosts)
	}
	if pitch.XLockedAtZero {
		t.Error("the pitch locks X at 0, but the alternative cost includes X (CR 107.3b)")
	}
	opts := pitch.PayOptions
	if opts == nil || !opts.ManaValueEqualsX {
		t.Fatalf("pay_options = %+v, want mana_value_equals_x", opts)
	}
	if len(opts.Cards) != 1 || opts.Cards[0] != angel.String() {
		t.Errorf("pay_options cards = %v, want only the white card", opts.Cards)
	}
	if got, ok := opts.ManaValues[angel.String()]; !ok || got != 5 {
		t.Errorf("mana_values[angel] = %d (%v), want 5", got, ok)
	}
}

// March's additional cost lists the caster's blue cards, min 0: none is
// always a legal payment.
func TestMarchOfSwirlingMistCarriesItsHandExilePicker(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	march := pushViewHandCard(g, me, "March of Swirling Mist", "Instant", "debc69ea-372a-4720-838a-16856cd50b07", "{X}{U}")

	c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), march.String())
	if c.AdditionalCost == nil || c.AdditionalCost.ExileFromHandOptions == nil {
		t.Fatalf("additional_cost = %+v, want a hand-exile picker even with no blue card", c.AdditionalCost)
	}
	if got := c.AdditionalCost.ExileFromHandOptions; len(got.Cards) != 0 || got.Min != 0 {
		t.Errorf("with no blue card the picker = %+v, want min 0 and empty", got)
	}

	drake := pushViewHandCard(g, me, "Blue Drake", "Creature — Drake", "", "{1}{U}")
	pushViewHandCard(g, me, "Red Ogre", "Creature — Ogre", "", "{2}{R}")
	c = handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), march.String())
	got := c.AdditionalCost.ExileFromHandOptions
	if len(got.Cards) != 1 || got.Cards[0] != drake.String() || got.Min != 0 || got.Max != 1 {
		t.Errorf("picker = %+v, want the one blue card, 0..1", got)
	}
	if c.AdditionalCost.ExileFromHandLabel != "any number of blue cards from your hand" {
		t.Errorf("label = %q", c.AdditionalCost.ExileFromHandLabel)
	}
}

// A viewer shown the March does not learn which of its owner's hidden
// cards are blue.
func TestHandPublicAdditionalCostDropsTheHandExileOptions(t *testing.T) {
	ac := &AdditionalCostView{
		ExileFromHandLabel:   "any number of blue cards from your hand",
		ExileFromHandOptions: &LegalTargetsView{Cards: []string{"a"}},
	}
	got := handPublicCastSurface(CastSurfaceView{AdditionalCost: ac}).AdditionalCost
	if got == nil || got.ExileFromHandOptions != nil || got.ExileFromHandLabel == "" {
		t.Errorf("public additional cost = %+v, want the label without the options", got)
	}
	if ac.ExileFromHandOptions == nil {
		t.Error("the owner's own stamp lost its options")
	}
}
