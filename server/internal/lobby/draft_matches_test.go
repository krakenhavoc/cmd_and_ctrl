package lobby

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/accountdelete"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/capacity"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/draft"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/drafthistory"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// withRealBots gives the stack the bot host main wires: real runners
// on real rooms.
func (s *draftStack) withRealBots(t *testing.T) *aiseat.Manager {
	t.Helper()
	bots := aiseat.NewManager(s.hub, slog.New(slog.NewTextHandler(io.Discard, nil)))
	s.lobby.SetBotHost(bots)
	t.Cleanup(bots.Shutdown)
	return bots
}

// draftEventually polls a condition that, once true, stays true.
func draftEventually(t *testing.T, what string, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !ok() {
		if time.Now().After(deadline) {
			t.Fatalf("never happened: %s", what)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// draftSeatOf is the draft seat a seat token is for.
func (s *draftStack) draftSeatOf(t *testing.T, token string) uuid.UUID {
	t.Helper()
	p, err := s.auth.Validate(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}
	return p.PlayerID
}

// pickEverything makes every pick of a draft for its people, by the
// first cards of each pack.
func pickEverything(t *testing.T, room *draft.Room, seats ...uuid.UUID) {
	t.Helper()
	for guard := 0; ; guard++ {
		if guard > 5000 {
			t.Fatal("the picks never ended")
		}
		picking := false
		for _, seat := range seats {
			v, ok := room.View(seat)
			if !ok {
				t.Fatal("a seat lost its view")
			}
			if v.State != draft.StatePicking {
				continue
			}
			picking = true
			if v.Take == 0 {
				continue
			}
			ids := make([]string, 0, v.Take)
			for _, c := range v.Pack[:v.Take] {
				ids = append(ids, c.ID)
			}
			if err := room.Pick(seat, ids); err != nil {
				t.Fatal(err)
			}
		}
		if !picking {
			return
		}
	}
}

// deckOf is a legal 40-card deck from a seat's pool.
func deckOf(t *testing.T, room *draft.Room, seat uuid.UUID) draft.LimitedDeck {
	t.Helper()
	v, _ := room.View(seat)
	if len(v.Pool) < 23 {
		t.Fatalf("pool of %d", len(v.Pool))
	}
	d := draft.LimitedDeck{Basics: map[string]int{"Forest": 17}}
	for _, c := range v.Pool[:23] {
		d.Main = append(d.Main, c.ID)
	}
	return d
}

// matchTable is the open table of a draft seat's match, once the
// lobby's watcher has opened it.
func (s *draftStack) matchTable(t *testing.T, draftID, seat uuid.UUID, round int) (*gameEntry, uuid.UUID) {
	t.Helper()
	var entry *gameEntry
	var player uuid.UUID
	draftEventually(t, "the round's table opens", func() bool {
		s.lobby.mu.Lock()
		defer s.lobby.mu.Unlock()
		for _, e := range s.lobby.games {
			ref := e.draftMatch
			if ref == nil || ref.draft != draftID || ref.round != round {
				continue
			}
			for i, id := range ref.seats {
				if id == seat {
					entry, player = e, ref.players[i]
					return true
				}
			}
		}
		return false
	})
	return entry, player
}

// concede concedes a match at its table as one of its players.
func (s *draftStack) concede(t *testing.T, entry *gameEntry, player uuid.UUID) {
	t.Helper()
	if _, _, err := entry.room.ApplyExternal(func() error { return entry.room.Game.ConcedeMatch(player) }); err != nil {
		t.Fatalf("ConcedeMatch: %v", err)
	}
}

func (s *draftStack) roundOf(room *draft.Room, seat uuid.UUID) (round int, phase draft.Phase) {
	v, ok := room.View(seat)
	if !ok {
		return 0, ""
	}
	if v.Play != nil {
		round = v.Play.Round
	}
	return round, v.Phase
}

// One person and seven bots, end to end: the visitor drafts over the
// socket, submits a deck over the socket, and plays three rounds at
// three real limited tables, each against a real bot seat holding the
// deck its draft seat built. The other six bots' matches are tossed
// for and never get a table.
func TestOnePersonAndSevenBotsPlayThreeRounds(t *testing.T) {
	s := newDraftStack(t)
	bots := s.withRealBots(t)
	lim := capacity.Limits{MaxLoad: 100, TableWeight: 2, HumanSeatWeight: 1, BotSeatWeight: 2}
	s.lobby.SetCapacity(lim, nil)

	_, out := s.call(t, http.MethodPost, "/draft", "", `{"set":"syn"}`)
	id, tok := out.Draft.ID, out.Token
	base := "/draft/" + id.String()
	conn := s.dial(t, id, tok)
	conn.next()
	s.call(t, http.MethodPost, base+"/start", tok, "")
	cur := conn.until("a pack", func(v draft.View) bool { return len(v.Pack) > 0 })
	for picks := 0; cur.State == draft.StatePicking; {
		conn.pick(cur)
		picks++
		want := picks
		cur = conn.until("the next pack or the build", func(v draft.View) bool {
			return v.State == draft.StateDone || (len(v.Picks) == want && len(v.Pack) > 0)
		})
	}
	if cur.Phase != draft.PhaseBuilding || cur.Build == nil || cur.Build.DeckMin != 40 || cur.BestOf != 1 {
		t.Fatalf("after the last pick: phase %q, build %+v", cur.Phase, cur.Build)
	}
	// No match to sit at yet.
	if resp, _ := s.call(t, http.MethodPost, base+"/match", tok, ""); resp.StatusCode != http.StatusConflict {
		t.Fatalf("POST /match while building: %d", resp.StatusCode)
	}

	// A deck that is too small is refused with a sentence for the
	// player, and the build stays open.
	short := protocol.DraftDeckPayload{Main: []string{cur.Pool[0].ID}, Basics: map[string]int{"Forest": 10}}
	if e := conn.expectError(conn.send(protocol.KindDraftDeck, short)); e.Code != protocol.CodeDraftDeckRefused || !strings.Contains(e.Message, "at least 40") {
		t.Fatalf("a short deck: %+v", e)
	}
	deck := protocol.DraftDeckPayload{Basics: map[string]int{"Forest": 17}}
	for _, c := range cur.Pool[:23] {
		deck.Main = append(deck.Main, c.ID)
	}
	conn.send(protocol.KindDraftDeck, deck)
	cur = conn.until("the first round", func(v draft.View) bool { return v.Phase == draft.PhaseRounds && v.Play != nil && v.Play.Table != nil })
	if cur.Play.Round != 1 || cur.Play.Rounds != 3 || cur.Play.Yours == nil || len(cur.Play.Matches[0]) != 4 {
		t.Fatalf("round one: %+v", cur.Play)
	}
	// The three bot matches were tossed for as the round began.
	coins := 0
	for _, m := range cur.Play.Matches[0] {
		if m.Coin && m.Done {
			coins++
		}
	}
	if coins != 3 {
		t.Fatalf("%d bot matches tossed for in round one, want 3", coins)
	}

	seat := s.draftSeatOf(t, tok)
	room := s.lobby.DraftRoom(id)
	seen := map[uuid.UUID]bool{}
	opponents := map[int]bool{}
	for round := 1; round <= draft.MatchRounds; round++ {
		// The person takes their seat the way the page does.
		req, _ := http.NewRequest(http.MethodPost, s.srv.URL+base+"/match", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		var resp *http.Response
		var seated sessionResponse
		draftEventually(t, "the round's table is open to sit at", func() bool {
			r, err := s.srv.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer r.Body.Close()
			if r.StatusCode != http.StatusOK {
				return false
			}
			resp = r
			seated = sessionResponse{}
			return json.NewDecoder(r.Body).Decode(&seated) == nil && seated.Game != nil && !seen[seated.Game.ID]
		})
		if _, ok := sessionCookie(resp); !ok {
			t.Fatal("sitting at the match set no session cookie for the game page")
		}
		meta := *seated.Game
		seen[meta.ID] = true
		if meta.Format != string(game.FormatLimited) || meta.DraftID != id.String() || len(meta.Players) != 2 || meta.Practice {
			t.Fatalf("round %d table: %+v", round, meta)
		}
		p, err := s.auth.Validate(context.Background(), seated.Token)
		if err != nil || p.Role != auth.RolePlayer || p.GameID != meta.ID || p.PlayerID != seated.PlayerID {
			t.Fatalf("the table seat's session: %+v, %v", p, err)
		}

		entry, me := s.matchTable(t, id, seat, round)
		if entry.meta.ID != meta.ID || me != seated.PlayerID {
			t.Fatalf("round %d: seated at %s as %s, the match is %s as %s", round, meta.ID, seated.PlayerID, entry.meta.ID, me)
		}
		g := entry.room.Game
		settings := g.TableSettingsSnapshot()
		if settings.EffectiveFormat() != game.FormatLimited || settings.StartingLife != 20 || g.SeatCap() != 2 {
			t.Fatalf("round %d: format %s, life %d, seats %d", round, settings.EffectiveFormat(), settings.StartingLife, g.SeatCap())
		}
		if g.CurrentState() != game.StateActive {
			t.Fatalf("round %d: the table is %s, want started", round, g.CurrentState())
		}
		// The bot seat is a real runner, and it holds a 40-card deck
		// with the rest of its pool as its sideboard.
		var bot uuid.UUID
		for _, ps := range meta.Players {
			if ps.IsBot {
				bot = ps.PlayerID
			}
		}
		if bot == uuid.Nil || len(bots.Runners(meta.ID)) != 1 {
			t.Fatalf("round %d: bot seat %s, %d runners", round, bot, len(bots.Runners(meta.ID)))
		}
		ref := entry.draftMatch
		for i, num := range ref.numbers {
			if ref.players[i] == bot {
				if opponents[num] {
					t.Fatalf("round %d pairs the person with seat %d again", round, num)
				}
				opponents[num] = true
			}
		}
		// The table counts against the cap as a table with a person and
		// a bot; the draft itself counts nothing more.
		if got, want := s.lobby.CapacityUsage().Load, lim.TableCost(1, 1); got != want {
			t.Fatalf("round %d: load %d, want %d (one table, a person and a bot)", round, got, want)
		}
		// It is unlisted, like the practice table.
		if resp, _ := s.call(t, http.MethodGet, "/games/"+meta.ID.String(), "", ""); resp.StatusCode == http.StatusOK {
			t.Fatalf("round %d: the match table answers a caller with no session", round)
		}

		// Round two the bot concedes, the others the person does.
		loser := me
		if round == 2 {
			loser = bot
		}
		s.concede(t, entry, loser)
		next := round + 1
		draftEventually(t, "the draft moves on", func() bool {
			r, phase := s.roundOf(room, seat)
			return r >= next || phase == draft.PhaseFinished
		})
	}

	last := conn.until("the end", func(v draft.View) bool { return v.Phase == draft.PhaseFinished })
	mine := last.Play.Standings[last.You]
	if mine.Wins != 1 || mine.Losses != 2 || mine.Draws != 0 {
		t.Fatalf("the person's record = %+v, want 1-2", mine)
	}
	for _, st := range last.Play.Standings {
		if st.Wins+st.Losses+st.Draws != 3 {
			t.Fatalf("seat %d has %d results after three rounds", st.Seat, st.Wins+st.Losses+st.Draws)
		}
	}
	if len(seen) != 3 || len(opponents) != 3 {
		t.Fatalf("%d tables against %d different bots, want 3 and 3", len(seen), len(opponents))
	}
	// A finished draft and its ended tables count nothing.
	draftEventually(t, "the load drains", func() bool { return s.lobby.CapacityUsage().Load == 0 })
	// A visitor's draft leaves nothing behind once it closes: no draft,
	// no table, no runner.
	if resp, _ := s.call(t, http.MethodPost, base+"/leave", tok, ""); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("leave: %d", resp.StatusCode)
	}
	if s.lobby.DraftCount() != 0 {
		t.Fatal("the draft is still live with nobody in it")
	}
	s.lobby.mu.Lock()
	tables := len(s.lobby.games)
	s.lobby.mu.Unlock()
	if tables != 0 {
		t.Fatalf("%d tables are left after the draft closed", tables)
	}
	for tableID := range seen {
		if n := len(bots.Runners(tableID)); n != 0 {
			t.Fatalf("%d bot runners are left on a closed table", n)
		}
	}
}

// Each match a draft opens counts as a table against the load cap (ADR
// 0148, ADR 0222 §9), and the draft itself as one table until then.
func TestADraftsMatchesCountAsTablesAgainstTheCap(t *testing.T) {
	s := newDraftStack(t)
	s.withRealBots(t)
	lim := capacity.Limits{MaxLoad: 100, TableWeight: 2, HumanSeatWeight: 1, BotSeatWeight: 2}
	s.lobby.SetCapacity(lim, nil)
	hostTok, _ := s.signIn(t, "Ian")
	host := s.create(t, hostTok, `{"set":"syn","seats":8}`)
	id := host.Draft.ID
	base := "/draft/" + id.String()
	var seats []uuid.UUID
	seats = append(seats, s.draftSeatOf(t, host.Token))
	for _, name := range []string{"Sam", "Gwen"} {
		_, joined := s.call(t, http.MethodPost, base+"/join", "", `{"invite":"`+host.Draft.Invite+`","name":"`+name+`"}`)
		seats = append(seats, s.draftSeatOf(t, joined.Token))
	}
	s.call(t, http.MethodPost, base+"/start", host.Token, "")
	room := s.lobby.DraftRoom(id)
	pickEverything(t, room, seats...)
	// Building: one table and its three people.
	if got, want := s.lobby.CapacityUsage().Load, lim.TableCost(3, 0); got != want {
		t.Fatalf("load %d while building, want %d", got, want)
	}
	for _, seat := range seats {
		if err := room.SubmitDeck(seat, deckOf(t, room, seat)); err != nil {
			t.Fatal(err)
		}
	}
	// Round one: two of the three people meet, the third plays a bot.
	// Two tables: two people, and a person and a bot. The three bot
	// matches have no table and cost nothing.
	want := lim.TableCost(2, 0) + lim.TableCost(1, 1)
	draftEventually(t, "both tables open", func() bool { return s.lobby.CapacityUsage().Load == want })
	s.lobby.mu.Lock()
	tables := 0
	for _, e := range s.lobby.games {
		if e.draftMatch != nil {
			tables++
		}
	}
	s.lobby.mu.Unlock()
	if tables != 2 {
		t.Fatalf("%d match tables, want 2", tables)
	}
	// The open tables hold load a new table is refused for.
	s.lobby.SetCapacity(capacity.Limits{MaxLoad: want + 1, TableWeight: 2, HumanSeatWeight: 1, BotSeatWeight: 2}, nil)
	if _, err := s.lobby.Create("One more"); err == nil {
		t.Fatal("a table was made past the cap the draft's matches fill")
	}
}

// A person may leave after a round: their seat plays on as a bot, the
// match they were in the middle of is conceded, and the draft ends when
// every person has left.
func TestLeavingADraftDuringItsRounds(t *testing.T) {
	s := newDraftStack(t)
	s.withRealBots(t)
	hostTok, _ := s.signIn(t, "Ian")
	host := s.create(t, hostTok, `{"set":"syn","seats":4}`)
	id := host.Draft.ID
	base := "/draft/" + id.String()
	_, joined := s.call(t, http.MethodPost, base+"/join", "", `{"invite":"`+host.Draft.Invite+`","name":"Sam"}`)
	a, b := s.draftSeatOf(t, host.Token), s.draftSeatOf(t, joined.Token)
	s.call(t, http.MethodPost, base+"/start", host.Token, "")
	room := s.lobby.DraftRoom(id)
	pickEverything(t, room, a, b)
	for _, seat := range []uuid.UUID{a, b} {
		if err := room.SubmitDeck(seat, deckOf(t, room, seat)); err != nil {
			t.Fatal(err)
		}
	}
	// Round one seats the two people together.
	entry, _ := s.matchTable(t, id, a, 1)
	if got, _ := s.matchTable(t, id, b, 1); got != entry {
		t.Fatal("round one did not pair the two people")
	}
	// Sam leaves in the middle of the match: it is conceded to Ian, and
	// the round is over.
	if resp, _ := s.call(t, http.MethodPost, base+"/leave", joined.Token, ""); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("leave: %d", resp.StatusCode)
	}
	draftEventually(t, "round two", func() bool { r, _ := s.roundOf(room, a); return r >= 2 })
	v, _ := room.View(a)
	if st := v.Play.Standings[v.You]; st.Wins != 1 {
		t.Fatalf("Ian's record after Sam left = %+v, want one win", st)
	}
	if !v.Players[1-v.You].Bot {
		t.Fatal("the seat Sam left is not a bot")
	}
	// Sam's token is spent.
	if resp, _ := s.call(t, http.MethodPost, base+"/match", joined.Token, ""); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("a seat that left asks for its match: %d", resp.StatusCode)
	}
	// Ian leaves between results: nobody is left and the draft closes
	// with its tables.
	s.matchTable(t, id, a, 2)
	if resp, _ := s.call(t, http.MethodPost, base+"/leave", host.Token, ""); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("leave: %d", resp.StatusCode)
	}
	if s.lobby.DraftCount() != 0 {
		t.Fatal("the draft outlived its last person")
	}
	s.lobby.mu.Lock()
	left := len(s.lobby.games)
	s.lobby.mu.Unlock()
	if left != 0 {
		t.Fatalf("%d tables outlived the draft", left)
	}
}

// What is kept of a draft (ADR 0222 §9, ADR 0149): a signed-in person's
// picks, deck and results, readable by them alone; nothing of a guest;
// and nothing at all once the person deletes their account.
func TestADraftsHistoryIsKeptForItsSignedInPeopleOnly(t *testing.T) {
	s := newDraftStack(t)
	d, err := db.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.Close() })
	store := drafthistory.NewSQLStore(d)
	s.lobby.SetDraftRecorder(store)

	hostTok, user := s.signIn(t, "Ian")
	if _, err := d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, 'Ian', 1, 1)`, user.String()); err != nil {
		t.Fatal(err)
	}
	host := s.create(t, hostTok, `{"set":"syn","seats":2,"best_of":3,"name":"Friday draft"}`)
	id := host.Draft.ID
	base := "/draft/" + id.String()
	_, joined := s.call(t, http.MethodPost, base+"/join", "", `{"invite":"`+host.Draft.Invite+`","name":"Gwen the Guest"}`)
	a, b := s.draftSeatOf(t, host.Token), s.draftSeatOf(t, joined.Token)
	s.call(t, http.MethodPost, base+"/start", host.Token, "")
	room := s.lobby.DraftRoom(id)
	pickEverything(t, room, a, b)

	// Nothing is kept while the draft runs.
	rows := func(q string, args ...any) int {
		t.Helper()
		var n int
		if err := d.QueryRow(q, args...).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	if n := rows(`SELECT COUNT(*) FROM draft_history`); n != 0 {
		t.Fatalf("%d rows before the draft is finished", n)
	}
	for _, seat := range []uuid.UUID{a, b} {
		if err := room.SubmitDeck(seat, deckOf(t, room, seat)); err != nil {
			t.Fatal(err)
		}
	}
	// Three rounds of the one pairing a two-seat draft has; the guest
	// concedes each match.
	for round := 1; round <= draft.MatchRounds; round++ {
		entry, guest := s.matchTable(t, id, b, round)
		if got := entry.room.Game.TableSettingsSnapshot().EffectiveBestOf(); got != 3 {
			t.Fatalf("the table is a best of %d, want the draft's best of three", got)
		}
		s.concede(t, entry, guest)
		next := round + 1
		draftEventually(t, "the next round", func() bool {
			r, phase := s.roundOf(room, a)
			return r >= next || phase == draft.PhaseFinished
		})
	}
	draftEventually(t, "the finished draft is saved", func() bool { return rows(`SELECT COUNT(*) FROM draft_history`) == 1 })

	// One row, the signed-in person's. The guest has none, and no
	// column anywhere holds their name.
	if n := rows(`SELECT COUNT(*) FROM draft_history WHERE user_id = ?`, user.String()); n != 1 {
		t.Fatalf("the signed-in person has %d records, want 1", n)
	}
	draftEventually(t, "all three rounds are saved", func() bool {
		return rows(`SELECT COUNT(*) FROM draft_history_matches WHERE user_id = ? AND opponent_kind = 'guest' AND opponent_user_id IS NULL AND result = 'win'`, user.String()) == 3
	})
	if n := rows(`SELECT COUNT(*) FROM draft_history_matches`); n != 3 {
		t.Fatalf("%d round rows, want the signed-in person's 3", n)
	}
	if n := rows(`SELECT COUNT(*) FROM draft_history WHERE draft_name LIKE '%Gwen%' OR picks LIKE '%Gwen%' OR basics LIKE '%Gwen%'`); n != 0 {
		t.Fatal("a stored draft names the guest")
	}

	// The person reads it back through /me/drafts, and nobody else can.
	cfg := Config{Lobby: s.lobby, Auth: s.auth, DraftHistory: store, Cards: draftTestIndex()}
	h := Handler(cfg)
	get := func(path, tok string) (int, []byte) {
		t.Helper()
		req, _ := http.NewRequest(http.MethodGet, path, nil)
		if tok != "" {
			req.Header.Set("Authorization", "Bearer "+tok)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code, rec.Body.Bytes()
	}
	code, body := get("/me/drafts", hostTok)
	var list myDraftsResponse
	if code != http.StatusOK || json.Unmarshal(body, &list) != nil || len(list.Drafts) != 1 {
		t.Fatalf("GET /me/drafts: %d %s", code, body)
	}
	got := list.Drafts[0]
	if got.DraftID != id || got.Name != "Friday draft" || got.Set != "syn" || got.Wins != 3 || got.BestOf != 3 || len(got.Picks) != 0 {
		t.Fatalf("listed draft = %+v", got)
	}
	code, body = get("/me/drafts/"+id.String(), hostTok)
	var one myDraftResponse
	if code != http.StatusOK || json.Unmarshal(body, &one) != nil {
		t.Fatalf("GET /me/drafts/{id}: %d %s", code, body)
	}
	inDeck := 0
	for _, p := range one.Draft.Picks {
		if p.InDeck {
			inDeck++
		}
	}
	if len(one.Draft.Picks) != 42 || inDeck != 23 || one.Draft.Basics["Forest"] != 17 || len(one.Draft.Matches) != 3 {
		t.Fatalf("draft page: %d picks, %d in the deck, basics %v, %d rounds", len(one.Draft.Picks), inDeck, one.Draft.Basics, len(one.Draft.Matches))
	}
	if strings.Contains(string(body), user.String()) || strings.Contains(string(body), "Gwen") {
		t.Fatalf("the draft page carries an account id or a guest's name: %s", body)
	}
	otherTok, _ := s.signIn(t, "Bea")
	if code, _ := get("/me/drafts/"+id.String(), otherTok); code != http.StatusNotFound {
		t.Fatalf("another person reading the draft: %d", code)
	}
	if code, _ := get("/me/drafts", ""); code != http.StatusUnauthorized {
		t.Fatalf("GET /me/drafts with no session: %d", code)
	}

	// Account deletion removes it.
	if _, err := accountdelete.NewSQLStore(d).Delete(context.Background(), user); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if n := rows(`SELECT (SELECT COUNT(*) FROM draft_history) + (SELECT COUNT(*) FROM draft_history_matches)`); n != 0 {
		t.Fatalf("%d draft rows are left after the account was deleted", n)
	}
}

// A guest's draft is kept only while it runs: a visitor drafts, builds
// and plays with a recorder wired, and no row is ever written.
func TestAGuestsDraftLeavesNothingBehind(t *testing.T) {
	s := newDraftStack(t)
	s.withRealBots(t)
	d, err := db.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.Close() })
	s.lobby.SetDraftRecorder(drafthistory.NewSQLStore(d))

	_, out := s.call(t, http.MethodPost, "/draft", "", `{"set":"syn","seats":2}`)
	id := out.Draft.ID
	seat := s.draftSeatOf(t, out.Token)
	s.call(t, http.MethodPost, "/draft/"+id.String()+"/start", out.Token, "")
	room := s.lobby.DraftRoom(id)
	pickEverything(t, room, seat)
	if err := room.SubmitDeck(seat, deckOf(t, room, seat)); err != nil {
		t.Fatal(err)
	}
	for round := 1; round <= draft.MatchRounds; round++ {
		entry, me := s.matchTable(t, id, seat, round)
		s.concede(t, entry, me)
		next := round + 1
		draftEventually(t, "the next round", func() bool {
			r, phase := s.roundOf(room, seat)
			return r >= next || phase == draft.PhaseFinished
		})
	}
	if resp, _ := s.call(t, http.MethodPost, "/draft/"+id.String()+"/leave", out.Token, ""); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("leave: %d", resp.StatusCode)
	}
	var n int
	if err := d.QueryRow(`SELECT (SELECT COUNT(*) FROM draft_history) + (SELECT COUNT(*) FROM draft_history_matches)`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("a guest's draft wrote %d rows (%v)", n, err)
	}
	if s.lobby.DraftCount() != 0 {
		t.Fatal("the draft is still held in memory")
	}
}

// Limited is a draft's format: nobody makes a limited table by hand,
// and a Commander table beside a draft is what it always was.
func TestALimitedTableIsOnlyOpenedByADraft(t *testing.T) {
	s := newDraftStack(t)
	hostTok, _ := s.signIn(t, "Ian")
	req, _ := http.NewRequest(http.MethodPost, s.srv.URL+"/games", strings.NewReader(`{"name":"Sneaky","format":"limited"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+hostTok)
	resp, err := s.srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest || !strings.Contains(string(raw), "draft") {
		t.Fatalf("POST /games with format limited: %d %s", resp.StatusCode, raw)
	}
	// A table made the ordinary way is Commander, and cannot be turned
	// into a limited one.
	meta, err := s.lobby.Create("Commander night")
	if err != nil {
		t.Fatal(err)
	}
	limited := game.FormatLimited
	if _, err := s.lobby.UpdateSettings(meta.ID, uuid.Nil, game.SettingsPatch{Format: &limited}); err == nil {
		t.Fatal("a Commander table was changed to limited")
	}
	settings := s.lobby.games[meta.ID].room.Game.TableSettingsSnapshot()
	if settings.EffectiveFormat() != game.FormatCommander || settings.StartingLife != 40 || settings.Format != "" {
		t.Fatalf("the Commander table's settings = %+v", settings)
	}
}
