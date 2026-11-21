package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// raiding_schemes_tap_cost_test.go — #5688: conspire on a spell whose own
// cost taps permanents. The spell's own taps ride teamwork_ids and
// conspire's two creatures ride conspire_ids, so both costs are paid on
// one cast (CR 601.2b, 601.2h) with different permanents: a tapped
// creature can't be tapped again to pay a cost (CR 118.3).

// rsTapTable is tapTable with Raiding Schemes under the caster's control.
func RsTapTable(t *testing.T) (g *game.Game, me, opp *game.Player) {
	t.Helper()
	g, me, opp = TapTable(t)
	PushCatalogPermanent(g, me.ID, "Raiding Schemes", "Enchantment", RaidingSchemesOracle, false)
	return g, me, opp
}
