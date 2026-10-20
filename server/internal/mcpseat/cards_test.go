package mcpseat

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// sentCardIDs is the card_ids the last action carried.
func sentCardIDs(t *testing.T, f *fakeServer) []string {
	t.Helper()
	var got struct {
		CardIDs []string `json:"card_ids"`
	}
	if err := json.Unmarshal(f.lastAction().Params, &got); err != nil {
		t.Fatal(err)
	}
	return got.CardIDs
}

// #4186: with 13 cards and a hand size of 7, the cleanup discard has 1,716
// answers and the list holds only some. The agent names the six it
// discards, including a set no listed combination holds, and the seat
// checks the set before sending it.
func TestTheAgentNamesTheCardsItDiscardsToHandSize(t *testing.T) {
	f := newFakeServer(t)
	s := newTestSeat(t, nil)
	joinFake(t, f, s)
	v := activeView(f)
	v.Turn.ActiveSeat, v.Turn.Step = 0, "cleanup"
	var hand []string
	for i := 0; i < 13; i++ {
		id := uuid.NewString()
		hand = append(hand, id)
		v.Seats[0].Hand.Cards = append(v.Seats[0].Hand.Cards, protocol.CardView{InstanceID: id, Name: fmt.Sprintf("Card %02d", i)})
	}
	v.DiscardPending = map[string]int{f.playerID.String(): 6}
	// Every listed answer discards the first card, as the capped list did.
	discard := func(ids []string) legal.Move {
		label := "Discard to hand size:"
		for _, id := range ids {
			label += " " + id[:4]
		}
		return legal.Move{Type: legal.TypeDiscardSelection, Player: f.playerID, Kind: legal.KindChoice, Label: label,
			Params: mustJSON(map[string]any{"card_ids": ids})}
	}
	moves := []legal.Move{discard(hand[0:6]), discard(append([]string{hand[0]}, hand[2:7]...))}
	f.setState(v, moves, false)
	tok, text := decisionWindow(t, s)
	for _, want := range []string{"CHOOSE 6 cards (discard to hand size) from:", "Card 12 [" + hand[12] + "]", "cards: [\"<id>\", …]"} {
		if !strings.Contains(text, want) {
			t.Errorf("the window does not say %q:\n%s", want, text)
		}
	}

	keep0 := hand[7:13] // never in a listed move: it keeps the first card
	for _, tc := range []struct {
		name, want string
		ids        []string
	}{
		{"too few", "takes 6 cards, not 5", keep0[:5]},
		{"too many", "takes 6 cards, not 7", hand[6:13]},
		{"a card not in hand", "not one of the cards", append(append([]string(nil), keep0[:5]...), uuid.NewString())},
		{"a card twice", "named twice", append(append([]string(nil), keep0[:5]...), keep0[0])},
	} {
		r, _ := s.Act(context.Background(), ActInput{Window: tok, Move: 0, Cards: tc.ids})
		if !r.IsError || !strings.Contains(r.Text, tc.want) {
			t.Errorf("%s: %s", tc.name, r.Text)
		}
	}
	if f.actionCount() != 0 {
		t.Fatal("a refused set was sent")
	}

	r, _ := s.Act(context.Background(), ActInput{Window: tok, Move: 0, Cards: keep0})
	if !strings.Contains(r.Text, "status: accepted") {
		t.Fatalf("act: %s", r.Text)
	}
	if a := f.lastAction(); fmt.Sprint(a.Type) != "discard_selection" {
		t.Errorf("sent %s, want discard_selection", a.Type)
	}
	if got := sentCardIDs(t, f); !slices.Equal(got, keep0) {
		t.Errorf("card_ids = %v, want %v", got, keep0)
	}
}

// A discard prompt (discard_from_hand) takes the agent's set the same way,
// from the cards the prompt lets it choose; a move that chooses no cards
// refuses one.
func TestTheAgentNamesTheCardsADiscardPromptTakes(t *testing.T) {
	f := newFakeServer(t)
	s := newTestSeat(t, nil)
	joinFake(t, f, s)
	src, choice := uuid.New(), uuid.NewString()
	v := activeView(f)
	var hand []protocol.CardView
	for i := 0; i < 5; i++ {
		hand = append(hand, protocol.CardView{InstanceID: uuid.NewString(), Name: fmt.Sprintf("Card %d", i)})
	}
	v.Seats[0].Hand.Cards = hand
	eligible := []string{hand[1].InstanceID, hand[2].InstanceID, hand[4].InstanceID}
	v.PendingChoices = []protocol.PendingChoiceView{{ID: choice, Kind: "discard_from_hand", Chooser: f.playerID.String(),
		FromPlayer: f.playerID.String(), Count: 2, Source: src.String(), Reason: "Mind Rot", Options: hand, Eligible: eligible}}
	moves := []legal.Move{{Type: legal.TypeResolveChoice, Player: f.playerID, Kind: legal.KindChoice, Label: "Mind Rot: discard Card 1 Card 2",
		Source: src, Params: mustJSON(map[string]any{"choice_id": choice, "card_ids": eligible[:2]})},
		{Type: legal.TypeResolveChoice, Player: f.playerID, Kind: legal.KindChoice, Label: "Mind Rot: discard Card 1 Card 4",
			Source: src, Params: mustJSON(map[string]any{"choice_id": choice, "card_ids": []string{eligible[0], eligible[2]}})}}
	f.setState(v, moves, false)
	tok, text := decisionWindow(t, s)
	if !strings.Contains(text, "CHOOSE 2 cards (Mind Rot) from:") || strings.Contains(text, hand[0].InstanceID+"]") {
		t.Errorf("the CHOOSE line:\n%s", text)
	}

	r, _ := s.Act(context.Background(), ActInput{Window: tok, Move: 0, Cards: []string{eligible[0], hand[0].InstanceID}})
	if !r.IsError || !strings.Contains(r.Text, "not one of the cards") {
		t.Errorf("an ineligible card: %s", r.Text)
	}
	pick := []string{eligible[2], eligible[0]}
	r, _ = s.Act(context.Background(), ActInput{Window: tok, Move: 0, Cards: pick})
	if !strings.Contains(r.Text, "status: accepted") {
		t.Fatalf("act: %s", r.Text)
	}
	var got struct {
		ChoiceID string   `json:"choice_id"`
		CardIDs  []string `json:"card_ids"`
	}
	if err := json.Unmarshal(f.lastAction().Params, &got); err != nil {
		t.Fatal(err)
	}
	if got.ChoiceID != choice || !slices.Equal(got.CardIDs, pick) {
		t.Errorf("sent %s", f.lastAction().Params)
	}
}

func TestAMoveThatChoosesNoCardsRefusesCards(t *testing.T) {
	f := newFakeServer(t)
	s := newTestSeat(t, nil)
	joinFake(t, f, s)
	f.setState(activeView(f), []legal.Move{f.pass(), f.cast(uuid.New(), "Cast Shock")}, false)
	tok, text := decisionWindow(t, s)
	if strings.Contains(text, "CHOOSE") {
		t.Errorf("a CHOOSE line with no prompt:\n%s", text)
	}
	r, _ := s.Act(context.Background(), ActInput{Window: tok, Move: 1, Cards: []string{uuid.NewString()}})
	if !r.IsError || !strings.Contains(r.Text, "does not choose cards") {
		t.Errorf("cards on a cast: %s", r.Text)
	}
	if f.actionCount() != 0 {
		t.Fatal("a refused move was sent")
	}
}
