package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Kindle the Inner Flame
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flashback_behold_view_test.go — #4689: Kindle the Inner Flame's
// "Flashback—{1}{R}, Behold three Elementals" ships on the graveyard copy
// as an ordinary offer whose `pay_options` are the Elementals the caster
// could behold (hand cards, then permanents), min and max three, marked
// `behold`. With fewer than three the offer is not stamped (#695).

const viewKindleOracle = "c7ca6371-73bb-4717-a148-0804b8f4c1e6"

func kindleGraveyardView(t *testing.T, g *game.Game, viewer, owner *game.Player, id uuid.UUID) CardView {
	t.Helper()
	v := ViewOfGameFor(g, viewer.ID.String())
	for _, s := range v.Seats {
		if s.ID != owner.ID.String() {
			continue
		}
		for _, c := range s.Graveyard.Cards {
			if c.InstanceID == id.String() {
				return c
			}
		}
	}
	t.Fatalf("Kindle is not in its owner's graveyard view")
	return CardView{}
}

func TestFlashbackBeholdOfferListsTheElementals(t *testing.T) {
	g := buildActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	all := map[uuid.UUID]bool{me.ID: true, them.ID: true}
	kindle, inHand, a, b := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: kindle, Name: "Kindle the Inner Flame", TypeLine: "Kindred Sorcery — Elemental",
			ManaCost: "{3}{R}", OracleID: viewKindleOracle, Owner: me.ID, Controller: me.ID, KnownBy: all})
		me.Hand.PushTop(game.Card{InstanceID: inHand, Name: "Flame Elemental", TypeLine: "Creature — Elemental", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Creature — Goblin", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
		g.Battlefield.PushTop(game.Card{InstanceID: a, Name: "Smoke Elemental", TypeLine: "Creature — Elemental", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
		g.Battlefield.PushTop(game.Card{InstanceID: b, Name: "Ember Elemental", TypeLine: "Creature — Elemental", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Elemental", TypeLine: "Creature — Elemental", Power: 1, Toughness: 1, Owner: them.ID, Controller: them.ID})
	})

	c := kindleGraveyardView(t, g, me, me, kindle)
	if len(c.AlternativeCosts) != 1 {
		t.Fatalf("offers = %+v, want the flashback offer", c.AlternativeCosts)
	}
	o := c.AlternativeCosts[0]
	if o.Key != "flashback" || o.Label != "Flashback—{1}{R}, Behold three Elementals" || o.ManaCost != "{1}{R}" || !o.Behold || o.PayLabel != "three Elementals" {
		t.Errorf("offer = %+v, want the flashback behold offer", o)
	}
	if o.PayOptions == nil || o.PayOptions.Min != 3 || o.PayOptions.Max != 3 {
		t.Fatalf("pay_options = %+v, want exactly three", o.PayOptions)
	}
	want := []string{inHand.String(), a.String(), b.String()}
	if got := o.PayOptions.Cards; len(got) != 3 || !slices.Equal(got, want) {
		t.Errorf("pay_options = %v, want the hand Elemental then the two permanents %v (not the Goblin, the opponent's Elemental or Kindle)", got, want)
	}
	if o.Discards || o.SacrificeOptions != nil || o.TapOptions != nil {
		t.Errorf("offer = %+v, want no discard, sacrifice or tap block", o)
	}

	// The opponent sees the printed offer on the public pile, not what the
	// caster could pay with (#1172).
	theirs := kindleGraveyardView(t, g, them, me, kindle)
	for _, alt := range theirs.AlternativeCosts {
		if alt.PayOptions != nil {
			t.Errorf("the opponent's view carries pay_options %v", alt.PayOptions.Cards)
		}
	}
}

func TestFlashbackBeholdOfferWithheldWithTwoElementals(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	kindle := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: kindle, Name: "Kindle the Inner Flame", TypeLine: "Kindred Sorcery — Elemental",
			ManaCost: "{3}{R}", OracleID: viewKindleOracle, Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Flame Elemental", TypeLine: "Creature — Elemental", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Smoke Elemental", TypeLine: "Creature — Elemental", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	})
	c := kindleGraveyardView(t, g, me, me, kindle)
	for _, o := range c.AlternativeCosts {
		if o.Key == "flashback" {
			t.Fatalf("the flashback offer is stamped with two Elementals: %+v", o)
		}
	}
}
