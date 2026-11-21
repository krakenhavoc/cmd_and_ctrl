package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tap_alt_cost_cards_test.go — ADR 0135 §1 (#2030): alternative costs that
// tap permanents you control, against the real catalog. The tap reuses the
// activated abilities' TapOthersCost (owner decision 1): untapped, yours,
// not targeted, no summoning-sickness check (CR 302.6), paid with the spell
// on the stack.

const (
	OrimsCureOracle        = "61eacaab-625b-499c-8602-f37ad5dda138"
	SivvisValorOracle      = "e02164ba-34f8-4a5f-a05b-dd3ef3f8ceae"
	AngelicFavorOracle     = "37162c6e-ac31-4386-9d1e-76aa66070b35"
	RamosianRallyOracle    = "2731c237-eda1-4b74-a294-502d574f4435"
	LashknifeOracle        = "98d790c0-985f-44a4-b247-0951beea7637"
	PrismaticStrandsOracle = "8e414ad6-ba19-44a1-a291-2e420734a6dd"
	BattleScreechOracle    = "e73131cd-b454-405b-9539-9d777e232b9e"
	GroupProjectOracle     = "e1ce210e-2d1d-4ab4-bfb7-8e36884797fc"
	ZahidOracle            = "89b07037-64df-4e66-acd8-87d97df61e3a"
	TapAltCostKey          = "tap"
)

// tapTable is a main phase on seat 0's turn with an empty hand.
func TapTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := NewCatalogGame(t)
	AdvanceToMain(t, g)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	return g, me, g.Seats[1]
}
