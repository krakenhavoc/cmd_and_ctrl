package legal_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// equip_discount_test.go — #3778, the enumerator half: an equip
// discount that reads the target ("equip abilities you activate that
// target this creature cost {2} less") prices each target on its own,
// so a seat that can afford only the discounted equip is offered that
// one and no other, and the engine accepts it (#544).

const (
	oracleHelitrooper      = "a2900c79-8e55-49f9-b548-5890a753932e"
	oracleLoxodonWarhammer = "dba35ac5-7ad3-488a-a006-6b9a1d54eea5"
)

func TestEquipDiscountOffersOnlyTheAffordableTarget(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	heli := battlefieldCard(g, active, game.Card{
		Name: "Helitrooper", TypeLine: "Creature — Human Soldier",
		OracleID: oracleHelitrooper, Power: 1, Toughness: 2,
	})
	bear := battlefieldCard(g, active, game.Card{
		Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
	})
	// Equip {3}: one Plains pays it only with Helitrooper's {2} off.
	lands(g, active, "Plains", "Plains", 1)
	hammer := battlefieldCard(g, active, game.Card{
		Name: "Loxodon Warhammer", TypeLine: "Artifact — Equipment",
		OracleID: oracleLoxodonWarhammer,
	})
	advanceTo(t, g, game.StepPrecombatMain)

	acts := activationsOf(legal.EnumerateFor(g, active.ID), hammer)
	if len(acts) == 0 {
		t.Fatal("no equip offered; the {1} equip onto Helitrooper is affordable")
	}
	for _, m := range acts {
		p := string(m.Params)
		if strings.Contains(p, bear.String()) {
			t.Errorf("equip onto the bear costs {3} with one Plains and was offered: %s", p)
		}
		if !strings.Contains(p, heli.String()) {
			t.Errorf("an equip move that does not target Helitrooper: %s", p)
		}
	}
	dispatchAll(t, g, active.ID, acts)
}
