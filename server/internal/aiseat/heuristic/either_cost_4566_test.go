package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// either_cost_4566_test.go — #4566: the bot answering its own "you may
// sacrifice an artifact or discard a card" (Bullseye, K'un-Lun Warrior)
// and an opponent's "Ward—Discard a card or pay {2}" (Titania). It pays
// with the branch that costs it least: a spare Treasure before a real
// spell, mana before a card. A sacrificed artifact is not a land, so the
// land floor Worms of the Earth's branch keeps does not stop it.

type eitherCostRanks struct{ first, second, decline float64 }

// eitherCostOffer ranks the bot at seat 1 answering an either payment of
// `branches`, whose first move pays branch 0 with `firstCards` and second
// pays branch 1 with `secondCards`. The bot controls one land and
// `bf`, and holds `hand`.
func eitherCostOffer(t *testing.T, branches []protocol.PayOptionView, firstCards, secondCards []string, bf, hand []protocol.CardView) eitherCostRanks {
	t.Helper()
	const id = "45664566-4566-4566-4566-456645664566"
	board := append([]protocol.CardView{land(cardID(90), 1)}, bf...)
	seats := []protocol.PlayerView{newSeat(0), newSeat(1, withHand(hand...)), newSeat(2)}
	v := newView(seats, withBattlefield(board...), withTurn(5, 1, "main1"), withChoice(protocol.PendingChoiceView{
		ID: id, Kind: "pay_unless", Chooser: seatID(1).String(), Source: cardID(1),
		PayCost:   branches[0].Words + " or " + branches[1].Words,
		PayEither: branches,
	}))
	first := map[string]any{"apply": true, "pay_option": 0}
	if len(firstCards) > 0 {
		first["card_ids"] = firstCards
	}
	second := map[string]any{"apply": true, "pay_option": 1}
	if len(secondCards) > 0 {
		second["card_ids"] = secondCards
	}
	in := input(1, v)
	in.Moves = append(in.Moves,
		choiceMove(t, 1, id, "first", first),
		choiceMove(t, 1, id, "second", second),
		choiceMove(t, 1, id, "decline", map[string]any{"apply": false}),
	)
	pol := heuristic.New()
	return eitherCostRanks{
		first:   rankValue(t, pol, in, "first"),
		second:  rankValue(t, pol, in, "second"),
		decline: rankValue(t, pol, in, "decline"),
	}
}

// A spare Treasure goes before a real card from hand, and with one land
// on the board the land floor does not stop an artifact sacrifice.
func TestABotSacrificesATreasureBeforeDiscardingASpell(t *testing.T) {
	tr := treasure(cardID(20), 1)
	card := spell(cardID(30), 1, "Expensive Answer", "{3}{U}{U}")
	r := eitherCostOffer(t, []protocol.PayOptionView{
		{Action: "sacrifice", Count: 1, Words: "sacrifice an artifact", Payable: true, Options: []string{tr.InstanceID}},
		{Action: "discard", Count: 1, Words: "discard a card", Payable: true, Options: []string{card.InstanceID}},
	}, []string{tr.InstanceID}, []string{card.InstanceID}, []protocol.CardView{tr}, []protocol.CardView{card})
	if r.first <= r.decline || r.second <= r.decline {
		t.Errorf("sacrifice %.3f, discard %.3f, decline %.3f: both branches pay over declining", r.first, r.second, r.decline)
	}
	if r.first <= r.second {
		t.Errorf("sacrifice the Treasure %.3f vs discard the spell %.3f: the Treasure must win", r.first, r.second)
	}
}

// A ward with "discard a card or pay {2}": the mana keeps the card.
func TestABotPaysAWardsManaBeforeDiscarding(t *testing.T) {
	card := spell(cardID(30), 1, "Cheap Trick", "{R}")
	r := eitherCostOffer(t, []protocol.PayOptionView{
		{Action: "discard", Count: 1, Words: "discard a card", Payable: true, Options: []string{card.InstanceID}},
		{Action: "mana", Words: "pay {2}", Payable: true},
	}, []string{card.InstanceID}, nil, nil, []protocol.CardView{card})
	if r.second <= r.first || r.second <= r.decline {
		t.Errorf("mana %.3f, discard %.3f, decline %.3f: the mana must win", r.second, r.first, r.decline)
	}
}
