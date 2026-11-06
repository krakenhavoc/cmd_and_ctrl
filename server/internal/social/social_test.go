package social

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

// Tests for ADR 0213 §2 and §3 over migration 0032: requests and their
// answers, two people asking each other, the limits, expiry, blocks and
// what they end, and the settings.

var t0 = time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)

type fixture struct {
	s   *Store
	d   *db.DB
	now time.Time
}

func newFixture(t *testing.T, people ...string) *fixture {
	t.Helper()
	d, err := db.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	f := &fixture{s: NewStore(d), d: d, now: t0}
	f.s.SetClock(func() time.Time { return f.now })
	for _, p := range people {
		f.user(t, p)
	}
	return f
}

func (f *fixture) user(t *testing.T, id string) {
	t.Helper()
	if _, err := f.d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, ?, 1, 1)`,
		id, "Name "+id); err != nil {
		t.Fatal(err)
	}
}

func (f *fixture) request(t *testing.T, from, to string) Outcome {
	t.Helper()
	out, err := f.s.Request(context.Background(), from, to)
	if err != nil {
		t.Fatalf("Request %s -> %s: %v", from, to, err)
	}
	return out
}

func (f *fixture) friends(t *testing.T, a, b string) bool {
	t.Helper()
	ok, err := f.s.AreFriends(context.Background(), a, b)
	if err != nil {
		t.Fatal(err)
	}
	return ok
}

func (f *fixture) overview(t *testing.T, user string) Overview {
	t.Helper()
	o, err := f.s.Overview(context.Background(), user)
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func ids(ps []Person) []string {
	out := []string{}
	for _, p := range ps {
		out = append(out, p.UserID)
	}
	return out
}

func same(a []string, b ...string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestARequestAcceptedMakesFriendsBothWays(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t, "a", "b")
	if got := f.request(t, "a", "b"); got != OutcomeSent {
		t.Fatalf("first request: %q, want sent", got)
	}
	if got := f.request(t, "a", "b"); got != OutcomePending {
		t.Errorf("the same request again: %q, want pending", got)
	}
	if o := f.overview(t, "a"); !same(ids(o.Outgoing), "b") || len(o.Incoming) != 0 {
		t.Errorf("sender's overview: %+v", o)
	}
	if o := f.overview(t, "b"); !same(ids(o.Incoming), "a") || len(o.Outgoing) != 0 {
		t.Errorf("receiver's overview: %+v", o)
	}
	if err := f.s.Accept(ctx, "b", "a"); err != nil {
		t.Fatalf("Accept: %v", err)
	}
	if !f.friends(t, "a", "b") || !f.friends(t, "b", "a") {
		t.Fatal("not friends after an accept")
	}
	for _, who := range []string{"a", "b"} {
		o := f.overview(t, who)
		if len(o.Friends) != 1 || len(o.Incoming)+len(o.Outgoing) != 0 {
			t.Errorf("%s after the accept: %+v; want one friend and no requests", who, o)
		}
	}
	if got := f.request(t, "b", "a"); got != OutcomeAlreadyFriends {
		t.Errorf("a request between friends: %q", got)
	}
	if err := f.s.Accept(ctx, "b", "a"); !errors.Is(err, ErrNoRequest) {
		t.Errorf("accepting again: %v, want ErrNoRequest", err)
	}
}

func TestTwoPeopleWhoAskEachOtherAreFriendsAtOnce(t *testing.T) {
	f := newFixture(t, "a", "b")
	f.request(t, "a", "b")
	if got := f.request(t, "b", "a"); got != OutcomeFriends {
		t.Fatalf("the answering request: %q, want friends", got)
	}
	if !f.friends(t, "a", "b") {
		t.Fatal("not friends")
	}
	if o := f.overview(t, "a"); len(o.Outgoing)+len(o.Incoming) != 0 {
		t.Errorf("requests left behind: %+v", o)
	}
}

func TestIgnoreIsSilentAndCancelWithdraws(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t, "a", "b", "c")
	f.request(t, "a", "b")
	if err := f.s.Ignore(ctx, "b", "a"); err != nil {
		t.Fatalf("Ignore: %v", err)
	}
	if o := f.overview(t, "b"); len(o.Incoming) != 0 {
		t.Errorf("an ignored request is still on the receiver's list: %+v", o.Incoming)
	}
	if o := f.overview(t, "a"); !same(ids(o.Outgoing), "b") {
		t.Errorf("the sender is told: outgoing %v, want b still pending", ids(o.Outgoing))
	}
	if r, _ := f.s.Relations(ctx, "b"); r.Incoming["a"] {
		t.Error("Relations marks an ignored request as incoming")
	}
	// Asking back still makes friends: both people asked.
	if got := f.request(t, "b", "a"); got != OutcomeFriends {
		t.Errorf("asking back after ignoring: %q, want friends", got)
	}

	f.request(t, "a", "c")
	if err := f.s.Cancel(ctx, "a", "c"); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if o := f.overview(t, "c"); len(o.Incoming) != 0 {
		t.Errorf("a cancelled request reached its receiver: %+v", o.Incoming)
	}
	if err := f.s.Cancel(ctx, "a", "c"); !errors.Is(err, ErrNoRequest) {
		t.Errorf("cancelling twice: %v", err)
	}
	if err := f.s.Ignore(ctx, "c", "a"); !errors.Is(err, ErrNoRequest) {
		t.Errorf("ignoring nothing: %v", err)
	}
}

func TestUnfriendIsSilentFromEitherSide(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t, "a", "b")
	f.request(t, "a", "b")
	f.request(t, "b", "a")
	if err := f.s.Unfriend(ctx, "b", "a"); err != nil {
		t.Fatal(err)
	}
	if f.friends(t, "a", "b") {
		t.Fatal("still friends")
	}
	if err := f.s.Unfriend(ctx, "a", "b"); err != nil {
		t.Errorf("unfriending again: %v", err)
	}
}

func TestRequestsExpireAfterThirtyDays(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t, "a", "b")
	f.request(t, "a", "b")
	f.now = t0.Add(RequestTTL)
	if o := f.overview(t, "b"); len(o.Incoming) != 0 {
		t.Errorf("an expired request is still incoming: %+v", o.Incoming)
	}
	if err := f.s.Accept(ctx, "b", "a"); !errors.Is(err, ErrNoRequest) {
		t.Errorf("accepting an expired request: %v", err)
	}
	var n int
	if err := f.d.QueryRow(`SELECT COUNT(*) FROM friend_requests`).Scan(&n); err != nil || n != 0 {
		t.Errorf("the expired row is still stored: %d (%v)", n, err)
	}
	// Asking again after it lapsed is a new request, and the other side's
	// old one does not make them friends.
	if got := f.request(t, "b", "a"); got != OutcomeSent {
		t.Errorf("a request after the other's lapsed: %q, want sent", got)
	}
}

func TestBlockEndsEverythingAndRefusesBothWays(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t, "a", "b", "c")
	f.request(t, "a", "b")
	f.request(t, "b", "a") // friends
	f.request(t, "c", "a")
	f.request(t, "a", "c")
	f.request(t, "b", "c")

	if err := f.s.Block(ctx, "a", "b"); err != nil {
		t.Fatalf("Block: %v", err)
	}
	if err := f.s.Block(ctx, "a", "b"); err != nil {
		t.Errorf("blocking twice: %v", err)
	}
	if f.friends(t, "a", "b") {
		t.Error("the block left the friendship")
	}
	if _, err := f.s.Request(ctx, "a", "b"); !errors.Is(err, ErrYouBlocked) {
		t.Errorf("the blocker asking: %v, want ErrYouBlocked", err)
	}
	if _, err := f.s.Request(ctx, "b", "a"); !errors.Is(err, ErrRefused) {
		t.Errorf("the blocked person asking: %v, want ErrRefused", err)
	}
	for _, pair := range [][2]string{{"a", "b"}, {"b", "a"}} {
		if ok, _ := f.s.Blocked(ctx, pair[0], pair[1]); !ok {
			t.Errorf("Blocked(%s, %s) = false", pair[0], pair[1])
		}
	}
	if mine, _ := f.s.IsBlocking(ctx, "b", "a"); mine {
		t.Error("IsBlocking is one direction: b did not block a")
	}
	with, _ := f.s.BlockedWith(ctx, "b")
	if !with["a"] || with["c"] {
		t.Errorf("BlockedWith(b) = %v", with)
	}
	if hit, _ := f.s.BlockedWithAny(ctx, "b", []string{"c", "a"}); !hit {
		t.Error("BlockedWithAny(b, [c a]) = false")
	}
	if hit, _ := f.s.BlockedWithAny(ctx, "c", []string{"a", "b"}); hit {
		t.Error("BlockedWithAny(c, [a b]) = true; c has no block")
	}
	// Requests with someone else are untouched (a and c are friends now:
	// they asked each other).
	if !f.friends(t, "a", "c") {
		t.Error("a's friendship with c went with the block on b")
	}
	if o := f.overview(t, "c"); !same(ids(o.Incoming), "b") {
		t.Errorf("b's request to c went with the block: %v", ids(o.Incoming))
	}
	bl, _ := f.s.Blocks(ctx, "a")
	if !same(ids(bl), "b") {
		t.Errorf("a's blocks: %v", ids(bl))
	}

	// Unblocking brings nothing back.
	if err := f.s.Unblock(ctx, "a", "b"); err != nil {
		t.Fatal(err)
	}
	if f.friends(t, "a", "b") {
		t.Error("unblocking restored the friendship")
	}
	if got := f.request(t, "b", "a"); got != OutcomeSent {
		t.Errorf("a request after the unblock: %q", got)
	}
}

func TestNobodyTakesNoRequests(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t, "a", "b")
	p, err := f.s.Prefs(ctx, "b")
	if err != nil || p.Presence != PresenceFriends || p.Requests != RequestsPlayedWith {
		t.Fatalf("defaults: %+v %v", p, err)
	}
	nobody := RequestsNobody
	if p, err = f.s.SetPrefs(ctx, "b", nil, &nobody); err != nil || p.Requests != RequestsNobody || p.Presence != PresenceFriends {
		t.Fatalf("SetPrefs: %+v %v", p, err)
	}
	if _, err := f.s.Request(ctx, "a", "b"); !errors.Is(err, ErrRefused) {
		t.Errorf("a request to someone taking none: %v, want ErrRefused", err)
	}
	// Their own request still goes out, and asking back still answers it.
	f.request(t, "b", "a")
	if got := f.request(t, "a", "b"); got != OutcomeFriends {
		t.Errorf("answering their request: %q", got)
	}
	bad := Presence("everyone")
	if _, err := f.s.SetPrefs(ctx, "b", &bad, nil); err == nil {
		t.Error("SetPrefs took a presence that is not a setting")
	}
	hidden := PresenceNobody
	if p, _ = f.s.SetPrefs(ctx, "b", &hidden, nil); p.Requests != RequestsNobody || p.Presence != PresenceNobody {
		t.Errorf("a second SetPrefs lost the first: %+v", p)
	}
}

func TestTheLimits(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t, "me")
	for i := 0; i < MaxOutgoing+2; i++ {
		f.user(t, fmt.Sprintf("p%02d", i))
	}
	// Ten new requests an hour.
	for i := 0; i < RequestsPerHour; i++ {
		f.request(t, "me", fmt.Sprintf("p%02d", i))
	}
	if _, err := f.s.Request(ctx, "me", "p10"); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("request %d this hour: %v, want ErrRateLimited", RequestsPerHour+1, err)
	}
	if got := f.request(t, "me", "p00"); got != OutcomePending {
		t.Errorf("an already pending request is not a new one: %q", got)
	}
	f.now = f.now.Add(time.Hour)
	for i := RequestsPerHour; i < MaxOutgoing; i++ {
		f.request(t, "me", fmt.Sprintf("p%02d", i))
	}
	f.now = f.now.Add(time.Hour)
	if _, err := f.s.Request(ctx, "me", fmt.Sprintf("p%02d", MaxOutgoing)); !errors.Is(err, ErrTooManyPending) {
		t.Fatalf("request %d pending: %v, want ErrTooManyPending", MaxOutgoing+1, err)
	}
	f.s.Forget("me")
	if len(f.s.sent) != 0 {
		t.Errorf("Forget left the rate window: %v", f.s.sent)
	}

	// Two hundred friends.
	g := newFixture(t, "me", "them", "late")
	for i := 0; i < MaxFriends; i++ {
		id := fmt.Sprintf("f%03d", i)
		g.user(t, id)
		a, b := pair("me", id)
		if _, err := g.d.Exec(`INSERT INTO friendships (user_a, user_b, created_at) VALUES (?, ?, 1)`, a, b); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := g.s.Request(ctx, "me", "them"); !errors.Is(err, ErrFriendCap) {
		t.Errorf("a request with a full list: %v, want ErrFriendCap", err)
	}
	g.request(t, "late", "me")
	if err := g.s.Accept(ctx, "me", "late"); !errors.Is(err, ErrFriendCap) {
		t.Errorf("accepting with a full list: %v, want ErrFriendCap", err)
	}
	if err := g.s.Accept(ctx, "me", "late"); !errors.Is(err, ErrFriendCap) {
		t.Errorf("the request went with the refused accept: %v", err)
	}
	g.request(t, "them", "late")
	g.request(t, "late", "them")
	if !g.friends(t, "them", "late") {
		t.Error("two others could not be friends")
	}
	// "them" asks "me", whose list is full: the request goes, the
	// answering accept is refused as theirs.
	g.request(t, "them", "me")
	g2, err := g.s.Request(ctx, "them", "me")
	if err != nil || g2 != OutcomePending {
		t.Errorf("again: %q %v", g2, err)
	}
}

func TestTheirFullListRefusesAMutualRequest(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t, "me", "full")
	for i := 0; i < MaxFriends; i++ {
		id := fmt.Sprintf("f%03d", i)
		f.user(t, id)
		a, b := pair("full", id)
		if _, err := f.d.Exec(`INSERT INTO friendships (user_a, user_b, created_at) VALUES (?, ?, 1)`, a, b); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := f.d.Exec(`INSERT INTO friend_requests (from_user, to_user, created_at) VALUES ('full', 'me', ?)`, t0.UnixMilli()); err != nil {
		t.Fatal(err)
	}
	if _, err := f.s.Request(ctx, "me", "full"); !errors.Is(err, ErrTheirFriendCap) {
		t.Errorf("answering a full list's request: %v, want ErrTheirFriendCap", err)
	}
}

func TestOverviewLeavesOutDeletedAccounts(t *testing.T) {
	f := newFixture(t, "a", "b", "c")
	f.request(t, "a", "b")
	f.request(t, "b", "a")
	f.request(t, "c", "a")
	if _, err := f.d.Exec(`UPDATE users SET deleted_at = 9 WHERE id IN ('b', 'c')`); err != nil {
		t.Fatal(err)
	}
	if o := f.overview(t, "a"); len(o.Friends)+len(o.Incoming)+len(o.Outgoing) != 0 {
		t.Errorf("a tombstone is listed: %+v", o)
	}
}

func TestSelfIsNeverAnAnswer(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t, "a")
	if _, err := f.s.Request(ctx, "a", "a"); !errors.Is(err, ErrSelf) {
		t.Errorf("request to yourself: %v", err)
	}
	if err := f.s.Block(ctx, "a", "a"); !errors.Is(err, ErrSelf) {
		t.Errorf("block yourself: %v", err)
	}
	if ok, _ := f.s.Blocked(ctx, "a", "a"); ok {
		t.Error("Blocked(a, a)")
	}
}
