package heuristic

import (
	"math"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// TestTappingBeforeCombatCostsTheAttackAsAttackValuePricesIt is #2690's
// consistency half: in the bot's own first main phase, tapping a
// creature that could attack costs the attack it gives up. That is now
// the attack's own price, attackValue, so Mary Read and Anne Bonny's loot
// (review game 2, seq 313) no longer pays 3.0 for an attack priced below
// zero, and into an open board it pays the damage the attack would deal.
func TestTappingBeforeCombatCostsTheAttackAsAttackValuePricesIt(t *testing.T) {
	me, them := uuid.New(), uuid.New()
	mary := protocol.CardView{
		InstanceID: uuid.NewString(), Name: "Mary Read and Anne Bonny",
		Controller: me.String(), Owner: me.String(),
		TypeLine: "Legendary Creature — Human Pirate", Power: 3, Toughness: 3, IsCommander: true,
	}
	body := func(name string, power, tough int, token bool) protocol.CardView {
		return protocol.CardView{
			InstanceID: uuid.NewString(), Name: name, Controller: them.String(), Owner: them.String(),
			TypeLine: "Creature — Soldier", Power: power, Toughness: tough, IsToken: token,
		}
	}
	view := func(theirs ...protocol.CardView) protocol.GameView {
		var v protocol.GameView
		v.Seats = []protocol.PlayerView{{ID: me.String(), Life: 34}, {ID: them.String(), Seat: 1, Life: 33}}
		v.Turn.ActiveSeat = 0
		v.Turn.Step = "precombat_main"
		v.Battlefield.Cards = append([]protocol.CardView{mary}, theirs...)
		return v
	}
	gang := view(body("Y'shtola", 2, 4, false), body("Soldier", 1, 1, true), body("Soldier", 1, 1, true), body("Soldier", 1, 1, true))
	open := view()

	cost := func(p *Policy, v protocol.GameView) float64 {
		st := p.newState(aiseat.Input{Seat: me, View: v})
		return p.tapCreatureCost(st, st.bf[mary.InstanceID])
	}
	near := func(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

	on := New()
	if got := cost(on, gang); !near(got, tappedBlocker) {
		t.Errorf("into a gang that kills her: tap cost %v, want the blocker alone, %v", got, tappedBlocker)
	}
	want := tappedBlocker + 3*on.cfg.DamageToOpponent
	if got := cost(on, open); !near(got, want) {
		t.Errorf("into an open board: tap cost %v, want the blocker plus 3 damage, %v", got, want)
	}

	cfg := DefaultConfig()
	cfg.GangAwareAttacks = false
	off := NewWithConfig(cfg)
	want = tappedBlocker + cfg.Weights.Power*3
	if got := cost(off, gang); !near(got, want) {
		t.Errorf("with GangAwareAttacks off: tap cost %v, want the old Power price, %v", got, want)
	}
}
