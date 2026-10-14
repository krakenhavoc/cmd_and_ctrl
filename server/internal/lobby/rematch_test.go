package lobby

// rematch_test.go: #3285, the host of a finished table offers the same
// game again (ADR 0164). The owner's four decisions are each pinned here:
// host-only, random turn order (the opening roll, as any new game),
// expiry when every other human has left the results screen, and a
// private table.

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// fakePresence scripts who is on a game's results screen.
type fakePresence struct {
	mu      sync.Mutex
	present map[uuid.UUID]bool
}

func (f *fakePresence) ConnectedSeats(uuid.UUID) map[uuid.UUID]bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make(map[uuid.UUID]bool, len(f.present))
	for k, v := range f.present {
		out[k] = v
	}
	return out
}

func (f *fakePresence) set(id uuid.UUID, here bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.present == nil {
		f.present = map[uuid.UUID]bool{}
	}
	f.present[id] = here
}

// rematchTable is a finished table: its host and the other humans.
type rematchTable struct {
	s     *seatRemoveStack
	meta  GameMeta
	seats map[string]sessionResponse
	bot   uuid.UUID
	pres  *fakePresence
}

// finishedTable seats the named humans (the first hosts) and, when bot is
// set, one bot, starts the game and ends it with `winner` the last seat
// standing. The table settings are not the defaults, so a copy shows.
func finishedTable(t *testing.T, names []string, bot bool, winner string) *rematchTable {
	t.Helper()
	s := newSeatRemoveStack(t)
	pres := &fakePresence{}
	s.l.SetSeatPresence(pres)
	s.l.setRematchTiming(5*time.Millisecond, 20*time.Millisecond)
	meta, err := s.l.CreateBy("Friday night", uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	rt := &rematchTable{s: s, meta: meta, seats: map[string]sessionResponse{}, pres: pres}
	for _, n := range names {
		rt.seats[n] = joinAs(t, s.srv, meta, n)
		if _, err := s.l.SetDeck(meta.ID, rt.seats[n].PlayerID, n+"'s deck", botDeck(30)); err != nil {
			t.Fatal(err)
		}
	}
	if bot {
		_, rt.bot, err = s.l.AddBot(meta.ID, "Bot", "random", "mono-red", "Mono Red", botDeck(20))
		if err != nil {
			t.Fatal(err)
		}
	}
	life := 30
	if _, err := s.l.UpdateSettings(meta.ID, uuid.Nil, game.SettingsPatch{StartingLife: &life}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	room := s.l.RoomOf(meta.ID)
	for _, n := range names {
		if n == winner {
			continue
		}
		id := rt.seats[n].PlayerID
		if _, _, err := room.Apply(id, func() error { return room.Game.Concede(id) }); err != nil {
			t.Fatalf("%s concedes: %v", n, err)
		}
	}
	if bot && winner != "Bot" {
		if _, _, err := room.Apply(rt.bot, func() error { return room.Game.Concede(rt.bot) }); err != nil {
			t.Fatalf("bot concedes: %v", err)
		}
	}
	if room.Game.CurrentState() != game.StateEnded {
		t.Fatalf("setup: state = %s, want ended", room.Game.CurrentState())
	}
	// A commit after the end lets the watcher record it.
	s.l.mu.Lock()
	s.l.syncStateLocked(s.l.games[meta.ID])
	s.l.mu.Unlock()
	return rt
}

func (rt *rematchTable) post(t *testing.T, suffix, token string) *http.Response {
	t.Helper()
	return postJSON(t, rt.s.srv, "/games/"+rt.meta.ID.String()+"/rematch"+suffix, token, map[string]string{})
}

func (rt *rematchTable) offer(t *testing.T, name string) sessionResponse {
	t.Helper()
	resp := rt.post(t, "", rt.seats[name].Token)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("offer: status %d, want 201: %s", resp.StatusCode, body)
	}
	var out sessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out
}

// offerView is the rematch on the old table's view as one seat sees it.
func (rt *rematchTable) offerView(t *testing.T, name string) *protocol.RematchView {
	t.Helper()
	return rt.s.l.RoomOf(rt.meta.ID).ViewFor(rt.seats[name].PlayerID.String()).Rematch
}

func eventually(t *testing.T, what string, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func TestRematchOfferMakesAPrivateTableWithTheSameSetup(t *testing.T) {
	rt := finishedTable(t, []string{"Alice", "Bob"}, true, "Alice")
	rt.pres.set(rt.seats["Bob"].PlayerID, true)

	sess := rt.offer(t, "Alice")
	if sess.Game == nil || sess.Game.ID == rt.meta.ID {
		t.Fatalf("the offer answered with %+v, want the NEW table", sess.Game)
	}
	if sess.Principal.GameID != sess.Game.ID || sess.Principal.PlayerID != sess.PlayerID {
		t.Errorf("the session is bound to %s/%s, want the new table and seat", sess.Principal.GameID, sess.Principal.PlayerID)
	}
	if _, err := rt.s.a.Validate(context.Background(), sess.Token); err != nil {
		t.Errorf("the new session does not validate: %v", err)
	}

	n := mustGet(t, rt.s.l, sess.Game.ID)
	if n.State != string(game.StateLobby) {
		t.Errorf("new table state = %q, want lobby", n.State)
	}
	if n.Name != "Friday night (rematch)" {
		t.Errorf("new table name = %q", n.Name)
	}
	// Decision 4: private. No invite exists, and nothing resolves.
	if n.InviteToken != "" || n.SpectatorInvite != "" {
		t.Errorf("the new table still shows an invite: %q / %q", n.InviteToken, n.SpectatorInvite)
	}
	ms := rt.s.l.store.(*memoryStore)
	ms.mu.Lock()
	live := map[uuid.UUID]int{}
	for _, inv := range ms.invites {
		if inv.Usable(time.Now()) {
			live[inv.GameID]++
		}
	}
	ms.mu.Unlock()
	if live[n.ID] != 0 {
		t.Errorf("the new table has %d usable invites, want none: it is private", live[n.ID])
	}
	if live[rt.meta.ID] != 2 {
		t.Errorf("the old table has %d usable invites, want its own 2", live[rt.meta.ID])
	}
	if got := rt.s.l.RoomOf(n.ID).Game.TableSettingsSnapshot().StartingLife; got != 30 {
		t.Errorf("starting life = %d, want the old table's 30", got)
	}
	// The host and the bot, each with the deck they had.
	if len(n.Players) != 2 {
		t.Fatalf("seats = %+v, want the host and the bot", n.Players)
	}
	host, bot := n.Players[0], n.Players[1]
	if host.Name != "Alice" || host.DeckName != "Alice's deck" || !host.DeckUploaded || host.PlayerID != sess.PlayerID {
		t.Errorf("host seat = %+v", host)
	}
	if !bot.IsBot || bot.Name != "Bot" || bot.BotTier != "random" || bot.BotDeck != "mono-red" || bot.DeckName != "Mono Red" || !bot.DeckUploaded {
		t.Errorf("bot seat = %+v", bot)
	}
	if n.HostPlayerID != sess.PlayerID {
		t.Errorf("host of the new table = %s, want the offering host's new seat", n.HostPlayerID)
	}
	if cards := rt.s.l.RoomOf(n.ID).Game.PlayerByID(sess.PlayerID); cards == nil || len(cards.Library.Cards) == 0 || len(cards.Command.Cards) != 1 {
		t.Errorf("the host's deck did not come across: %+v", cards)
	}

	// The old table is untouched: still ended, with its own id and seats.
	old := mustGet(t, rt.s.l, rt.meta.ID)
	if old.State != string(game.StateEnded) || len(old.Players) != 3 {
		t.Errorf("old table = %s with %d seats, want it ended and whole", old.State, len(old.Players))
	}
	v := rt.offerView(t, "Bob")
	if v == nil || v.GameID != n.ID.String() || v.You != protocol.RematchUndecided || !v.Open {
		t.Errorf("Bob's view of the offer = %+v, want an open offer he has not answered", v)
	}
	if v := rt.offerView(t, "Alice"); v == nil || v.You != protocol.RematchHost {
		t.Errorf("Alice's view of the offer = %+v, want host", v)
	}
	if v := rt.offerView(t, "Bob"); v.Accepted != nil || v.Left != nil {
		t.Errorf("a seat's view carries other seats' answers: %+v", v)
	}
}

func TestRematchIsHostOnlyAndOnlyAfterTheEnd(t *testing.T) {
	rt := finishedTable(t, []string{"Alice", "Bob", "Carol"}, false, "Alice")
	status := func(token string) int {
		resp := rt.post(t, "", token)
		defer resp.Body.Close()
		return resp.StatusCode
	}
	if got := status(rt.seats["Bob"].Token); got != http.StatusForbidden {
		t.Errorf("a non-host offered a rematch: %d, want 403", got)
	}
	if got := status(adminToken(t, rt.s.srv)); got != http.StatusForbidden {
		t.Errorf("the admin token offered a rematch: %d, want 403 (host only)", got)
	}
	if got := status(""); got != http.StatusUnauthorized {
		t.Errorf("no session: %d, want 401", got)
	}
	if got := status(rt.seats["Alice"].Token); got != http.StatusCreated {
		t.Fatalf("the host's offer: %d, want 201", got)
	}
	if got := status(rt.seats["Alice"].Token); got != http.StatusConflict {
		t.Errorf("a second offer: %d, want 409", got)
	}
	// A table still being played cannot be rematched.
	s := newSeatRemoveStack(t)
	meta, seats := s.startedTable(t, uuid.New(), "Zed", "Yan")
	resp := postJSON(t, s.srv, "/games/"+meta.ID.String()+"/rematch", seats[0].Token, map[string]string{})
	resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("a rematch of a live game: %d, want 409", resp.StatusCode)
	}
}

func TestRematchHostFallsBackWhenEveryHumanLost(t *testing.T) {
	// A bot wins, so the engine has nobody left to host: the first human
	// seat offers, because a person against bots would otherwise never
	// get a rematch.
	rt := finishedTable(t, []string{"Alice"}, true, "Bot")
	if rt.s.l.RoomOf(rt.meta.ID).HostPlayerID() != uuid.Nil {
		t.Fatal("setup: the engine still names a host; the fallback is not in play")
	}
	sess := rt.offer(t, "Alice")
	if n := mustGet(t, rt.s.l, sess.Game.ID); len(n.Players) != 2 || n.HostPlayerID != sess.PlayerID {
		t.Errorf("new table = %+v", n)
	}
}

func TestRematchAcceptAndLeave(t *testing.T) {
	rt := finishedTable(t, []string{"Alice", "Bob", "Carol"}, false, "Alice")
	rt.pres.set(rt.seats["Bob"].PlayerID, true)
	rt.pres.set(rt.seats["Carol"].PlayerID, true)
	host := rt.offer(t, "Alice")
	newID := host.Game.ID

	bob := decodeSession(t, rt.post(t, "/accept", rt.seats["Bob"].Token), http.StatusOK)
	if bob.Game == nil || bob.Game.ID != newID {
		t.Fatalf("Bob's accept answered with %+v, want the new table", bob.Game)
	}
	if p, err := rt.s.a.Validate(context.Background(), bob.Token); err != nil || p.GameID != newID || p.PlayerID != bob.PlayerID {
		t.Errorf("Bob's new session = %+v, %v", p, err)
	}
	n := mustGet(t, rt.s.l, newID)
	if len(n.Players) != 2 || n.Players[1].Name != "Bob" || n.Players[1].DeckName != "Bob's deck" || !n.Players[1].DeckUploaded {
		t.Errorf("new table seats = %+v, want Alice and Bob with his deck", n.Players)
	}
	if v := rt.offerView(t, "Carol"); v == nil || !v.Open || v.You != protocol.RematchUndecided {
		t.Errorf("Carol's view = %+v, want the offer still open for her", v)
	}
	// Bob cannot accept twice, nor leave once he is at the new table.
	if resp := rt.post(t, "/accept", rt.seats["Bob"].Token); resp.StatusCode != http.StatusConflict {
		t.Errorf("a second accept: %d, want 409", resp.StatusCode)
	}
	if resp := rt.post(t, "/leave", rt.seats["Bob"].Token); resp.StatusCode != http.StatusConflict {
		t.Errorf("leaving after accepting: %d, want 409", resp.StatusCode)
	}
	// The host is not asked.
	if resp := rt.post(t, "/accept", rt.seats["Alice"].Token); resp.StatusCode != http.StatusForbidden {
		t.Errorf("the host accepting its own offer: %d, want 403", resp.StatusCode)
	}

	// Carol leaves: every other human has answered, so the offer closes.
	if resp := rt.post(t, "/leave", rt.seats["Carol"].Token); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("leave: %d, want 204", resp.StatusCode)
	}
	if v := rt.offerView(t, "Carol"); v == nil || v.Open || v.You != protocol.RematchDeclined {
		t.Errorf("Carol's view after leaving = %+v, want a closed offer she declined", v)
	}
	if v := rt.s.l.RoomOf(rt.meta.ID).ViewFor("").Rematch; v == nil || len(v.Accepted) != 1 || len(v.Left) != 1 {
		t.Errorf("the admin's view = %+v, want one accepted and one left", v)
	}
	if rt.s.l.games[rt.meta.ID].rematchDecks != nil {
		t.Error("the old decks are still kept after the offer closed")
	}
	// A late accept is refused, and the new table is as it was.
	if resp := rt.post(t, "/accept", rt.seats["Carol"].Token); resp.StatusCode != http.StatusConflict {
		t.Errorf("accepting a closed offer: %d, want 409", resp.StatusCode)
	}
	if n := mustGet(t, rt.s.l, newID); len(n.Players) != 2 {
		t.Errorf("seats = %d after a refused accept, want 2", len(n.Players))
	}
}

func TestRematchAcceptIsRefusedOnceTheTableStartedOrIsFull(t *testing.T) {
	rt := finishedTable(t, []string{"Alice", "Bob", "Carol"}, false, "Alice")
	rt.pres.set(rt.seats["Bob"].PlayerID, true)
	rt.pres.set(rt.seats["Carol"].PlayerID, true)
	host := rt.offer(t, "Alice")

	// The host fills the free seats with bots; nobody else fits.
	for _, name := range []string{"B1", "B2", "B3"} {
		if _, _, err := rt.s.l.AddBot(host.Game.ID, name, "random", "", "d", botDeck(5)); err != nil {
			t.Fatal(err)
		}
	}
	if resp := rt.post(t, "/accept", rt.seats["Bob"].Token); resp.StatusCode != http.StatusConflict {
		t.Errorf("accepting a full table: %d, want 409", resp.StatusCode)
	}
	// A refused accept leaves Bob free to answer again.
	if v := rt.offerView(t, "Bob"); v.You != protocol.RematchUndecided {
		t.Errorf("Bob = %q after a refused accept, want undecided", v.You)
	}

	// Started without Carol: the table is no longer joinable.
	if _, err := rt.s.l.RemoveBot(host.Game.ID, mustGet(t, rt.s.l, host.Game.ID).Players[3].PlayerID); err != nil {
		t.Fatal(err)
	}
	started, err := rt.s.l.Start(host.Game.ID)
	if err != nil {
		t.Fatalf("starting the rematch: %v", err)
	}
	if started.State != string(game.StateActive) {
		t.Errorf("rematch state = %s, want active", started.State)
	}
	if resp := rt.post(t, "/accept", rt.seats["Carol"].Token); resp.StatusCode != http.StatusConflict {
		t.Errorf("accepting a started table: %d, want 409", resp.StatusCode)
	}
	// The old game is still there, ended, with its own id.
	if old := mustGet(t, rt.s.l, rt.meta.ID); old.State != string(game.StateEnded) || old.ID == started.ID {
		t.Errorf("old table = %+v", old)
	}
}

func TestRematchLastsUntilEveryOtherHumanHasLeftTheResultsScreen(t *testing.T) {
	rt := finishedTable(t, []string{"Alice", "Bob", "Carol"}, false, "Alice")
	bob, carol := rt.seats["Bob"].PlayerID, rt.seats["Carol"].PlayerID
	rt.pres.set(bob, true)
	rt.pres.set(carol, true)
	rt.offer(t, "Alice")

	// Carol goes to the lobby; Bob is still looking at the board.
	rt.pres.set(carol, false)
	time.Sleep(150 * time.Millisecond) // several grace periods
	if v := rt.offerView(t, "Bob"); !v.Open {
		t.Fatal("the offer closed while Bob was still on the results screen")
	}
	// Bob goes too: the offer closes, and everyone still watching is told.
	rt.pres.set(bob, false)
	eventually(t, "the offer to close", func() bool { return !rt.offerView(t, "Alice").Open })
	if resp := rt.post(t, "/accept", rt.seats["Bob"].Token); resp.StatusCode != http.StatusConflict {
		t.Errorf("accepting after leaving the screen: %d, want 409", resp.StatusCode)
	}
}

func TestRematchOfferWithNoOtherHumanClosesAtOnce(t *testing.T) {
	rt := finishedTable(t, []string{"Alice"}, true, "Alice")
	sess := rt.offer(t, "Alice")
	if v := rt.offerView(t, "Alice"); v == nil || v.Open {
		t.Errorf("view = %+v, want a closed offer: there is nobody to ask", v)
	}
	if n := mustGet(t, rt.s.l, sess.Game.ID); len(n.Players) != 2 {
		t.Errorf("seats = %d, want the host and the bot", len(n.Players))
	}
}

func TestRematchNeedsTheKeptDecks(t *testing.T) {
	rt := finishedTable(t, []string{"Alice", "Bob"}, false, "Alice")
	rt.s.l.mu.Lock()
	rt.s.l.games[rt.meta.ID].rematchDecks = nil
	rt.s.l.mu.Unlock()
	resp := rt.post(t, "", rt.seats["Alice"].Token)
	resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("a rematch with no kept decks: %d, want 409", resp.StatusCode)
	}
	// And nothing was left behind to block a later offer.
	if rt.s.l.games[rt.meta.ID].rematch != nil {
		t.Error("a refused offer left an offer behind")
	}
}

func TestRematchStartedTableKeepsItsDecksForTheNextRematch(t *testing.T) {
	rt := finishedTable(t, []string{"Alice", "Bob"}, true, "Alice")
	first := rt.offer(t, "Alice")
	if _, err := rt.s.l.Start(first.Game.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	// The rematch is itself a table whose decks are kept as it starts.
	rt.s.l.mu.Lock()
	kept := len(rt.s.l.games[first.Game.ID].rematchDecks)
	rt.s.l.mu.Unlock()
	if kept != 2 {
		t.Errorf("the rematch kept %d decks, want 2 (the host and the bot)", kept)
	}
}

func TestRematchTableName(t *testing.T) {
	if got := rematchTableName("Friday night"); got != "Friday night (rematch)" {
		t.Errorf("name = %q", got)
	}
	if got := rematchTableName("Friday night (rematch)"); got != "Friday night (rematch)" {
		t.Errorf("a rematch of a rematch = %q, want one suffix", got)
	}
	long := ""
	for i := 0; i < 100; i++ {
		long += "x"
	}
	if got := rematchTableName(long); len(got) > maxTableNameLen {
		t.Errorf("name is %d bytes, over the %d limit", len(got), maxTableNameLen)
	}
}
