package pods

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/social"
)

// Tests for ADR 0220 §4 over hand-built games: the signal, each
// exclusion, dismissal and making a pod from a suggestion.

var gameSeq int

// game builds one game with the given seats. A seat is a user id, or
// "guest", "bot" or "agent:<user id>" for the seats that are not people.
func (f *fixture) game(t *testing.T, state string, ended time.Time, seats ...string) {
	t.Helper()
	gameSeq++
	id := fmt.Sprintf("game-%d", gameSeq)
	var end any
	if !ended.IsZero() {
		end = ended.UnixMilli()
	}
	if _, err := f.d.Exec(`INSERT INTO games (id, name, state, created_at, ended_at) VALUES (?, 'T', ?, 1, ?)`, id, state, end); err != nil {
		t.Fatal(err)
	}
	for i, s := range seats {
		var user, bot, agent any
		switch {
		case s == "guest":
		case s == "bot":
			bot = "easy"
		case strings.HasPrefix(s, "agent:"):
			user, agent = strings.TrimPrefix(s, "agent:"), "claude"
		default:
			user = s
		}
		if _, err := f.d.Exec(`INSERT INTO seats (game_id, seat, player_id, user_id, bot_tier, agent_client) VALUES (?, ?, ?, ?, ?, ?)`,
			id, i, fmt.Sprintf("p%d", i), user, bot, agent); err != nil {
			t.Fatal(err)
		}
	}
}

// games builds n ended games a day apart, ending within the window.
func (f *fixture) games(t *testing.T, n int, seats ...string) {
	t.Helper()
	for i := 0; i < n; i++ {
		f.game(t, "ended", f.now.Add(-time.Duration(i+1)*24*time.Hour), seats...)
	}
}

func (f *fixture) suggested(t *testing.T, user string) []Suggestion {
	t.Helper()
	got, err := f.s.Suggestions(ctx, user)
	if err != nil {
		t.Fatalf("Suggestions(%s): %v", user, err)
	}
	return got
}

func names(s Suggestion) string {
	var n []string
	for _, p := range s.Others {
		n = append(n, p.DisplayName)
	}
	return strings.Join(n, ",")
}

func TestSuggestionSignal(t *testing.T) {
	f := newFixture(t, "ann", "bea", "cal", "dee", "eve")

	f.games(t, 2, "ann", "bea", "cal")
	if got := f.suggested(t, "ann"); len(got) != 0 {
		t.Fatalf("2 games is not the signal: %+v", got)
	}
	f.games(t, 1, "ann", "bea", "cal")
	for _, who := range []string{"ann", "bea", "cal"} {
		got := f.suggested(t, who)
		if len(got) != 1 || got[0].Games != 3 {
			t.Fatalf("%s: want one suggestion of 3 games, got %+v", who, got)
		}
	}
	if got := f.suggested(t, "ann"); names(got[0]) != "Bea,Cal" {
		t.Errorf("others: %s", names(got[0]))
	}
	if got := f.suggested(t, "dee"); len(got) != 0 {
		t.Errorf("a stranger to the set sees %+v", got)
	}
	if a, b := f.suggested(t, "ann")[0].Key, f.suggested(t, "bea")[0].Key; a != b || a != GroupKey([]string{"cal", "bea", "ann"}) {
		t.Errorf("one key for the set whoever reads it: %s %s", a, b)
	}
}

func TestSuggestionSetsAreExact(t *testing.T) {
	f := newFixture(t, "ann", "bea", "cal", "dee", "eve")
	// Two games of {ann,bea,cal} and two of {ann,bea,cal,dee}: two different
	// sets, neither at 3.
	f.games(t, 2, "ann", "bea", "cal")
	f.games(t, 2, "ann", "bea", "cal", "dee")
	if got := f.suggested(t, "ann"); len(got) != 0 {
		t.Fatalf("the sets are not pooled: %+v", got)
	}
	f.games(t, 1, "ann", "bea", "cal", "dee")
	got := f.suggested(t, "ann")
	if len(got) != 1 || got[0].Games != 3 || names(got[0]) != "Bea,Cal,Dee" {
		t.Fatalf("four people at 3 games: %+v", got)
	}
	// Five people is a table of another size: never a suggestion.
	f.games(t, 3, "ann", "bea", "cal", "dee", "eve")
	if got := f.suggested(t, "ann"); len(got) != 1 {
		t.Errorf("five-person sets are not suggested: %+v", got)
	}
}

func TestSuggestionNoPairs(t *testing.T) {
	f := newFixture(t, "ann", "bea")
	f.games(t, 5, "ann", "bea")
	f.games(t, 5, "ann", "bea", "guest", "bot")
	if got := f.suggested(t, "ann"); len(got) != 0 {
		t.Fatalf("a pair is what friends are for: %+v", got)
	}
}

func TestSuggestionGuestsBotsAndAgentsAreNeverInTheSet(t *testing.T) {
	f := newFixture(t, "ann", "bea", "cal", "dee")
	f.games(t, 1, "ann", "bea", "cal", "guest")
	f.games(t, 1, "ann", "bea", "cal", "bot")
	f.games(t, 1, "ann", "bea", "cal", "agent:dee")
	got := f.suggested(t, "ann")
	if len(got) != 1 || got[0].Games != 3 || names(got[0]) != "Bea,Cal" {
		t.Fatalf("guests, bots and agents do not stop a game counting or join the set: %+v", got)
	}
	if got := f.suggested(t, "dee"); len(got) != 0 {
		t.Errorf("an agent's seat is not a person: %+v", got)
	}
}

func TestSuggestionCountsOnlyEndedGamesInTheWindow(t *testing.T) {
	f := newFixture(t, "ann", "bea", "cal")
	f.game(t, "ended", f.now.Add(-time.Hour), "ann", "bea", "cal")
	f.game(t, "ended", f.now.Add(-29*24*time.Hour), "ann", "bea", "cal")
	f.game(t, "ended", f.now.Add(-31*24*time.Hour), "ann", "bea", "cal") // too old
	f.game(t, "active", time.Time{}, "ann", "bea", "cal")                // not ended
	f.game(t, "lobby", time.Time{}, "ann", "bea", "cal")
	if got := f.suggested(t, "ann"); len(got) != 0 {
		t.Fatalf("2 ended games in the window: %+v", got)
	}
	f.game(t, "ended", f.now.Add(-48*time.Hour), "ann", "bea", "cal")
	if got := f.suggested(t, "ann"); len(got) != 1 || got[0].Games != 3 {
		t.Fatalf("3 ended games in the window: %+v", got)
	}
}

func TestSuggestionOrderIsMostGamesThenMostRecent(t *testing.T) {
	f := newFixture(t, "ann", "bea", "cal", "dee", "eve")
	f.games(t, 3, "ann", "bea", "cal")
	f.games(t, 5, "ann", "dee", "eve")
	got := f.suggested(t, "ann")
	if len(got) != 2 || names(got[0]) != "Dee,Eve" || got[0].Games != 5 || names(got[1]) != "Bea,Cal" {
		t.Fatalf("order: %+v", got)
	}
}

func TestSuggestionExclusions(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, f *fixture)
	}{
		{"a block between two of them, either way", func(t *testing.T, f *fixture) { f.block(t, "cal", "bea") }},
		{"a block on the caller", func(t *testing.T, f *fixture) { f.block(t, "bea", "ann") }},
		{"someone takes requests from nobody", func(t *testing.T, f *fixture) {
			n := social.RequestsNobody
			if _, err := f.soc.SetPrefs(ctx, "cal", nil, &n); err != nil {
				t.Fatal(err)
			}
		}},
		{"the caller takes requests from nobody", func(t *testing.T, f *fixture) {
			n := social.RequestsNobody
			if _, err := f.soc.SetPrefs(ctx, "ann", nil, &n); err != nil {
				t.Fatal(err)
			}
		}},
		{"a deleted account", func(t *testing.T, f *fixture) {
			if _, err := f.d.Exec(`UPDATE users SET deleted_at = 5 WHERE id = 'bea'`); err != nil {
				t.Fatal(err)
			}
		}},
		{"a pod that holds all of them", func(t *testing.T, f *fixture) {
			p := f.create(t, "bea", "Home")
			f.addMember(t, p, "ann")
			f.addMember(t, p, "cal")
		}},
		{"a pod that holds all of them and more", func(t *testing.T, f *fixture) {
			p := f.create(t, "bea", "Home")
			f.addMember(t, p, "ann")
			f.addMember(t, p, "cal")
			f.addMember(t, p, "dee")
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t, "ann", "bea", "cal", "dee")
			f.games(t, 4, "ann", "bea", "cal")
			if got := f.suggested(t, "ann"); len(got) != 1 {
				t.Fatalf("before: %+v", got)
			}
			c.setup(t, f)
			for _, who := range []string{"ann", "bea", "cal"} {
				if got := f.suggested(t, who); len(got) != 0 {
					t.Errorf("%s still sees %+v", who, got)
				}
			}
		})
	}
}

func TestSuggestionAPodHoldingOnlySomeOfThemIsNotEnough(t *testing.T) {
	f := newFixture(t, "ann", "bea", "cal")
	f.games(t, 3, "ann", "bea", "cal")
	p := f.create(t, "ann", "Two")
	f.addMember(t, p, "bea")
	if got := f.suggested(t, "ann"); len(got) != 1 {
		t.Fatalf("two of three in a pod: %+v", got)
	}
}

func TestSuggestionDismissalHidesFromTheCallerOnlyFor180Days(t *testing.T) {
	f := newFixture(t, "ann", "bea", "cal")
	f.games(t, 3, "ann", "bea", "cal")
	key := f.suggested(t, "ann")[0].Key
	if err := f.s.Dismiss(ctx, "ann", key); err != nil {
		t.Fatal(err)
	}
	if got := f.suggested(t, "ann"); len(got) != 0 {
		t.Errorf("dismissed: %+v", got)
	}
	if got := f.suggested(t, "bea"); len(got) != 1 {
		t.Errorf("it hides it from nobody else: %+v", got)
	}
	if err := f.s.Dismiss(ctx, "ann", key); err != nil {
		t.Errorf("dismissing twice: %v", err)
	}
	// 179 days on the games are gone from the window, so keep them fresh.
	f.now = f.now.Add(DismissTTL - time.Hour)
	f.games(t, 3, "ann", "bea", "cal")
	if got := f.suggested(t, "ann"); len(got) != 0 {
		t.Errorf("still hidden just inside 180 days: %+v", got)
	}
	f.now = f.now.Add(2 * time.Hour)
	f.games(t, 3, "ann", "bea", "cal")
	if got := f.suggested(t, "ann"); len(got) != 1 {
		t.Errorf("back after 180 days while the group still meets: %+v", got)
	}
}

func TestDismissAKeyThatIsNotYoursIs404(t *testing.T) {
	f := newFixture(t, "ann", "bea", "cal", "dee")
	f.games(t, 3, "ann", "bea", "cal")
	key := f.suggested(t, "ann")[0].Key
	for _, c := range []struct{ who, key string }{{"dee", key}, {"ann", "nope"}, {"ann", GroupKey([]string{"ann", "bea", "dee"})}} {
		if err := f.s.Dismiss(ctx, c.who, c.key); !errors.Is(err, ErrNoSuggestion) {
			t.Errorf("%s %s: %v", c.who, c.key, err)
		}
	}
}

func TestMakeAPodFromASuggestion(t *testing.T) {
	f := newFixture(t, "ann", "bea", "cal", "dee")
	f.games(t, 3, "ann", "bea", "cal")
	key := f.suggested(t, "ann")[0].Key
	if _, err := f.s.CreateFromSuggestion(ctx, "dee", "Nope", key); !errors.Is(err, ErrNoSuggestion) {
		t.Fatalf("someone outside the set: %v", err)
	}
	if _, err := f.s.CreateFromSuggestion(ctx, "ann", "  ", key); !errors.Is(err, ErrNameInvalid) {
		t.Fatalf("name: %v", err)
	}
	s, err := f.s.CreateFromSuggestion(ctx, "ann", "Thursday", key)
	if err != nil {
		t.Fatal(err)
	}
	if s.Role != RoleOwner || s.Members != 1 {
		t.Errorf("summary: %+v", s)
	}
	// They are not friends, yet each gets an ordinary invite; nobody else does.
	for _, who := range []string{"bea", "cal"} {
		o, err := f.s.Mine(ctx, who)
		if err != nil || len(o.Invites) != 1 || o.Invites[0].PodName != "Thursday" || o.Invites[0].From.DisplayName != "Ann" {
			t.Errorf("%s invites: %+v %v", who, o, err)
		}
	}
	if o, _ := f.s.Mine(ctx, "dee"); len(o.Invites) != 0 {
		t.Errorf("dee: %+v", o.Invites)
	}
	if _, err := f.s.Accept(ctx, "bea", s.ID); err != nil {
		t.Errorf("accepting: %v", err)
	}
	if _, err := f.s.Accept(ctx, "cal", s.ID); err != nil {
		t.Errorf("accepting: %v", err)
	}
	if got := f.suggested(t, "ann"); len(got) != 0 {
		t.Errorf("a pod holds all of them now: %+v", got)
	}
}

func TestMakeAPodFromASuggestionHonoursTheLimitsAndCanBeMadeAfterNotNow(t *testing.T) {
	f := newFixture(t, "ann", "bea", "cal")
	f.games(t, 3, "ann", "bea", "cal")
	key := f.suggested(t, "ann")[0].Key
	if err := f.s.Dismiss(ctx, "ann", key); err != nil {
		t.Fatal(err)
	}
	if _, err := f.s.CreateFromSuggestion(ctx, "ann", "Late", key); err != nil {
		t.Errorf("a dismissed suggestion can still be made: %v", err)
	}
	// A block after the suggestion was shown: the key no longer holds.
	f2 := newFixture(t, "ann", "bea", "cal")
	f2.games(t, 3, "ann", "bea", "cal")
	k2 := f2.suggested(t, "ann")[0].Key
	f2.block(t, "bea", "cal")
	if _, err := f2.s.CreateFromSuggestion(ctx, "ann", "Blocked", k2); !errors.Is(err, ErrNoSuggestion) {
		t.Errorf("a block made since: %v", err)
	}
	if o, _ := f2.s.Mine(ctx, "ann"); len(o.Pods) != 0 {
		t.Errorf("no pod made: %+v", o.Pods)
	}
}
