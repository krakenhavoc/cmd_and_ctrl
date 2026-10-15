package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// any_player_each_test.go — ADR 0187 §7 (#3573): the bot as the last
// player of a per-land any-player offer (Cleansing, 1 life a land). Alice
// (seat 0) chose her own land and the bot's first land; Bob (seat 1)
// controls the effect and chose nothing; the bot (seat 2) answers now.

const eachChoiceID = "55555555-5555-5555-5555-555555555555"

type eachSet struct {
	label string
	ids   []string
}

// eachValues ranks the sets as the bot at `life` and returns label→value.
func eachValues(t *testing.T, life int, sets ...eachSet) map[string]float64 {
	t.Helper()
	pol := heuristic.New()
	lands := []protocol.CardView{
		land(cardID(1), 0),                     // Alice's
		land(cardID(2), 2), land(cardID(3), 2), // mine; Alice chose cardID(2)
		land(cardID(4), 2), land(cardID(5), 2), // mine, nobody chose
		land(cardID(6), 1), // Bob's
	}
	// The bot's third land is tapped, which the policy prices below an
	// untapped one.
	lands[2].Tapped = true
	ch := protocol.PendingChoiceView{
		ID: eachChoiceID, Kind: "choose_cards", Chooser: seatID(2).String(), FromPlayer: seatID(2).String(),
		ChooseMin: 0, ChooseMax: 5, Options: lands,
		AnyPlayerOffer: &protocol.AnyPlayerOfferView{
			Controller: seatID(1).String(),
			Order:      []string{seatID(0).String(), seatID(1).String(), seatID(2).String()},
			Declined:   []string{seatID(1).String()},
			EachLife:   1,
		},
		EarlierChoices: []protocol.EarlierChoiceView{
			{PlayerID: seatID(0).String(), CardIDs: []string{cardID(1), cardID(2)}},
			{PlayerID: seatID(1).String(), CardIDs: []string{}},
		},
	}
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2, withLife(life))},
		withBattlefield(lands...), withChoice(ch))
	moves := make([]legal.Move, 0, len(sets))
	for _, s := range sets {
		moves = append(moves, choiceMove(t, 2, eachChoiceID, s.label, map[string]any{"card_ids": s.ids}))
	}
	in := input(2, v, moves...)
	out := map[string]float64{}
	for _, s := range sets {
		out[s.label] = rankValue(t, pol, in, s.label)
	}
	return out
}

func TestABotPaysForItsOwnLandsNobodyElseChose(t *testing.T) {
	got := eachValues(t, 40,
		eachSet{"none", nil},
		eachSet{"mine unchosen", []string{cardID(3), cardID(4)}},
		eachSet{"one", []string{cardID(3)}},
	)
	if got["mine unchosen"] <= got["none"] {
		t.Errorf("own unchosen lands: %.3f, none %.3f; the bot must pay for them", got["mine unchosen"], got["none"])
	}
	if got["mine unchosen"] <= got["one"] {
		t.Errorf("two lands %.3f, one land %.3f; the bot saves as many as it can afford", got["mine unchosen"], got["one"])
	}
}

func TestABotNeverPaysForALandAnEarlierPlayerPaysFor(t *testing.T) {
	got := eachValues(t, 40,
		eachSet{"none", nil},
		eachSet{"already paid", []string{cardID(2)}},
		eachSet{"mixed", []string{cardID(2), cardID(3)}},
	)
	for _, l := range []string{"already paid", "mixed"} {
		if got[l] >= got["none"] {
			t.Errorf("%s: %.3f, none %.3f; the bot must not pay twice for a land", l, got[l], got["none"])
		}
	}
}

func TestABotNeverPaysForAnOpponentsLand(t *testing.T) {
	got := eachValues(t, 40,
		eachSet{"none", nil},
		eachSet{"alice's", []string{cardID(1)}},
		eachSet{"bob's", []string{cardID(6)}},
		eachSet{"mixed", []string{cardID(3), cardID(6)}},
	)
	for _, l := range []string{"alice's", "bob's", "mixed"} {
		if got[l] >= got["none"] {
			t.Errorf("%s: %.3f, none %.3f; the bot must not pay for an opponent's land", l, got[l], got["none"])
		}
	}
}

// At 12 life two lands leave the bot on the floor (10) and a third takes
// it under, so the best answer is two lands and three is worse than none.
func TestABotStopsPayingAtTheLifeFloor(t *testing.T) {
	got := eachValues(t, 12,
		eachSet{"none", nil},
		eachSet{"two", []string{cardID(3), cardID(4)}},
		eachSet{"three", []string{cardID(3), cardID(4), cardID(5)}},
	)
	if got["two"] <= got["none"] {
		t.Errorf("two lands at 12 life: %.3f, none %.3f; the floor allows two", got["two"], got["none"])
	}
	if got["three"] >= got["none"] {
		t.Errorf("three lands at 12 life: %.3f, none %.3f; the third goes under the floor", got["three"], got["none"])
	}
	if low := eachValues(t, 10, eachSet{"none", nil}, eachSet{"one", []string{cardID(3)}}); low["one"] >= low["none"] {
		t.Errorf("one land at 10 life: %.3f, none %.3f; the bot is already on the floor", low["one"], low["none"])
	}
}

// With room for one land, the bot pays for the most valuable of its own.
func TestABotPaysForItsMostValuableLandFirst(t *testing.T) {
	got := eachValues(t, 11,
		eachSet{"tapped", []string{cardID(3)}},
		eachSet{"untapped", []string{cardID(4)}},
	)
	if got["untapped"] <= got["tapped"] {
		t.Errorf("an untapped land %.3f, a tapped one %.3f; the bot saves the better land first", got["untapped"], got["tapped"])
	}
}
