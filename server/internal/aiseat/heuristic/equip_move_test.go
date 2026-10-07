package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// equip_move_test.go — #2449, the policy half.
//
// A heuristic seat with Lightning Greaves (Equip {0}) moved it between
// its own two creatures for ever: every equip was priced like a pump on
// its own creature, ActivateBase plus OwnPermanentTarget, whatever the
// Greaves was already on. The CR 732 breaker named the ability and the
// runner held on it (#810), so the table stopped. Moving an Equipment
// between your own creatures buys only what the new host gains.

const greavesEquip = "Equip {0}"

// greaves is a Lightning Greaves with its equip row, as the wire sends
// it: `equip` set on the row (#2449).
func greaves(id string, controller int, opts ...cardOpt) protocol.CardView {
	c := equipment(id, controller, "Lightning Greaves", opts...)
	c.ManaCost = "{2}"
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, Ref: "own:0", Label: greavesEquip, ManaCost: "{0}", SorcerySpeed: true, Equip: true,
	}}
	return c
}

// equipMove is the activate_ability move that equips src onto target.
func equipMove(t *testing.T, seat int, src, target string) legal.Move {
	return legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(seat), Kind: legal.KindActivate,
		Label: greavesEquip + " → " + target, Source: uuid.MustParse(src),
		Params: mustJSON(t, map[string]any{
			"source_card_id": src, "ability_index": 0,
			"targets": []map[string]string{{"kind": "card", "id": target}},
		}),
	}
}

// The board the soak stalled on, near enough: the Greaves on one of the
// bot's creatures (which shows its haste and shroud), another creature
// no better than it, and the equip offered onto either.
func greavesBoard(hostOpts ...cardOpt) protocol.GameView {
	host := creature(cardID(1), 0, "Ray Fillet, Wave Warrior", 0, 2, append([]cardOpt{keywords("flying", "haste", "shroud")}, hostOpts...)...)
	return newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(
			host,
			creature(cardID(2), 0, "Covetous Castaway", 1, 3),
			greaves(cardID(3), 0, attachedToCard(cardID(1))),
			land(cardID(10), 0),
		))
}

// The bug: offered the pass and the equip onto either creature, the bot
// passes. Before #2449 it took one of the equips, and from the other
// side it took the one back.
func TestTheBotDoesNotMoveAFreeEquipmentForNothing(t *testing.T) {
	in := input(0, greavesBoard(),
		passMove(0),
		equipMove(t, 0, cardID(3), cardID(2)),
		equipMove(t, 0, cardID(3), cardID(1)),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != passMove(0).Label {
		t.Errorf("the bot chose %q: moving the Greaves to a creature no better than its host buys nothing (#2449)", got)
	}
}

// CR 701.3b: re-equipping the creature it is already on does nothing,
// even when it is the only equip on offer.
func TestTheBotDoesNotReEquipTheSameCreature(t *testing.T) {
	in := input(0, greavesBoard(),
		passMove(0),
		equipMove(t, 0, cardID(3), cardID(1)),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != passMove(0).Label {
		t.Errorf("the bot chose %q: equipping the Greaves' own host does nothing (CR 701.3b)", got)
	}
}

// The control: a creature plainly better than the host is still worth
// moving the Equipment to, so the rule is about moves that buy nothing
// rather than a bot that has stopped equipping.
func TestTheBotStillMovesAnEquipmentToABetterHost(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(
			creature(cardID(1), 0, "Ray Fillet, Wave Warrior", 0, 2, keywords("flying", "haste", "shroud")),
			creature(cardID(2), 0, "Craw Wurm", 6, 4),
			greaves(cardID(3), 0, attachedToCard(cardID(1))),
			land(cardID(10), 0),
		))
	move := equipMove(t, 0, cardID(3), cardID(2))
	in := input(0, v, passMove(0), move)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != move.Label {
		t.Errorf("the bot chose %q over moving the Greaves from a 0/2 to a 6/4", got)
	}

	// And it settles there: from the Wurm, the move back is declined.
	back := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(
			creature(cardID(1), 0, "Ray Fillet, Wave Warrior", 0, 2, keywords("flying")),
			creature(cardID(2), 0, "Craw Wurm", 6, 4, keywords("haste", "shroud")),
			greaves(cardID(3), 0, attachedToCard(cardID(2))),
			land(cardID(10), 0),
		))
	in = input(0, back, passMove(0), equipMove(t, 0, cardID(3), cardID(1)))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != passMove(0).Label {
		t.Errorf("the bot chose %q: it moved the Greaves back off the better host", got)
	}
}

// The first equip of an unattached Equipment is not a lateral move and
// keeps the price it always had.
func TestTheBotStillEquipsAnUnattachedEquipment(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(
			creature(cardID(1), 0, "Ray Fillet, Wave Warrior", 0, 2, keywords("flying")),
			greaves(cardID(3), 0),
			land(cardID(10), 0),
		))
	move := equipMove(t, 0, cardID(3), cardID(1))
	in := input(0, v, passMove(0), move)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != move.Label {
		t.Errorf("the bot chose %q over equipping an unattached Greaves", got)
	}
}
