package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// waterbend_grant_view_test.go — #3989: a cast permission priced by
// waterbending the card's mana value (Hama, the Bloodbender). The view's
// tap picker, the enumerator and the engine read one tap cost and one
// price.

func stubWaterbendTapCost(t *testing.T) {
	t.Helper()
	prev := game.WaterbendTapCost
	game.WaterbendTapCost = waterbendClause
	t.Cleanup(func() { game.WaterbendTapCost = prev })
}

// seatWaterbendGrant exiles an instant costing {2}{U} (mana value 3) and
// gives `holder` a waterbend-priced cast of it.
func seatWaterbendGrant(t *testing.T, g *game.Game, holder uuid.UUID) uuid.UUID {
	t.Helper()
	c := game.NewCard("Exiled Insight", holder)
	c.TypeLine, c.ManaCost = "Instant", "{2}{U}"
	c.KnownBy = map[uuid.UUID]bool{}
	for _, p := range g.Seats {
		c.KnownBy[p.ID] = true
	}
	id := c.InstanceID
	g.WithWriteLock(func() {
		g.Exile.PushTop(c)
		if !g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
			Player: holder, Zone: game.ZoneExile, Scope: game.ScopeCards, CastOnly: true,
			WaterbendManaValue: true, Duration: game.WhileInZoneDuration(),
		}) {
			t.Fatal("the grant was refused")
		}
	})
	return id
}

func TestWaterbendGrantShipsTheTapPickerSizedToTheManaValue(t *testing.T) {
	stubWaterbendTapCost(t)
	g := buildActiveGame(t)
	castWindowOpen(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	card := seatWaterbendGrant(t, g, me.ID)
	mine := seatCreature(g, me.ID, "Helper", 1, false, true)
	g.BumpLayerVersionForTest()

	c := exileCardView(t, ViewOfGameFor(g, me.ID.String()), card)
	if c.TapCost == nil {
		t.Fatal("no tap_cost on a card castable by waterbending")
	}
	if c.TapCost.Key != "waterbend" || c.TapCost.Label != "Waterbend {3}" || c.TapCost.Max != 3 || c.TapCost.DemandsX {
		t.Errorf("tap_cost = %+v, want waterbend {3}, max 3, no X", c.TapCost)
	}
	if !containsID(c.TapCost.Options, mine) {
		t.Errorf("options %v do not offer the holder's creature", c.TapCost.Options.Cards)
	}
	if c.ExilePlay == nil || c.ExilePlay.CostOverride != "" {
		t.Errorf("exile_play = %+v: the waterbend is not a flat mana override", c.ExilePlay)
	}
}

// The enumerator offers the cast priced at the waterbend {3} in mana,
// not the printed {2}{U}: three colourless pay for it, two do not.
func TestWaterbendGrantIsEnumeratedAtTheWaterbendPrice(t *testing.T) {
	stubWaterbendTapCost(t)
	g := buildActiveGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	castWindowOpen(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	card := seatWaterbendGrant(t, g, me.ID)
	offered := func() bool {
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Kind == legal.KindCast && m.Source == card {
				return true
			}
		}
		return false
	}
	g.WithWriteLock(func() {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	})
	if offered() {
		t.Fatal("offered with two mana for a waterbend {3}")
	}
	g.WithWriteLock(func() { me.ManaPool.AddMana(game.ManaToken{Color: "C"}) })
	if !offered() {
		t.Fatal("not offered with three colourless mana: the printed {2}{U} was charged, or the {3} was not")
	}
}
