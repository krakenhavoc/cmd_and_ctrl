package protocol

import (
	"slices"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stack_ability_view_test.go — the view half of #1865 (ADR 0106 §1
// amendment 2026-10-16): a spell on the stack whose ability functions
// there carries it on `zone_abilities`, in every seat's frame that may
// activate it, stamped with that seat as the activator — so a discard
// option lists the VIEWER's land cards and nobody else's — and the
// viewer's digest lists it under the spell's ID.

const oracleLightningStormView = "2c0fceb6-d92d-4f4a-85a2-0af5be5bba81"

func TestStackAbilityReachesEverySeatWithItsOwnOptions(t *testing.T) {
	g := busyTable(t, 0)
	active := g.Seats[g.Turn.ActiveSeat]
	target := g.Seats[(g.Turn.ActiveSeat+1)%4]
	storm := put(active.Hand, active, game.Card{
		Name: "Lightning Storm", TypeLine: "Instant", ManaCost: "{1}{R}{R}", OracleID: oracleLightningStormView,
	})
	knowTheTable(g)
	if err := g.CastSpell(active.ID, storm, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}}, Strict: true, AutoTap: true,
	}); err != nil {
		t.Fatalf("cast Lightning Storm: %v", err)
	}
	knowTheTable(g)

	for _, viewer := range g.Seats {
		v := ViewOfGameFor(g, viewer.ID.String())
		c := cardInFrame(v, storm.String())
		if c == nil {
			t.Fatalf("%s: the storm is not in the frame", viewer.Name)
		}
		if len(c.ZoneAbilities) != 1 {
			t.Fatalf("%s: zone_abilities = %+v, want the one stack row", viewer.Name, c.ZoneAbilities)
		}
		row := c.ZoneAbilities[0]
		if !row.AnyPlayer || row.Ref != "own:0" {
			t.Errorf("%s: row %+v, want any_player own:0", viewer.Name, row)
		}
		// busyTable gives every hand three lands: the options are the
		// viewer's own, never another seat's (CR 602.1a, #1369).
		mine := map[string]bool{}
		for _, h := range viewer.Hand.Cards {
			mine[h.InstanceID.String()] = true
		}
		if len(row.DiscardCostOptions) != 3 {
			t.Errorf("%s: discard options %v, want the three land cards in hand", viewer.Name, row.DiscardCostOptions)
		}
		for _, id := range row.DiscardCostOptions {
			if !mine[id] {
				t.Errorf("%s: discard option %s is not in the viewer's hand", viewer.Name, id)
			}
		}
	}

	// The caster holds priority (CR 117.3c): their digest names the row.
	av := ViewOfGameFor(g, active.ID.String())
	if av.LegalActions == nil {
		t.Fatal("the caster got no digest")
	}
	e := av.LegalActions.Sources[storm.String()]
	if e == nil || !slices.Equal(e.Abilities, []string{"own:0"}) {
		t.Fatalf("the caster's digest entry for the storm = %+v, want abilities [own:0]", e)
	}

	// A spectator gets the card and no rows: they are nobody's to
	// activate from there.
	if c := cardInFrame(ViewOfGameFor(g, ""), storm.String()); c == nil || len(c.ZoneAbilities) != 0 {
		t.Errorf("spectator frame: %+v, want the storm with no zone_abilities", c)
	}
}
