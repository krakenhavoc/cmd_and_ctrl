package lobby

// Tests for pod_stats.go (ADR 0220 slice 5 = ADR 0217 slice 7, #4880):
// the sharing switch, members only, "not sharing" with no numbers, pod
// games, the leaderboard's 5-game minimum, the range, and blocks.

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/gamestats"
)

// podSeat is one seat of a seeded pod game: a person, or a bot when
// who is the zero value.
type podSeat struct {
	who    podPerson
	result gamestats.Result
	place  int
}

// seedPodGame writes an ended game with a version-2 record, daysAgo days
// back; the seats are numbered in order and a zero person is a bot.
func (s podStack) seedPodGame(t *testing.T, daysAgo int, seats ...podSeat) {
	t.Helper()
	id := uuid.New()
	ended := time.Now().Add(-time.Duration(daysAgo) * 24 * time.Hour).UnixMilli()
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := s.db.Exec(q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	exec(`INSERT INTO games (id, name, state, created_at, started_at, ended_at) VALUES (?, 'Friday', 'ended', 1, ?, ?)`,
		id.String(), ended-40*60000, ended)
	pi := func(n int) *int { return &n }
	rec := gamestats.Game{GameID: id, Version: gamestats.Version, Source: gamestats.SourceLive, LogComplete: true,
		Turns: 10, Rounds: 3, StartingSeat: pi(0), ComputedAt: time.UnixMilli(ended), FramesComplete: true}
	for i, st := range seats {
		if st.who.id == uuid.Nil {
			exec(`INSERT INTO seats (game_id, seat, player_id, bot_tier) VALUES (?, ?, ?, 'heuristic')`, id.String(), i, uuid.NewString())
		} else {
			exec(`INSERT INTO seats (game_id, seat, player_id, user_id) VALUES (?, ?, ?, ?)`, id.String(), i, uuid.NewString(), st.who.id.String())
		}
		rec.Seats = append(rec.Seats, gamestats.Seat{Seat: i, Result: st.result, Place: pi(st.place),
			Commanders: []string{}, TurnPosition: pi(i + 1)})
		if st.result == gamestats.ResultWin {
			rec.WinKind = gamestats.KindCombat
		}
	}
	if err := gamestats.NewSQLStore(s.db).Put(context.Background(), rec); err != nil {
		t.Fatal(err)
	}
}

func (s podStack) share(t *testing.T, who podPerson, on bool) {
	t.Helper()
	if code, raw := s.call(t, "PUT", "/me/stats/prefs", who.tok, map[string]bool{"share_pods": on}); code != http.StatusOK {
		t.Fatalf("share_pods: %d %s", code, raw)
	}
}

func (s podStack) stats(t *testing.T, who podPerson, pod, query string) podStatsResponse {
	t.Helper()
	code, raw := s.call(t, "GET", "/pods/"+pod+"/stats"+query, who.tok, nil)
	if code != http.StatusOK {
		t.Fatalf("GET stats: %d %s", code, raw)
	}
	var out podStatsResponse
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func (s podStack) board(t *testing.T, who podPerson, pod, query string) podLeaderboardResponse {
	t.Helper()
	code, raw := s.call(t, "GET", "/pods/"+pod+"/leaderboard"+query, who.tok, nil)
	if code != http.StatusOK {
		t.Fatalf("GET leaderboard: %d %s", code, raw)
	}
	var out podLeaderboardResponse
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

// podOfThree makes a pod of Ann (owner), Bea and Cal.
func (s podStack) podOfThree(t *testing.T) string {
	t.Helper()
	pod := s.makePod(t, s.ann, "Friday night")
	tok := s.link(t, s.ann, pod)
	s.join(t, s.bea, tok)
	s.join(t, s.cal, tok)
	return pod
}

func win(p podPerson) podSeat             { return podSeat{p, gamestats.ResultWin, 1} }
func lose(p podPerson, place int) podSeat { return podSeat{p, gamestats.ResultLoss, place} }

func memberRow(out podStatsResponse, name string) *podStatsMember {
	for i := range out.Members {
		if out.Members[i].DisplayName == name {
			return &out.Members[i]
		}
	}
	return nil
}

func TestSharePodsSwitchLeavesTheAdvancedSwitchAlone(t *testing.T) {
	s := newPodStack(t)
	if code, _ := s.call(t, "PUT", "/me/stats/prefs", s.ann.tok, map[string]string{}); code != http.StatusBadRequest {
		t.Errorf("an empty body: %d, want 400", code)
	}
	s.share(t, s.ann, true)
	code, raw := s.call(t, "GET", "/me/stats/prefs", s.ann.tok, nil)
	var got statsPrefsResponse
	if err := json.Unmarshal([]byte(raw), &got); code != http.StatusOK || err != nil {
		t.Fatalf("GET prefs: %d %s %v", code, raw, err)
	}
	if !got.SharePods || got.Choice != statsChoiceDefault {
		t.Errorf("after sharing: %+v, want share_pods true and advanced still following the default", got)
	}
	s.share(t, s.ann, false)
	_, raw = s.call(t, "GET", "/me/stats/prefs", s.ann.tok, nil)
	got = statsPrefsResponse{}
	_ = json.Unmarshal([]byte(raw), &got)
	if got.SharePods {
		t.Error("share_pods did not switch off")
	}
}

func TestPodStatsAreForMembersOnly(t *testing.T) {
	s := newPodStack(t)
	pod := s.podOfThree(t)
	guest := playerTokenWithUser(t, s.auth, uuid.New(), uuid.New(), uuid.Nil, "Guest")
	for _, path := range []string{"/pods/" + pod + "/stats", "/pods/" + pod + "/leaderboard"} {
		if code, _ := s.call(t, "GET", path, "", nil); code != http.StatusUnauthorized {
			t.Errorf("signed out %s: %d, want 401", path, code)
		}
		if code, _ := s.call(t, "GET", path, guest, nil); code != http.StatusForbidden {
			t.Errorf("guest %s: %d, want 403", path, code)
		}
		if code, _ := s.call(t, "GET", path, s.dee.tok, nil); code != http.StatusNotFound {
			t.Errorf("a non-member %s: %d, want 404", path, code)
		}
		if code, _ := s.call(t, "GET", path+"?range=7d", s.ann.tok, nil); code != http.StatusBadRequest {
			t.Errorf("a bad range %s: %d, want 400", path, code)
		}
		other := "/pods/" + uuid.NewString() + path[len("/pods/"+pod):]
		if code, _ := s.call(t, "GET", other, s.ann.tok, nil); code != http.StatusNotFound {
			t.Errorf("a pod that does not exist %s: %d, want 404", other, code)
		}
	}
}

func TestPodStatsShowOnlyWhoShares(t *testing.T) {
	s := newPodStack(t)
	pod := s.podOfThree(t)
	s.share(t, s.ann, true)
	s.share(t, s.bea, true)
	// Cal is a member who has not switched sharing on.
	for i := 0; i < 5; i++ {
		s.seedPodGame(t, i+1, win(s.ann), lose(s.bea, 2), lose(s.cal, 3), lose(podPerson{}, 4))
	}
	s.seedPodGame(t, 6, win(s.bea), lose(s.ann, 2), lose(s.cal, 3))
	s.seedPodGame(t, 7, win(s.ann), lose(podPerson{}, 2)) // Ann alone: not a pod game
	s.seedPodGame(t, 8, win(s.dee), lose(s.ann, 2))       // Dee is not in the pod

	out := s.stats(t, s.bea, pod, "")
	if out.Games != 6 {
		t.Errorf("pod games %d, want 6 (the games two sharing members sat in)", out.Games)
	}
	ann := memberRow(out, "Ann")
	if ann == nil || !ann.Sharing || ann.PodGame == nil || ann.Overall == nil {
		t.Fatalf("Ann: %+v", ann)
	}
	if ann.PodGame.Games != 6 || ann.PodGame.Wins != 5 {
		t.Errorf("Ann in pod games: %+v, want 6 games, 5 wins", ann.PodGame)
	}
	if ann.Overall.Games != 8 || ann.Overall.Wins != 6 {
		t.Errorf("Ann overall: %+v, want 8 games, 6 wins", ann.Overall)
	}
	cal := memberRow(out, "Cal")
	if cal == nil || cal.Sharing || cal.PodGame != nil || cal.Overall != nil {
		t.Errorf("Cal has not switched sharing on and must show no numbers: %+v", cal)
	}
	if !memberRow(out, "Bea").You {
		t.Error("the viewer's own row is not marked")
	}
	if len(out.WinKinds) != 1 || out.WinKinds[0].Kind != gamestats.KindCombat || out.WinKinds[0].Games != 6 {
		t.Errorf("win kinds %+v, want 6 combat wins", out.WinKinds)
	}
	if len(out.HeadToHead) != 1 {
		t.Fatalf("head to head %+v, want the one pair of sharing members", out.HeadToHead)
	}
	if h := out.HeadToHead[0]; h.Games != 6 || h.AWins+h.BWins != 6 {
		t.Errorf("head to head %+v, want 6 games all won by one of the two", h)
	}

	board := s.board(t, s.cal, pod, "")
	if len(board.Rows) != 2 || board.Rows[0].DisplayName != "Ann" || board.Rows[0].Rank != 1 || board.Rows[1].DisplayName != "Bea" || board.Rows[1].Rank != 2 {
		t.Errorf("leaderboard %+v, want Ann then Bea", board.Rows)
	}
	for _, r := range board.Rows {
		if r.Games == 0 {
			t.Errorf("a row without its count: %+v", r)
		}
	}
	if len(board.Unranked) != 0 {
		t.Errorf("unranked %+v: a member who is not sharing is not on the board at all", board.Unranked)
	}
	// The viewer need not share to look.
	if cal := s.board(t, s.cal, pod, ""); cal.MinGames != 5 {
		t.Errorf("min games %d, want 5", cal.MinGames)
	}
}

func TestPodLeaderboardNeedsFiveGamesAndHonoursTheRange(t *testing.T) {
	s := newPodStack(t)
	pod := s.podOfThree(t)
	s.share(t, s.ann, true)
	s.share(t, s.bea, true)
	s.share(t, s.cal, true)
	for i := 0; i < 5; i++ {
		s.seedPodGame(t, 100+i, win(s.ann), lose(s.bea, 2))
	}
	for i := 0; i < 2; i++ {
		s.seedPodGame(t, i+1, win(s.cal), lose(s.bea, 2))
	}
	all := s.board(t, s.ann, pod, "?range=all")
	if len(all.Rows) != 2 || all.Rows[0].DisplayName != "Ann" || all.Rows[1].DisplayName != "Bea" || all.Rows[1].Games != 7 {
		t.Errorf("all time: %+v, want Ann (5-0) then Bea (7 games)", all.Rows)
	}
	if len(all.Unranked) != 1 || all.Unranked[0].DisplayName != "Cal" || all.Unranked[0].Needs != 3 {
		t.Errorf("all time unranked %+v, want Cal needing 3 more", all.Unranked)
	}
	recent := s.board(t, s.ann, pod, "?range=90d")
	if recent.Games != 2 || len(recent.Rows) != 0 {
		t.Errorf("the last 90 days: %d games and rows %+v, want 2 games and nobody ranked", recent.Games, recent.Rows)
	}
	if st := s.stats(t, s.ann, pod, "?range=90d"); st.Games != 2 || st.Range != gamestats.Range90 {
		t.Errorf("stats over 90 days: %d games, range %q", st.Games, st.Range)
	}
}

func TestPodStatsHideABlockedMemberButKeepThePodsNumbers(t *testing.T) {
	s := newPodStack(t)
	pod := s.podOfThree(t)
	s.share(t, s.ann, true)
	s.share(t, s.bea, true)
	s.share(t, s.cal, true)
	for i := 0; i < 6; i++ {
		s.seedPodGame(t, i+1, win(s.ann), lose(s.bea, 2), lose(s.cal, 3))
	}
	before := s.board(t, s.cal, pod, "")
	s.block(t, s.ann, s.bea) // after both joined: neither is removed

	annView := s.stats(t, s.ann, pod, "")
	if memberRow(annView, "Bea") != nil {
		t.Error("Ann blocked Bea, and Bea is still in Ann's member list")
	}
	if annView.Games != 6 {
		t.Errorf("the pod's games %d, want 6: a block hides a person, not the group's history", annView.Games)
	}
	if a := memberRow(annView, "Ann"); a.PodGame.Games != 6 || a.PodGame.Wins != 6 {
		t.Errorf("Ann's own record changed: %+v", a.PodGame)
	}
	for _, h := range annView.HeadToHead {
		if h.A.DisplayName == "Bea" || h.B.DisplayName == "Bea" {
			t.Errorf("head to head still names Bea to Ann: %+v", h)
		}
	}
	annBoard := s.board(t, s.ann, pod, "")
	if len(annBoard.Rows) != 2 || annBoard.Rows[0].DisplayName != "Ann" || annBoard.Rows[1].DisplayName != "Cal" || annBoard.Rows[1].Rank != 2 {
		t.Errorf("Ann's board %+v, want Ann then Cal with ranks closed up", annBoard.Rows)
	}
	// Both directions: Bea no longer sees Ann.
	beaBoard := s.board(t, s.bea, pod, "")
	for _, r := range beaBoard.Rows {
		if r.DisplayName == "Ann" {
			t.Error("Bea still sees Ann, who blocked them")
		}
	}
	// Cal blocked nobody: everyone, with the numbers they had.
	after := s.board(t, s.cal, pod, "")
	if len(after.Rows) != 3 || len(before.Rows) != 3 || after.Rows[1].Games != before.Rows[1].Games {
		t.Errorf("Cal's board changed: before %+v after %+v", before.Rows, after.Rows)
	}
}

func TestPodStatsNeverCountABotOrAnOutsiderSeat(t *testing.T) {
	s := newPodStack(t)
	pod := s.podOfThree(t)
	s.share(t, s.ann, true)
	s.share(t, s.bea, true)
	// A bot is the only other seat at the table: Ann and Bea never met.
	s.seedPodGame(t, 1, win(s.ann), lose(podPerson{}, 2))
	s.seedPodGame(t, 2, win(s.bea), lose(podPerson{}, 2))
	if out := s.stats(t, s.ann, pod, ""); out.Games != 0 || len(out.WinKinds) != 0 {
		t.Errorf("games with a bot and one member are not pod games: %+v", out)
	}
}
