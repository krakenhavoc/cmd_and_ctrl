package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// x_zero_test.go — #810's policy half.
//
// The enumerator no longer offers an {X} ability at X=0 when X is the
// whole of what it does, so on a real board this is belt and braces.
// It is worth pinning all the same, because the policy is the other
// half of the pair that kept a table spinning: `ActivateBase` priced
// every activation at a flat 0.50, passing scores 0, and a bot handed
// a free ability that does nothing therefore took it, forever. A move
// that buys nothing has to rank below passing wherever it comes from.

const soothsayingLook = "Soothsaying: {X}: Look at the top X cards of your library."

// soothsaying is a permanent whose one ability carries an {X} in its
// mana component — the `demands_x` the server already derives from the
// cost string and ships on the filtered view for the client's X
// picker, which is where the policy reads it (ADR 0033 §3: a policy
// may not import the engine).
func soothsaying(id string, controller int) protocol.CardView {
	return protocol.CardView{
		InstanceID: id, Name: "Soothsaying", TypeLine: "Enchantment",
		Owner: seatID(controller).String(), Controller: seatID(controller).String(),
		ManaCost: "{U}", KnownByYou: true,
		ActivatedAbilities: []protocol.ActivatedAbilityView{{
			Label:    "{X}: Look at the top X cards of your library, then put them back in any order.",
			ManaCost: "{X}",
			DemandsX: true,
			XSlots:   1,
		}},
	}
}

// activateX is an activate_ability move announcing `x`.
func activateX(t *testing.T, seat int, src, label string, x int) legal.Move {
	return legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(seat), Kind: legal.KindActivate,
		Label: label, Source: uuid.MustParse(src),
		Params: mustJSON(t, map[string]any{
			"source_card_id": src, "ability_index": 0, "x_value": x,
		}),
	}
}

// The bug, in one assertion: offered nothing but "pass" and a free
// X=0 activation, the bot passes.
func TestAZeroXActivationRanksBelowPassing(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(soothsaying(cardID(1), 0), land(cardID(10), 0)))
	in := input(0, v,
		passMove(0),
		activateX(t, 0, cardID(1), soothsayingLook, 0),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got == soothsayingLook {
		t.Error("the bot activated an {X} ability for X=0 — a move that costs nothing and does nothing (#810)")
	}
}

// And the control, so the fix is a rule about X=0 rather than a bot
// that has stopped using its abilities: the same ability at X=3 is
// worth more than passing.
func TestAnXActivationAboveZeroIsStillWorthTaking(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(soothsaying(cardID(1), 0), land(cardID(10), 0)))
	in := input(0, v,
		passMove(0),
		activateX(t, 0, cardID(1), soothsayingLook, 3),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != soothsayingLook {
		t.Errorf("the bot chose %q over a paid-for X=3 activation", got)
	}
}

// An ability with no {X} at all is untouched: the rule keys on the
// cost carrying an {X}, not on the x_value field being absent from
// every other payload.
func TestAnOrdinaryActivationIsNotPricedAsXZero(t *testing.T) {
	c := soothsaying(cardID(1), 0)
	c.ActivatedAbilities[0].DemandsX = false
	c.ActivatedAbilities[0].XSlots = 0
	c.ActivatedAbilities[0].ManaCost = "{1}"
	const ability = "Soothsaying: {1}: Do something worth doing."
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(c, land(cardID(10), 0)))
	in := input(0, v, passMove(0), activateX(t, 0, cardID(1), ability, 0))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != ability {
		t.Errorf("the bot chose %q over an ordinary activation whose payload simply has no X", got)
	}
}
