package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Hyperion's Atomic Vision
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// optional_behold_view_test.go — #4642: a "you may behold a Hero" offer
// ships as an ordinary optional cost with `reveal` / `behold` and the
// cards that could pay it, off the engine's own walk (#544).

const viewAtomicVisionOracle = "e1e82283-01ec-447f-b0d4-5bd4619ab3d8"

func atomicVisionOffer(t *testing.T, g *game.Game, me *game.Player) OptionalCostView {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Hyperion's Atomic Vision", TypeLine: "Instant", OracleID: viewAtomicVisionOracle,
			ManaCost: "{1}{W}", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})
	c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	if len(c.OptionalCosts) != 1 {
		t.Fatalf("optional_costs = %+v, want the one behold offer", c.OptionalCosts)
	}
	return c.OptionalCosts[0]
}

func TestOptionalBeholdOfferListsWhatCouldPayIt(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	hero := uuid.New()
	onTable := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: hero, Name: "Hero Friend", TypeLine: "Creature — Hero Human", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Creature — Goblin", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
		g.Battlefield.PushTop(game.Card{InstanceID: onTable, Name: "Hero Scout", TypeLine: "Creature — Hero", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	})
	o := atomicVisionOffer(t, g, me)
	if o.Index != 0 || o.Key != game.BeholdKey || o.Label != "Behold a Hero" || o.MaxTimes != 1 || o.ManaCost != "" {
		t.Errorf("offer = %+v, want index 0, key behold, label \"Behold a Hero\", once, no mana", o)
	}
	if !o.Reveal || !o.Behold || o.RevealOptions == nil {
		t.Fatalf("offer = %+v, want reveal, behold and options", o)
	}
	got := o.RevealOptions.Cards
	if len(got) != 2 || !slices.Contains(got, hero.String()) || !slices.Contains(got, onTable.String()) {
		t.Errorf("reveal_options = %v, want the Hero card and the Hero permanent (not the Goblin or the spell)", got)
	}
	if o.RevealOptions.Min != 1 || o.RevealOptions.Max != 1 {
		t.Errorf("reveal_options bounds = %d..%d, want exactly one", o.RevealOptions.Min, o.RevealOptions.Max)
	}
}

// Nothing to behold: present-and-empty options, which the client greys.
func TestOptionalBeholdOfferWithNothingToBeholdIsEmpty(t *testing.T) {
	g := buildActiveGame(t)
	o := atomicVisionOffer(t, g, g.Seats[0])
	if !o.Reveal || o.RevealOptions == nil {
		t.Fatalf("offer = %+v, want the offer shown with its options", o)
	}
	if len(o.RevealOptions.Cards) != 0 {
		t.Errorf("reveal_options = %v, want none", o.RevealOptions.Cards)
	}
}
