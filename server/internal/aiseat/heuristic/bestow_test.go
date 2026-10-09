package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ADR 0141 (#2862): a bestowed cast onto the bot's own creature that can
// attack now is worth more than the creature cast, and one onto an
// opponent's creature is worth less than passing. The baseline
// (BestowShare 0) prices the bestowed target as any other pick, which
// reads an opponent's creature as removed.
func TestBestowIsPricedOntoTheBotsOwnCreature(t *testing.T) {
	satyr := creature(cardID(1), 0, "Boon Satyr", 4, 2)
	satyr.TypeLine = "Enchantment Creature — Satyr"
	satyr.ManaCost = "{1}{G}{G}"
	satyr.AlternativeCosts = []protocol.AlternativeCostView{{
		Key: "bestow", Label: "Bestow {3}{G}{G}", ManaCost: "{3}{G}{G}", TargetMode: "creature",
	}}
	mine := creature(cardID(2), 0, "My Bear", 2, 2)
	theirs := creature(cardID(3), 1, "Their Bear", 2, 2)

	hard := castMove(t, 0, satyr.InstanceID, "Cast Boon Satyr")
	bestowMove := func(target, label string) legal.Move {
		m := castMove(t, 0, satyr.InstanceID, label)
		m.Params = mustJSON(t, map[string]any{"instance_id": satyr.InstanceID, "from_zone": "hand",
			"alternative_cost": "bestow", "targets": []map[string]string{cardTarget(target)}})
		return m
	}
	onMine := bestowMove(mine.InstanceID, "Cast Boon Satyr bestowed on My Bear")
	onTheirs := bestowMove(theirs.InstanceID, "Cast Boon Satyr bestowed on Their Bear")
	v := newView([]protocol.PlayerView{newSeat(0, withHand(satyr)), newSeat(1)},
		withBattlefield(mine, theirs), withTurn(9, 0, "precombat_main"))
	in := input(0, v, passMove(0), hard, onMine, onTheirs)

	pol := heuristic.NewWithConfig(heuristic.DefaultConfig())
	h, m, o := rankValue(t, pol, in, hard.Label), rankValue(t, pol, in, onMine.Label), rankValue(t, pol, in, onTheirs.Label)
	if m <= h {
		t.Errorf("bestowed on its own bear priced %.3f, the creature cast %.3f; want bestow above", m, h)
	}
	if o >= 0 {
		t.Errorf("bestowed on an opponent's bear priced %.3f; want below passing", o)
	}

	base := heuristic.NewWithConfig(heuristic.BaselineConfig())
	if bo := rankValue(t, base, in, onTheirs.Label); bo <= rankValue(t, base, in, onMine.Label) {
		t.Errorf("baseline: the opponent's bear reads as removal (%.3f) above the bot's own (%.3f), as before", bo, rankValue(t, base, in, onMine.Label))
	}
}
