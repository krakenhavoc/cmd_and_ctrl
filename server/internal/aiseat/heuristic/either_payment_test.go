package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// either_payment_test.go — ADR 0187 §7 (#3575): the bot as an offeree of
// Worms of the Earth's either payment. It takes the damage while that
// leaves it at or above the life floor, sacrifices its cheapest lands
// below it, and declines when neither is worth it. It never pays against
// its own Worms.

type eitherRanks struct{ damage, sacrifice, decline float64 }

// wormsOffer ranks the bot at seat `me` answering Worms (controlled by
// seat `wormsController`) at `life`, with `lands` lands of its own.
func wormsOffer(t *testing.T, me, wormsController, life, lands int) eitherRanks {
	t.Helper()
	const id = "66666666-6666-6666-6666-666666666666"
	var bf []protocol.CardView
	var ids []string
	for i := 0; i < lands; i++ {
		l := land(cardID(10+i), me)
		bf = append(bf, l)
		ids = append(ids, l.InstanceID)
	}
	seats := []protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2)}
	seats[me] = newSeat(me, withLife(life))
	pick := ids
	if len(pick) > 2 {
		pick = pick[:2]
	}
	v := newView(seats, withBattlefield(bf...), withTurn(5, 0, "upkeep"), withChoice(protocol.PendingChoiceView{
		ID: id, Kind: "pay_unless", Chooser: seatID(me).String(), Source: cardID(1),
		PayCost: "sacrifice two lands or take 5 damage",
		PayEither: []protocol.PayOptionView{
			{Action: "sacrifice", Count: 2, Words: "sacrifice two lands", Payable: lands >= 2, Options: ids},
			{Action: "damage", Count: 5, Words: "take 5 damage", Payable: true},
		},
		AnyPlayerOffer: &protocol.AnyPlayerOfferView{
			Controller: seatID(wormsController).String(),
			Order:      []string{seatID(0).String(), seatID(1).String(), seatID(2).String()},
		},
	}))
	moves := []protocol_move{
		{"damage", map[string]any{"apply": true, "pay_option": 1}},
		{"decline", map[string]any{"apply": false}},
	}
	if lands >= 2 {
		moves = append(moves, protocol_move{"sacrifice", map[string]any{"apply": true, "pay_option": 0, "card_ids": pick}})
	}
	in := input(me, v)
	for _, m := range moves {
		in.Moves = append(in.Moves, choiceMove(t, me, id, m.label, m.params))
	}
	pol := heuristic.New()
	out := eitherRanks{damage: rankValue(t, pol, in, "damage"), decline: rankValue(t, pol, in, "decline")}
	if lands >= 2 {
		out.sacrifice = rankValue(t, pol, in, "sacrifice")
	}
	return out
}

type protocol_move struct {
	label  string
	params map[string]any
}

// Above the floor the bot takes the damage, ahead of the lands and of
// declining.
func TestABotTakesWormsDamageAboveItsLifeFloor(t *testing.T) {
	r := wormsOffer(t, 1, 0, 40, 5)
	if r.damage <= r.sacrifice || r.damage <= r.decline {
		t.Errorf("at 40 life: damage %.3f, sacrifice %.3f, decline %.3f; the damage must win", r.damage, r.sacrifice, r.decline)
	}
}

// Taking 5 from 14 leaves 9, under the floor: the damage is declined and
// the cheapest lands are sacrificed instead.
func TestABotSacrificesLandsToWormsBelowItsLifeFloor(t *testing.T) {
	r := wormsOffer(t, 1, 0, 14, 5)
	if r.damage >= r.decline {
		t.Errorf("at 14 life: damage %.3f vs decline %.3f; the damage would leave 9, under the floor", r.damage, r.decline)
	}
	if r.sacrifice <= r.decline || r.sacrifice <= r.damage {
		t.Errorf("at 14 life: sacrifice %.3f, damage %.3f, decline %.3f; the lands must win", r.sacrifice, r.damage, r.decline)
	}
}

// Exactly at the floor is still above it: 15 - 5 = 10.
func TestABotTakesWormsDamageThatLeavesItAtTheFloor(t *testing.T) {
	r := wormsOffer(t, 1, 0, 15, 5)
	if r.damage <= r.decline {
		t.Errorf("at 15 life: damage %.3f, decline %.3f; ending on 10 is at the floor", r.damage, r.decline)
	}
}

// With the damage unaffordable and no lands to spare, the bot declines.
func TestABotDeclinesWormsWhenNeitherPaymentIsWorthIt(t *testing.T) {
	r := wormsOffer(t, 1, 0, 14, 4)
	// Four lands, two sacrificed, two kept: under the floor of three.
	if r.sacrifice >= r.decline || r.damage >= r.decline {
		t.Errorf("14 life, 4 lands: sacrifice %.3f, damage %.3f, decline %.3f; the bot must decline", r.sacrifice, r.damage, r.decline)
	}
}

// It never pays against its own Worms, whatever it holds.
func TestABotNeverPaysAgainstItsOwnWorms(t *testing.T) {
	r := wormsOffer(t, 0, 0, 40, 6)
	if r.damage >= r.decline || r.sacrifice >= r.decline {
		t.Errorf("own Worms: damage %.3f, sacrifice %.3f, decline %.3f; the controller must decline", r.damage, r.sacrifice, r.decline)
	}
}
