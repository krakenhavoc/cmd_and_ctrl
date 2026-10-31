package mcpseat

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// cut_cards_test.go — #4737. A list under the wire's 48-move cap can still
// hold a card the enumerator cut to its first twelve moves. The snapshot
// says so only in that card's legal_actions entry, so the seat reads it
// there, fetches the card in full before the model sees the window, and
// counts the window as truncated.

// removalBoard is a seat holding a "target nonland permanent" spell over
// permanents Bob controls, one per name, in battlefield order. It returns
// the view, the spell, and every permanent's id.
func removalBoard(f *fakeServer, names []string) (protocol.GameView, uuid.UUID, []uuid.UUID) {
	v := activeView(f)
	spell := uuid.New()
	v.Seats[0].Hand.Cards = []protocol.CardView{{InstanceID: spell.String(), Name: "Anguished Unmaking", ManaCost: "{1}{W}{B}",
		Owner: f.playerID.String(), Controller: f.playerID.String()}}
	ids := make([]uuid.UUID, len(names))
	for i, n := range names {
		ids[i] = uuid.New()
		v.Battlefield.Cards = append(v.Battlefield.Cards, protocol.CardView{InstanceID: ids[i].String(), Name: n,
			TypeLine: "Artifact", Owner: f.oppID.String(), Controller: f.oppID.String()})
	}
	return v, spell, ids
}

// exileMoves is the spell's cast against each permanent.
func exileMoves(f *fakeServer, v protocol.GameView, spell uuid.UUID, ids []uuid.UUID) []legal.Move {
	out := make([]legal.Move, 0, len(ids))
	for _, id := range ids {
		c, _ := findCard(&v, id.String())
		out = append(out, castTargeting(f, spell, "Cast Anguished Unmaking targeting "+c.Name, map[string]any{"kind": "card", "id": id}))
	}
	return out
}

// digestCut is the spell's legal_actions entry with the enumerator's cut.
func digestCut(spell uuid.UUID, listed int, cut protocol.LegalCutView) *protocol.LegalActionsView {
	return &protocol.LegalActionsView{Pass: true, Sources: map[string]*protocol.LegalSourceView{
		spell.String(): {Kinds: []legal.Kind{legal.KindCast}, Moves: listed, Truncated: []protocol.LegalCutView{cut}},
	}}
}

// The table from #4737: twenty legal targets, the list cut to the first
// twelve, Pestilence among the eight the enumerator left out, and no
// legal_moves_truncated because the whole list is far under 48.
func TestACardTheEnumeratorCutIsFetchedInFullUnderTheWireCap(t *testing.T) {
	f := newFakeServer(t)
	s := newTestSeat(t, nil)
	joinFake(t, f, s)
	names := []string{"Sol Ring", "Talisman of Indulgence", "Magda, Brazen Outlaw", "Arcane Signet", "Plant",
		"Frog Lizard", "Dryad of the Ilysian Grove", "Mastermind Plum", "Gix, Yawgmoth Praetor", "Dowsing Dagger",
		"Raffine, Scheming Seer", "Rivaz of the Claw", "Pestilence", "Smaug the Impenetrable", "Goldspan Dragon",
		"Smaug, Wicked Worm", "Roaming Throne", "Academy Manufactor", "Ruby Medallion", "Smothering Tithe"}
	v, spell, ids := removalBoard(f, names)
	all := exileMoves(f, v, spell, ids)
	capped := append([]legal.Move{f.pass()}, all[:12]...)
	v.LegalActions = digestCut(spell, 12, protocol.LegalCutView{Cap: "per_source", Omitted: 8})
	f.mu.Lock()
	f.sourceMoves = map[string][]legal.Move{spell.String(): all}
	f.mu.Unlock()
	f.setState(v, capped, false)

	tok, text := decisionWindow(t, s)
	if !strings.Contains(text, fmt.Sprintf("MOVES (%d)", len(names)+1)) {
		t.Fatalf("the card was not fetched in full:\n%s", text)
	}
	for i, id := range ids {
		if !strings.Contains(text, "targeting "+names[i]+" (targets: "+names[i]+" ["+id.String()+"]") {
			t.Errorf("%s [%s] is a legal target and the window does not offer it:\n%s", names[i], id, text)
		}
	}
	if strings.Contains(text, "not listed") {
		t.Errorf("a card fetched in full still says it was cut:\n%s", text)
	}
	f.mu.Lock()
	reqs := append([]protocol.LegalMovesRequestPayload(nil), f.moveReqs...)
	f.mu.Unlock()
	if n := len(reqs); n == 0 || reqs[n-1] != (protocol.LegalMovesRequestPayload{Source: spell.String()}) {
		t.Errorf("requests = %+v; want the last to ask for the spell alone", reqs)
	}
	s.mu.Lock()
	st := s.stats
	report := s.stats.report()
	s.mu.Unlock()
	if st.truncatedWindows != 1 || st.fullListRequests != 1 {
		t.Errorf("stats = %+v; the window was cut and one card was fetched", st)
	}
	if !strings.Contains(report, "truncated windows: 1;") {
		t.Errorf("the report does not count the cut window:\n%s", report)
	}

	// The thirteenth target is a move like any other.
	m := regexp.MustCompile(`(?m)^  (\d+): Cast Anguished Unmaking targeting Pestilence `).FindStringSubmatch(text)
	if m == nil {
		t.Fatalf("no Pestilence line:\n%s", text)
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		t.Fatal(err)
	}
	if r, _ := s.Act(context.Background(), ActInput{Window: tok, Move: n}); !strings.Contains(r.Text, "status: accepted") {
		t.Fatalf("act: %s", r.Text)
	}
	if a := f.lastAction(); !strings.Contains(string(a.Params), ids[12].String()) {
		t.Errorf("sent %s; want the cast targeting Pestilence %s", a.Params, ids[12])
	}
}

// A server that cannot list a card's moves leaves the cut in place, and the
// window says so with the call that expands it.
func TestACutCardTheSeatCannotFetchIsNamed(t *testing.T) {
	f := newFakeServer(t)
	f.oldServer = true
	s := newTestSeat(t, func(c *Config) { c.AckTimeout = 300e6 })
	joinFake(t, f, s)
	v, spell, ids := removalBoard(f, []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "Pestilence"})
	v.LegalActions = digestCut(spell, 12, protocol.LegalCutView{Cap: "per_source", Omitted: 1})
	f.setState(v, append([]legal.Move{f.pass()}, exileMoves(f, v, spell, ids[:12])...), false)
	_, text := decisionWindow(t, s)
	want := fmt.Sprintf(`[1 more not listed, cap per_source: legal_moves(card: %q) expands it]`, spell.String())
	if !strings.Contains(text, want) {
		t.Errorf("the cut is not named; want %s in:\n%s", want, text)
	}
	s.mu.Lock()
	st := s.stats
	s.mu.Unlock()
	if st.truncatedWindows != 1 {
		t.Errorf("stats = %+v; the window was cut", st)
	}
}

// A card whose whole list would swamp the window keeps the moves it came
// with, and its cut now says exactly how many more there are. Asking for
// the card still lists them all.
func TestACardTooBigForTheWindowStaysNamedWithItsExactCount(t *testing.T) {
	f := newFakeServer(t)
	s := newTestSeat(t, nil)
	joinFake(t, f, s)
	names := make([]string, 12+expandLineBudget+10)
	for i := range names {
		names[i] = fmt.Sprintf("Relic %03d", i)
	}
	v, spell, ids := removalBoard(f, names)
	all := exileMoves(f, v, spell, ids)
	v.LegalActions = digestCut(spell, 12, protocol.LegalCutView{Cap: "per_source", Omitted: 1, AtLeast: true})
	f.mu.Lock()
	f.sourceMoves = map[string][]legal.Move{spell.String(): all}
	f.mu.Unlock()
	f.setState(v, append([]legal.Move{f.pass()}, all[:12]...), false)
	_, text := decisionWindow(t, s)
	more := len(names) - 12
	want := fmt.Sprintf(`[%d more not listed, cap per_source: legal_moves(card: %q) expands it]`, more, spell.String())
	if !strings.Contains(text, "MOVES (13)") || !strings.Contains(text, want) {
		t.Fatalf("want the twelve and %s:\n%s", want, text)
	}
	r, _ := s.LegalMoves(context.Background(), LegalMovesInput{Card: spell.String()})
	if !strings.Contains(r.Text, fmt.Sprintf("%d moves added", more)) || !strings.Contains(r.Text, "targeting "+names[len(names)-1]) {
		t.Errorf("legal_moves(card) does not list the rest:\n%s", r.Text)
	}
}

// Past the ceiling no request lists more, so the hint names the call that
// still reaches every target: picking it.
func TestACeilingCutPointsAtPickingTheTarget(t *testing.T) {
	f := newFakeServer(t)
	s := newTestSeat(t, nil)
	joinFake(t, f, s)
	v, spell, ids := removalBoard(f, []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N"})
	all := exileMoves(f, v, spell, ids)
	v.LegalActions = digestCut(spell, 12, protocol.LegalCutView{Cap: "per_source", Omitted: 2})
	f.mu.Lock()
	f.sourceMoves = map[string][]legal.Move{spell.String(): all[:13]}
	f.sourceCuts = map[string][]protocol.LegalCutView{spell.String(): {{Source: spell.String(), Cap: "ceiling", Omitted: 1}}}
	f.mu.Unlock()
	f.setState(v, append([]legal.Move{f.pass()}, all[:12]...), false)
	_, text := decisionWindow(t, s)
	want := "[1 more not listed, past the 512 moves one request returns: legal_moves(targets_for: 1) lists every target, and act(move: 1, targets: …) picks any]"
	if !strings.Contains(text, "MOVES (14)") || !strings.Contains(text, want) {
		t.Errorf("want 13 casts and %s:\n%s", want, text)
	}
}

// The board prints 296 Treasures as "Treasure x296"; the list prints the
// moves against them as one line too, with every number they have. A
// tapped Treasure is a different entry on the board, and so on the list.
func TestMovesAgainstIdenticalPermanentsShareOneLine(t *testing.T) {
	f := newFakeServer(t)
	s := newTestSeat(t, nil)
	joinFake(t, f, s)
	v := activeView(f)
	spell := uuid.New()
	v.Seats[0].Hand.Cards = []protocol.CardView{{InstanceID: spell.String(), Name: "Void Rend", ManaCost: "{W}{U}{B}"}}
	treasure := func(tapped bool) protocol.CardView {
		return protocol.CardView{InstanceID: uuid.NewString(), Name: "Treasure", TypeLine: "Token Artifact — Treasure",
			IsToken: true, Tapped: tapped, Owner: f.oppID.String(), Controller: f.oppID.String()}
	}
	for range 10 {
		v.Battlefield.Cards = append(v.Battlefield.Cards, treasure(false))
	}
	v.Battlefield.Cards = append(v.Battlefield.Cards, protocol.CardView{InstanceID: uuid.NewString(), Name: "Pestilence",
		TypeLine: "Enchantment", Owner: f.oppID.String(), Controller: f.oppID.String()})
	for range 20 {
		v.Battlefield.Cards = append(v.Battlefield.Cards, treasure(false))
	}
	v.Battlefield.Cards = append(v.Battlefield.Cards, treasure(true), treasure(true))
	moves := []legal.Move{f.pass()}
	for _, c := range v.Battlefield.Cards {
		moves = append(moves, castTargeting(f, spell, "Cast Void Rend targeting "+c.Name, map[string]any{"kind": "card", "id": c.InstanceID}))
	}
	f.setState(v, moves, false)
	tok, text := decisionWindow(t, s)

	first := v.Battlefield.Cards[0].InstanceID
	for _, want := range []string{
		"  1: Cast Void Rend targeting Treasure (targets: Treasure [" + first + "] (battlefield, controlled by opponent «Bob» (seat 1))) [and 29 more like it, each naming another identical permanent: moves 2-10, 12-31]\n",
		"  11: Cast Void Rend targeting Pestilence (",
		"  32: Cast Void Rend targeting Treasure (targets: Treasure [" + v.Battlefield.Cards[31].InstanceID + "]",
		"[and 1 more like it, each naming another identical permanent: move 33]\n",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("want %q in:\n%s", want, text)
		}
	}
	if n := strings.Count(text, "Cast Void Rend targeting"); n != 3 {
		t.Errorf("%d cast lines; want one per distinct permanent, 3:\n%s", n, text)
	}

	// match folds nothing: it lists each one.
	r, _ := s.LegalMoves(context.Background(), LegalMovesInput{Match: "Treasure"})
	if n := strings.Count(r.Text, "Cast Void Rend targeting Treasure"); n != 32 {
		t.Errorf("match listed %d Treasure moves; want 32:\n%s", n, r.Text)
	}
	// And a folded number is a move like any other.
	if r, _ := s.Act(context.Background(), ActInput{Window: tok, Move: 20}); !strings.Contains(r.Text, "status: accepted") {
		t.Fatalf("act: %s", r.Text)
	}
	if a := f.lastAction(); !strings.Contains(string(a.Params), v.Battlefield.Cards[19].InstanceID) {
		t.Errorf("sent %s; want move 20's Treasure %s", a.Params, v.Battlefield.Cards[19].InstanceID)
	}
}

// The fold reads the whole board label: a permanent with a counter, or
// another controller's, is told apart, and a face-down one never folds.
func TestTheFoldKeepsApartWhatTheBoardKeepsApart(t *testing.T) {
	f := newFakeServer(t)
	v := activeView(f)
	me := f.playerID.String()
	spell := uuid.New()
	mk := func(name, ctl string, mut func(*protocol.CardView)) protocol.CardView {
		c := protocol.CardView{InstanceID: uuid.NewString(), Name: name, TypeLine: "Creature — Soldier", Power: 1, Toughness: 1,
			IsToken: true, Owner: ctl, Controller: ctl}
		if mut != nil {
			mut(&c)
		}
		return c
	}
	opp := f.oppID.String()
	v.Battlefield.Cards = []protocol.CardView{
		mk("Soldier", opp, nil),
		mk("Soldier", opp, nil),
		mk("Soldier", opp, func(c *protocol.CardView) { c.Counters = map[string]int{"+1/+1": 1} }),
		mk("Soldier", me, nil),
		mk("Soldier", opp, func(c *protocol.CardView) { c.IsToken = false }),
		mk("face-down", opp, func(c *protocol.CardView) { c.FaceDown = true }),
		mk("face-down", opp, func(c *protocol.CardView) { c.FaceDown = true }),
	}
	w := &window{me: me}
	for _, c := range v.Battlefield.Cards {
		w.moves = append(w.moves, castTargeting(f, spell, "Cast Spell targeting "+c.Name, map[string]any{"kind": "card", "id": c.InstanceID}))
	}
	shown, under := newFolder(&v, me, newNameWrapper(&v)).fold(w.moves, []int{0, 1, 2, 3, 4, 5, 6})
	if fmt.Sprint(shown) != "[0 2 3 4 5 6]" || fmt.Sprint(under[0]) != "[1]" {
		t.Errorf("shown %v, under %v; want only the two plain Soldiers of one controller folded", shown, under)
	}
}

// The seat's own pacing mirrors the server's: n requests in any span, so
// the card fetches of one window are never refused for coming too fast.
func TestMoveRequestsArePacedToTheServersLimit(t *testing.T) {
	const per = 300 * time.Millisecond
	l := newSpanLimit(2, per, nil)
	ctx := context.Background()
	start := time.Now()
	for range 2 {
		if err := l.wait(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if d := time.Since(start); d >= per {
		t.Errorf("the first two waited %s; they fit the span", d)
	}
	if err := l.wait(ctx); err != nil {
		t.Fatal(err)
	}
	if d := time.Since(start); d < per {
		t.Errorf("the third went after %s, inside the span of %s", d, per)
	}
	full := newSpanLimit(1, time.Hour, nil)
	if err := full.wait(ctx); err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if err := full.wait(cancelled); err == nil {
		t.Error("a full span and a cancelled context still let a request through")
	}
}
