package game

import (
	"testing"

	"github.com/google/uuid"
)

// snow_any_color_test.go — #6161, CR 609.4b and 107.4h: "you may spend
// mana from snow sources as though it were mana of any color" (Draugr
// Necromancer, CastPermission.AnyColorFromSnow). A coloured symbol also
// takes snow mana of any colour; nonsnow mana pays as printed, and a {C}
// and an {S} are not widened.

func snowTok(color string, snow bool) ManaToken {
	k := ManaSourceLand
	if snow {
		k |= ManaSourceSnow
	}
	return ManaToken{Color: color, Source: uuid.New(), SourceKinds: k}
}

func TestSnowAnyColorWidensColouredSymbolsForSnowManaOnly(t *testing.T) {
	grant := &CastPermission{AnyColorFromSnow: true}
	cases := []struct {
		cost string
		pool []ManaToken
		want bool
	}{
		{"{1}{U}", []ManaToken{snowTok("B", true), snowTok("B", false)}, true},
		{"{1}{U}", []ManaToken{snowTok("B", false), snowTok("B", false)}, false},
		{"{U}{U}", []ManaToken{snowTok("B", true), snowTok("R", true)}, true},
		{"{U}{U}", []ManaToken{snowTok("B", true), snowTok("R", false)}, false},
		// {C} is not a colour (CR 106.1b): snow black does not pay it.
		{"{C}", []ManaToken{snowTok("B", true)}, false},
		// {S} is still paid only by snow mana (CR 107.4h).
		{"{S}", []ManaToken{snowTok("B", false)}, false},
		{"{S}", []ManaToken{snowTok("B", true)}, true},
	}
	for _, c := range cases {
		cost, err := ParseCost(c.cost)
		if err != nil {
			t.Fatalf("parse %s: %v", c.cost, err)
		}
		var pool ManaPool
		pool.AddMana(c.pool...)
		if got := pool.CanPayFor(spendAsThoughAny(grant, cost), 0, ManaSpendContext{}); got != c.want {
			t.Errorf("%s with %+v: payable %v, want %v", c.cost, c.pool, got, c.want)
		}
	}
	// Without the grant snow black never pays a blue symbol.
	cost, _ := ParseCost("{U}")
	var pool ManaPool
	pool.AddMana(snowTok("B", true))
	if pool.CanPayFor(spendAsThoughAny(&CastPermission{}, cost), 0, ManaSpendContext{}) {
		t.Error("snow black paid {U} with no grant")
	}
}
