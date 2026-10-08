package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ADR 0135 §4 (owner decision 5): an emerge payment is ranked by what the
// creature is worth to keep LESS the generic mana it saves, so a spent
// creature that saves six can rank ahead of one that saves nothing.
func TestAnEmergePaymentRanksByValueLessTheManaItSaves(t *testing.T) {
	sp := spell(cardID(1), 0, "Emerge Spell", "{8}")
	sp.AlternativeCosts = []protocol.AlternativeCostView{{Key: "emerge", ReducesByManaValue: true,
		SacrificeOptions: &protocol.LegalTargetsView{Min: 1, Max: 1, Cards: []string{cardID(2)}}}}
	bear := creature(cardID(2), 0, "Bear", 2, 2)
	v := newView([]protocol.PlayerView{newSeat(0, withHand(sp)), newSeat(1)},
		withBattlefield(bear), withTurn(9, 1, "precombat_main"))
	in := input(0, v, passMove(0))
	order := heuristic.New().CostFuelPrice(in)
	id := uuid.MustParse(bear.InstanceID)
	plain := order(legal.TargetCandidate{ID: id})
	saving := order(legal.TargetCandidate{ID: id, Saves: 6})
	if saving >= plain {
		t.Errorf("a Bear that saves six is priced %.3f, one that saves nothing %.3f; saving must be cheaper", saving, plain)
	}
	if more := order(legal.TargetCandidate{ID: id, Saves: 7}); more >= saving {
		t.Errorf("saving seven is priced %.3f, saving six %.3f; more saving must rank first", more, saving)
	}
}
