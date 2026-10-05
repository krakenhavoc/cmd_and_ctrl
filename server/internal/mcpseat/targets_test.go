package mcpseat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// searchMove is one answer to a search prompt raised by tutor.
func searchMove(f *fakeServer, tutor uuid.UUID, choice, name string, card uuid.UUID) legal.Move {
	return legal.Move{Type: legal.TypeResolveChoice, Player: f.playerID, Kind: legal.KindChoice,
		Label: "Demonic Tutor: " + name, Source: tutor,
		Params: mustJSON(map[string]any{"choice_id": choice, "card_ids": []string{card.String()}})}
}

// #2277: a search over a library larger than the cap is fetched in full
// before the model sees it, and the cut that remains names the choice.
func TestACappedSearchIsFetchedInFullAndNamesItsChoice(t *testing.T) {
	f := newFakeServer(t)
	s := newTestSeat(t, nil)
	joinFake(t, f, s)
	tutor, choice := uuid.New(), uuid.NewString()
	v := activeView(f)
	v.PendingChoices = []protocol.PendingChoiceView{{ID: choice, Kind: "search_library", Chooser: f.playerID.String(), Source: tutor.String()}}

	var capped, full []legal.Move
	for i := 0; i < 40; i++ {
		m := searchMove(f, tutor, choice, fmt.Sprintf("Card %02d", i), uuid.New())
		if i < 5 {
			capped = append(capped, m)
		}
		full = append(full, m)
	}
	full = append(full, searchMove(f, tutor, choice, "Black Lotus", uuid.New()))
	// The whole-list request is itself capped and says so, naming the
	// source AND the choice; only the choice request returns the rest.
	f.mu.Lock()
	f.fullMoves = capped
	f.cuts = []protocol.LegalCutView{{Source: tutor.String(), Choice: choice, Cap: "per_source", Omitted: 36}}
	f.choiceMoves = map[string][]legal.Move{choice: full}
	f.mu.Unlock()
	f.setState(v, capped, true)
	tok, text := decisionWindow(t, s)
	if !strings.Contains(text, "MOVES (41)") || !strings.Contains(text, "Black Lotus") {
		t.Fatalf("the search was not fetched in full:\n%s", text)
	}
	f.mu.Lock()
	reqs := append([]protocol.LegalMovesRequestPayload(nil), f.moveReqs...)
	f.mu.Unlock()
	if len(reqs) < 2 || reqs[len(reqs)-1].Choice != choice {
		t.Errorf("requests = %+v; want the last to ask for choice %s", reqs, choice)
	}

	// Find one card by name, numbered as in the full list.
	r, _ := s.LegalMoves(context.Background(), LegalMovesInput{Match: "black lotus"})
	if !strings.Contains(r.Text, "MOVES MATCHING") || !strings.Contains(r.Text, "  40: Demonic Tutor: Black Lotus") || strings.Contains(r.Text, "Card 03") {
		t.Errorf("match:\n%s", r.Text)
	}
	if r, _ = s.Act(context.Background(), ActInput{Window: tok, Move: 40}); !strings.Contains(r.Text, "status: accepted") {
		t.Fatalf("act: %s", r.Text)
	}
}

// A capped prompt that is not a search stays capped, and its hint carries
// the choice id even though a card raised it.
func TestACutPromptNamesItsChoiceEvenWithASource(t *testing.T) {
	f := newFakeServer(t)
	s := newTestSeat(t, nil)
	joinFake(t, f, s)
	src, choice := uuid.New(), uuid.NewString()
	v := activeView(f)
	v.PendingChoices = []protocol.PendingChoiceView{{ID: choice, Kind: "choose_cards", Chooser: f.playerID.String(), Source: src.String()}}
	moves := []legal.Move{
		{Type: legal.TypeResolveChoice, Player: f.playerID, Kind: legal.KindChoice, Label: "Pick A", Source: src,
			Params: mustJSON(map[string]any{"choice_id": choice})},
		{Type: legal.TypeResolveChoice, Player: f.playerID, Kind: legal.KindChoice, Label: "Pick B", Source: src,
			Params: mustJSON(map[string]any{"choice_id": choice})},
	}
	f.mu.Lock()
	f.fullMoves = moves
	f.cuts = []protocol.LegalCutView{{Source: src.String(), Choice: choice, Cap: "per_source", Omitted: 9}}
	f.mu.Unlock()
	f.setState(v, moves, true)
	_, text := decisionWindow(t, s)
	want := fmt.Sprintf("legal_moves(choice: %q) expands it", choice)
	if !strings.Contains(text, want) {
		t.Errorf("no choice hint:\n%s", text)
	}
	f.mu.Lock()
	reqs := append([]protocol.LegalMovesRequestPayload(nil), f.moveReqs...)
	f.mu.Unlock()
	for _, r := range reqs {
		if r.Choice != "" {
			t.Errorf("a prompt that is not a search was expanded on its own: %+v", r)
		}
	}
}

func targetView(f *fakeServer, spell uuid.UUID, lt *protocol.LegalTargetsView, cl []protocol.LegalTargetsView) protocol.GameView {
	v := activeView(f)
	v.Seats[0].Hand.Cards = []protocol.CardView{{InstanceID: spell.String(), Name: "Spell", ManaCost: "{1}",
		CastSurfaceView: protocol.CastSurfaceView{LegalTargets: lt, Clauses: cl}}}
	return v
}

func castTargeting(f *fakeServer, spell uuid.UUID, label string, ts ...map[string]any) legal.Move {
	return legal.Move{Type: legal.TypeCastSpell, Player: f.playerID, Kind: legal.KindCast, Label: label, Source: spell,
		Params: mustJSON(map[string]any{"instance_id": spell, "targets": ts})}
}

// #2276: a target says which player it is, by seat and by id.
func TestAPlayerTargetSaysWhoItIs(t *testing.T) {
	f := newFakeServer(t)
	s := newTestSeat(t, nil)
	joinFake(t, f, s)
	spell := uuid.New()
	v := targetView(f, spell, nil, nil)
	moves := []legal.Move{f.pass(),
		castTargeting(f, spell, "Cast Spell targeting «Agent»", map[string]any{"kind": "player", "id": f.playerID}),
		castTargeting(f, spell, "Cast Spell targeting «Agent»", map[string]any{"kind": "player", "id": f.oppID}),
	}
	f.setState(v, moves, false)
	_, text := decisionWindow(t, s)
	if !strings.Contains(text, "targets: you (seat 0)") || !strings.Contains(text, "targets: opponent «Bob» (seat 1)") {
		t.Errorf("targets not named:\n%s", text)
	}
}

// Same display name on two seats: the note still tells them apart.
func TestTwoSeatsWithOneNameAreToldApartByIDAndSeat(t *testing.T) {
	f := newFakeServer(t)
	v := activeView(f)
	v.Seats[1].Name = v.Seats[0].Name
	m := castTargeting(f, uuid.New(), "x", map[string]any{"kind": "player", "id": f.oppID})
	if got := targetNote(&v, f.playerID.String(), m); got != "targets: opponent «Agent» (seat 1)" {
		t.Errorf("note = %q", got)
	}
	m = castTargeting(f, uuid.New(), "x", map[string]any{"kind": "player", "id": f.playerID})
	if got := targetNote(&v, f.playerID.String(), m); got != "targets: you (seat 0)" {
		t.Errorf("note = %q", got)
	}
}

// #2276: the agent picks lands from the board, a set the enumerator's
// capped combinations might not hold.
func TestTheAgentPicksTargetsFromTheBoardPerClause(t *testing.T) {
	f := newFakeServer(t)
	s := newTestSeat(t, nil)
	joinFake(t, f, s)
	spell := uuid.New()
	var lands []string
	for i := 0; i < 6; i++ {
		lands = append(lands, uuid.NewString())
	}
	lt := &protocol.LegalTargetsView{Cards: lands, Min: 0, Max: 4, Label: "up to four target lands"}
	v := targetView(f, spell, lt, nil)
	for i, id := range lands {
		ctl := f.playerID.String()
		if i%2 == 1 {
			ctl = f.oppID.String()
		}
		v.Battlefield.Cards = append(v.Battlefield.Cards, protocol.CardView{InstanceID: id, Name: fmt.Sprintf("Forest %d", i), Controller: ctl})
	}
	moves := []legal.Move{f.pass(),
		castTargeting(f, spell, "Cast Spell targeting Forest 0", map[string]any{"kind": "card", "id": lands[0]}),
	}
	f.setState(v, moves, false)
	tok, _ := decisionWindow(t, s)

	r, _ := s.LegalMoves(context.Background(), LegalMovesInput{TargetsFor: ptr(1)})
	for _, want := range []string{"up to 4", "up to four target lands", "Forest 5", "controlled by opponent «Bob» (seat 1)", lands[3]} {
		if !strings.Contains(r.Text, want) {
			t.Errorf("targets_for missing %q:\n%s", want, r.Text)
		}
	}

	// Too many, and not a candidate: refused with a reason, nothing sent.
	r, _ = s.Act(context.Background(), ActInput{Window: tok, Move: 1, Targets: []TargetPick{{IDs: append(append([]string(nil), lands[:4]...), lands[5])}}})
	if !r.IsError || !strings.Contains(r.Text, "takes up to 4") {
		t.Errorf("five picks: %s", r.Text)
	}
	r, _ = s.Act(context.Background(), ActInput{Window: tok, Move: 1, Targets: []TargetPick{{IDs: []string{uuid.NewString()}}}})
	if !r.IsError || !strings.Contains(r.Text, "not a legal target") {
		t.Errorf("a stranger: %s", r.Text)
	}
	if f.actionCount() != 0 {
		t.Fatal("a refused pick was sent")
	}

	pick := []string{lands[1], lands[2], lands[4], lands[5]}
	r, _ = s.Act(context.Background(), ActInput{Window: tok, Move: 1, Targets: []TargetPick{{IDs: pick}}})
	if !strings.Contains(r.Text, "status: accepted") {
		t.Fatalf("act: %s", r.Text)
	}
	var got struct {
		InstanceID string       `json:"instance_id"`
		Targets    []wireTarget `json:"targets"`
	}
	if err := json.Unmarshal(f.lastAction().Params, &got); err != nil {
		t.Fatal(err)
	}
	if got.InstanceID != spell.String() || len(got.Targets) != 4 {
		t.Fatalf("sent %s", f.lastAction().Params)
	}
	for i, tg := range got.Targets {
		if tg.Kind != "card" || tg.ID != pick[i] {
			t.Errorf("target %d = %+v, want card %s", i, tg, pick[i])
		}
	}
}

// Two clauses are picked per slot, and a player is sent as a player.
func TestTargetsAreAnsweredPerSlot(t *testing.T) {
	f := newFakeServer(t)
	s := newTestSeat(t, nil)
	joinFake(t, f, s)
	spell, mine, theirs := uuid.New(), uuid.NewString(), uuid.NewString()
	cl := []protocol.LegalTargetsView{
		{Cards: []string{mine}, Min: 1, Max: 1, Label: "target creature you control"},
		{Players: []string{f.oppID.String()}, Cards: []string{theirs}, Min: 1, Max: 1, Label: "target creature or player you don't control"},
	}
	v := targetView(f, spell, nil, cl)
	moves := []legal.Move{f.pass(), castTargeting(f, spell, "Cast Spell",
		map[string]any{"kind": "card", "id": mine}, map[string]any{"kind": "card", "id": theirs, "slot": 1})}
	f.setState(v, moves, false)
	tok, _ := decisionWindow(t, s)
	// A slot left out is short of its minimum.
	r, _ := s.Act(context.Background(), ActInput{Window: tok, Move: 1, Targets: []TargetPick{{Slot: 0, IDs: []string{mine}}}})
	if !r.IsError {
		t.Errorf("a missing clause was accepted: %s", r.Text)
	}
	r, _ = s.Act(context.Background(), ActInput{Window: tok, Move: 1, Targets: []TargetPick{
		{Slot: 1, IDs: []string{f.oppID.String()}}, {Slot: 0, IDs: []string{mine}}}})
	if !strings.Contains(r.Text, "status: accepted") {
		t.Fatalf("act: %s", r.Text)
	}
	var got struct{ Targets []wireTarget }
	_ = json.Unmarshal(f.lastAction().Params, &got)
	want := []wireTarget{{Kind: "card", ID: mine}, {Kind: "player", ID: f.oppID.String(), Slot: 1}}
	if len(got.Targets) != 2 || got.Targets[0] != want[0] || got.Targets[1] != want[1] {
		t.Errorf("sent %+v, want %+v", got.Targets, want)
	}
}

func TestATargetlessMoveTakesNoTargets(t *testing.T) {
	f := newFakeServer(t)
	s := newTestSeat(t, nil)
	joinFake(t, f, s)
	f.setState(activeView(f), []legal.Move{f.pass(), f.cast(uuid.New(), "Cast Wrath")}, false)
	tok, _ := decisionWindow(t, s)
	r, _ := s.Act(context.Background(), ActInput{Window: tok, Move: 1, Targets: []TargetPick{{IDs: []string{uuid.NewString()}}}})
	if !r.IsError || !strings.Contains(r.Text, "no target clause") {
		t.Errorf("act: %s", r.Text)
	}
}

// With no clause in the view, targets_for assembles the candidates from
// the window's own moves and says it may be incomplete.
func TestTargetsForFallsBackToTheWindowsMoves(t *testing.T) {
	f := newFakeServer(t)
	s := newTestSeat(t, nil)
	joinFake(t, f, s)
	spell := uuid.New()
	moves := []legal.Move{f.pass(),
		castTargeting(f, spell, "a", map[string]any{"kind": "player", "id": f.playerID}),
		castTargeting(f, spell, "b", map[string]any{"kind": "player", "id": f.oppID})}
	f.setState(activeView(f), moves, false)
	decisionWindow(t, s)
	r, _ := s.LegalMoves(context.Background(), LegalMovesInput{TargetsFor: ptr(1)})
	for _, want := range []string{"you (seat 0)", "opponent «Bob» (seat 1)", "may have capped"} {
		if !strings.Contains(r.Text, want) {
			t.Errorf("missing %q:\n%s", want, r.Text)
		}
	}
}

func ptr[T any](v T) *T { return &v }
