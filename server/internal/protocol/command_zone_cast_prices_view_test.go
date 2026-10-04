package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// command_zone_cast_prices_view_test.go — #2202. A commander sits in
// the client's castable-from-other-zones strip beside the exile cards,
// and its price tag is `cast_prices[0]`, stamped for the command zone
// exactly as for exile: the engine's own total through
// game.PriceCastForEffect, so it carries the commander tax (CR 903.8)
// and every cost modifier, and the client never multiplies a cast
// count by two. Per viewer: the owner's frame gets it, nobody else's.

// seatTaxedCommander puts a vanilla commander costing `cost` in the
// seat's command zone, known to every seat as game start marks a
// commander, with `casts` prior casts from the command zone.
func seatTaxedCommander(g *game.Game, p *game.Player, cost string, casts int) uuid.UUID {
	id := uuid.New()
	knownBy := map[uuid.UUID]bool{}
	for _, s := range g.Seats {
		knownBy[s.ID] = true
	}
	p.Command.PushTop(game.Card{
		KnownBy:     knownBy,
		InstanceID:  id,
		Name:        "Taxed Commander",
		TypeLine:    "Legendary Creature — Elf Druid",
		ManaCost:    cost,
		OracleID:    "00000000-0000-0000-0000-0000000c2202",
		Power:       2,
		Toughness:   2,
		Owner:       p.ID,
		Controller:  p.ID,
		IsCommander: true,
	})
	if casts > 0 {
		if p.CommanderCasts == nil {
			p.CommanderCasts = make(map[uuid.UUID]int)
		}
		p.CommanderCasts[id] = casts
	}
	return id
}

func TestCommandZoneCastPricesCarryTheCommanderTax(t *testing.T) {
	cases := []struct {
		casts   int
		cost    string
		printed bool
	}{
		{0, "{2}{G}", true},
		{1, "{4}{G}", false},
		{2, "{6}{G}", false},
	}
	for _, tc := range cases {
		g := buildActiveGame(t)
		me := g.Seats[0]
		id := seatTaxedCommander(g, me, "{2}{G}", tc.casts)

		c := commandCardView(t, ViewOfGameFor(g, me.ID.String()), 0, id)
		if len(c.CastPrices) == 0 {
			t.Fatalf("%d casts: the owner's commander carries no cast_prices", tc.casts)
		}
		p := c.CastPrices[0]
		if p.Cost != tc.cost || p.Printed != tc.printed || p.AlternativeCost != "" {
			t.Errorf("%d casts: cast_prices[0] = %+v, want cost %s printed=%v", tc.casts, p, tc.cost, tc.printed)
		}
		if c.CastableHere {
			t.Errorf("%d casts: castable_here stays off in the command zone (the move list answers it)", tc.casts)
		}
	}
}

// A Medallion-shaped reduction comes off AFTER the tax (CR 601.2f
// applies cost modifiers to the total CR 903.8 has already raised), so
// the tag is the number the auto-tapper will tap for.
func TestCommandZoneCastPricesApplyCostModifiersOnTopOfTheTax(t *testing.T) {
	const medallion = "test-2202-emerald-medallion"
	installChargedCostModifier(t, medallion, game.CostModifier{
		Kind:  game.CostReduction,
		Label: "Green spells you cast cost {1} less to cast.",
		AppliesTo: func(q game.CostQuery) bool {
			return q.Ability == nil && q.Controller == q.Source.Controller
		},
		Amount: func(game.CostQuery) int { return 1 },
	})
	g := buildActiveGame(t)
	me := g.Seats[0]
	src := game.NewCard("Emerald Medallion", me.ID)
	src.TypeLine = "Artifact"
	src.OracleID = medallion
	src.Controller = me.ID
	g.Battlefield.PushTop(src)

	untaxed := seatTaxedCommander(g, me, "{2}{G}", 0)
	c := commandCardView(t, ViewOfGameFor(g, me.ID.String()), 0, untaxed)
	if len(c.CastPrices) == 0 || c.CastPrices[0].Cost != "{1}{G}" || c.CastPrices[0].Printed {
		t.Errorf("untaxed under the medallion: cast_prices = %+v, want {1}{G}, not printed", c.CastPrices)
	}

	g2 := buildActiveGame(t)
	me2 := g2.Seats[0]
	src2 := game.NewCard("Emerald Medallion", me2.ID)
	src2.TypeLine = "Artifact"
	src2.OracleID = medallion
	src2.Controller = me2.ID
	g2.Battlefield.PushTop(src2)
	taxed := seatTaxedCommander(g2, me2, "{2}{G}", 2)
	c2 := commandCardView(t, ViewOfGameFor(g2, me2.ID.String()), 0, taxed)
	if len(c2.CastPrices) == 0 || c2.CastPrices[0].Cost != "{5}{G}" || c2.CastPrices[0].Printed {
		t.Errorf("two casts under the medallion: cast_prices = %+v, want {5}{G} ({2} + {4} tax - {1})", c2.CastPrices)
	}
}

// One seat's price is not another's: an opponent and a spectator see
// the commander and no price.
func TestCommandZoneCastPricesReachTheOwnerOnly(t *testing.T) {
	g := buildActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	id := seatTaxedCommander(g, me, "{2}{G}", 1)

	if c := commandCardView(t, ViewOfGameFor(g, them.ID.String()), 0, id); len(c.CastPrices) > 0 {
		t.Errorf("an opponent got the owner's commander price: %+v", c.CastPrices)
	}
	if c := commandCardView(t, ViewOfGameFor(g, ""), 0, id); len(c.CastPrices) > 0 {
		t.Errorf("a spectator got the owner's commander price: %+v", c.CastPrices)
	}
}
