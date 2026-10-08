package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// replicate_energy_view_test.go — the wire half of ADR 0129 PR 4: a
// replicate cost paid in energy names the energy per payment, and its
// stepper stops at what the viewer's energy pays for (CR 118.3).

const viewReplicateEnergyOracle = "test-view-replicate-energy"

func TestReplicateEnergyOfferIsCappedByTheViewersEnergy(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	stubViewOptionalCosts(t, viewReplicateEnergyOracle, []game.AdditionalCost{
		{Optional: true, Key: game.ReplicateKey, Energy: 3, Repeat: 10, Label: "Replicate—Pay {E}{E}{E}"},
	})
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Bolt Thing", TypeLine: "Sorcery", OracleID: viewReplicateEnergyOracle,
			ManaCost: "{1}{R}", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})
	set := func(n int) {
		g.WithWriteLock(func() {
			if err := g.AddPlayerCounterForEffect(me.ID, game.CounterEnergy, n-game.PlayerEnergy(me)); err != nil {
				t.Fatalf("set energy: %v", err)
			}
		})
	}

	for _, tc := range []struct {
		energy, maxTimes int
		short            bool
	}{
		{energy: 0, maxTimes: 1, short: true},
		{energy: 2, maxTimes: 1, short: true},
		{energy: 3, maxTimes: 1},
		{energy: 7, maxTimes: 2},
		{energy: 40, maxTimes: 10},
	} {
		set(tc.energy)
		c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
		if len(c.OptionalCosts) != 1 {
			t.Fatalf("energy %d: optional_costs = %+v, want the replicate offer", tc.energy, c.OptionalCosts)
		}
		o := c.OptionalCosts[0]
		if o.Energy != 3 || o.MaxTimes != tc.maxTimes || o.EnergyShort != tc.short {
			t.Errorf("energy %d: offer = energy %d, max_times %d, energy_short %v; want 3, %d, %v",
				tc.energy, o.Energy, o.MaxTimes, o.EnergyShort, tc.maxTimes, tc.short)
		}
	}
}
