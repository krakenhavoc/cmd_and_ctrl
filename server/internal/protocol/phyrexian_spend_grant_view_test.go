package protocol

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// phyrexian_spend_grant_view_test.go — the wire half of #1589. Under an
// any-colour or any-type grant, `cast_prices` prices a stolen
// {1}{B/P}{B/P} with its Phyrexian symbols intact rather than as {3}, and
// `phyrexian_symbols` stays 2 — so the client offers both halves of
// each symbol: pay 2 life, or pay any mana (the grant's `any_color` /
// `any_type` label says which mana counts).
func TestCastPricesKeepPhyrexianSymbolsUnderSpendGrants(t *testing.T) {
	for _, tc := range []struct {
		name string
		perm game.CastPermission
	}{
		{"any color", game.CastPermission{AnyColor: true}},
		{"any type", game.CastPermission{AnyType: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := stripTable(t)
			tc.perm.Player = me.ID
			id := exileWithGrant(t, g, exiledSpell(opp.ID, "Wire Dismember "+tc.name, "Sorcery", "{1}{B/P}{B/P}"), tc.perm)
			c := stripCard(t, g, me.ID.String(), id)
			if !c.CastableHere {
				t.Error("the holder may cast it in their main phase")
			}
			assertPrice(t, tc.name, c, "{1}{B/P}{B/P}", true, "")
			if c.PhyrexianSymbols != 2 {
				t.Errorf("phyrexian_symbols = %d, want 2", c.PhyrexianSymbols)
			}
			if c.ExilePlay == nil || !c.ExilePlay.AnyColor {
				t.Errorf("exile_play = %+v, want the any-colour label", c.ExilePlay)
			}
		})
	}
	// A {C} beside it reads as {C} under either grant: any type widens
	// it, and the badge keeps the printed price.
	g, me, opp := stripTable(t)
	colour := exileWithGrant(t, g, exiledSpell(opp.ID, "Wire Colorless Phyrexian", "Sorcery", "{C}{B/P}"),
		game.CastPermission{Player: me.ID, AnyColor: true})
	assertPrice(t, "any color {C}", stripCard(t, g, me.ID.String(), colour), "{C}{B/P}", true, "")
	anyType := exileWithGrant(t, g, exiledSpell(opp.ID, "Wire Colorless Phyrexian 2", "Sorcery", "{C}{B/P}"),
		game.CastPermission{Player: me.ID, AnyType: true})
	assertPrice(t, "any type {C}", stripCard(t, g, me.ID.String(), anyType), "{C}{B/P}", true, "")
}
