package pods

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/moderation"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/social"
)

// Tests for ADR 0220 §2, §3, §5, §6 and §9 over migration 0033: each
// role's allowed and refused actions, the owner passing on, the ways in,
// every dead link, blocks before and after joining, the limits, expiry
// and account deletion.

var t0 = time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)

var ctx = context.Background()

type fixture struct {
	s   *Store
	soc *social.Store
	d   *db.DB
	now time.Time
}

func newFixture(t *testing.T, people ...string) *fixture {
	t.Helper()
	d, err := db.Open(ctx, t.TempDir())
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	f := &fixture{soc: social.NewStore(d), d: d, now: t0}
	f.s = NewStore(d, f.soc)
	clock := func() time.Time { return f.now }
	f.s.SetClock(clock)
	f.soc.SetClock(clock)
	for _, p := range people {
		f.user(t, p)
	}
	t.Cleanup(func() { f.ownerIsAMember(t) })
	return f
}

func (f *fixture) user(t *testing.T, id string) {
	t.Helper()
	if _, err := f.d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, ?, 1, 1)`,
		id, strings.ToUpper(id[:1])+id[1:]); err != nil {
		t.Fatal(err)
	}
}

func (f *fixture) friends(t *testing.T, a string, others ...string) {
	t.Helper()
	for _, b := range others {
		if _, err := f.soc.Request(ctx, a, b); err != nil {
			t.Fatalf("Request %s -> %s: %v", a, b, err)
		}
		if err := f.soc.Accept(ctx, b, a); err != nil {
			t.Fatalf("Accept %s <- %s: %v", b, a, err)
		}
		f.soc.Forget(a)
	}
}

func (f *fixture) block(t *testing.T, who, whom string) {
	t.Helper()
	if err := f.soc.Block(ctx, who, whom); err != nil {
		t.Fatalf("Block %s %s: %v", who, whom, err)
	}
}

func (f *fixture) create(t *testing.T, owner, name string) string {
	t.Helper()
	s, err := f.s.Create(ctx, owner, name)
	if err != nil {
		t.Fatalf("Create %s by %s: %v", name, owner, err)
	}
	f.s.Forget(owner)
	return s.ID
}

// addMember puts who straight into pod, a minute after the last change,
// so joined_at orders them.
func (f *fixture) addMember(t *testing.T, pod string, who ...string) {
	t.Helper()
	for _, w := range who {
		f.now = f.now.Add(time.Minute)
		if _, err := f.d.Exec(`INSERT INTO pod_members (pod_id, user_id, joined_at) VALUES (?, ?, ?)`,
			pod, w, ms(f.now)); err != nil {
			t.Fatal(err)
		}
	}
}

func (f *fixture) members(t *testing.T, pod string) []string {
	t.Helper()
	ids, err := memberIDs(ctx, f.d.DB, pod)
	if err != nil {
		t.Fatal(err)
	}
	return ids
}

func (f *fixture) owner(t *testing.T, pod string) string {
	t.Helper()
	var o string
	if err := f.d.QueryRow(`SELECT owner_id FROM pods WHERE id = ?`, pod).Scan(&o); err != nil {
		t.Fatalf("owner of %s: %v", pod, err)
	}
	return o
}

func (f *fixture) count(t *testing.T, q string, args ...any) int {
	t.Helper()
	var n int
	if err := f.d.QueryRow(q, args...).Scan(&n); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
	return n
}

// ownerIsAMember is §9's invariant, checked at the end of every test.
func (f *fixture) ownerIsAMember(t *testing.T) {
	t.Helper()
	if n := f.count(t, `SELECT COUNT(*) FROM pods p WHERE NOT EXISTS
		(SELECT 1 FROM pod_members m WHERE m.pod_id = p.id AND m.user_id = p.owner_id)`); n != 0 {
		t.Errorf("%d pods have an owner who is not a member", n)
	}
	if n := f.count(t, `SELECT COUNT(*) FROM pods p WHERE NOT EXISTS
		(SELECT 1 FROM pod_members m WHERE m.pod_id = p.id)`); n != 0 {
		t.Errorf("%d pods have no members", n)
	}
}

func wantErr(t *testing.T, what string, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Errorf("%s: err = %v, want %v", what, got, want)
	}
}

func ids(ms []Member) string {
	var out []string
	for _, m := range ms {
		s := m.UserID
		if m.Blocked {
			s += "(blocked)"
		}
		out = append(out, s)
	}
	return strings.Join(out, ",")
}

// --- roles (§2) --------------------------------------------------------------

func TestTheOwnerMayDoEverythingInTheTable(t *testing.T) {
	f := newFixture(t, "ann", "bea", "cal")
	f.friends(t, "ann", "cal")
	pod := f.create(t, "ann", "  Thursday  ")
	f.addMember(t, pod, "bea")

	d, err := f.s.Get(ctx, "ann", pod)
	if err != nil || d.Name != "Thursday" || d.Role != RoleOwner || ids(d.Members) != "ann,bea" ||
		d.Owner == nil || d.Owner.UserID != "ann" {
		t.Fatalf("Get = %+v, %v", d, err)
	}
	if out, err := f.s.Invite(ctx, "ann", pod, "cal"); err != nil || out != OutcomeSent {
		t.Errorf("invite a friend: %v %v", out, err)
	}
	if err := f.s.CancelInvite(ctx, "ann", pod, "cal"); err != nil {
		t.Errorf("cancel: %v", err)
	}
	token, exp, err := f.s.MintLink(ctx, "ann", pod)
	if err != nil || !exp.Equal(t0.Add(time.Minute).Add(LinkTTL)) || !linkTokenRe.MatchString(token) {
		t.Errorf("mint: %q %v %v", token, exp, err)
	}
	if d, _ := f.s.Get(ctx, "ann", pod); !d.LinkExpiresAt.Equal(exp) {
		t.Errorf("the owner's view of the link: %v, want %v", d.LinkExpiresAt, exp)
	}
	if err := f.s.ClearLink(ctx, "ann", pod); err != nil {
		t.Errorf("turn off: %v", err)
	}
	if name, err := f.s.Rename(ctx, "ann", pod, "Friday"); err != nil || name != "Friday" {
		t.Errorf("rename: %q %v", name, err)
	}
	if _, err := f.s.RemoveMember(ctx, "ann", pod, "bea", false); err != nil {
		t.Errorf("remove: %v", err)
	}
	f.addMember(t, pod, "bea")
	if err := f.s.HandOver(ctx, "ann", pod, "bea"); err != nil {
		t.Errorf("hand over: %v", err)
	}
	if o := f.owner(t, pod); o != "bea" || strings.Join(f.members(t, pod), ",") != "ann,bea" {
		t.Errorf("after the hand-over: owner %s, members %v; the old owner stays a member", o, f.members(t, pod))
	}
	if err := f.s.Delete(ctx, "bea", pod); err != nil {
		t.Errorf("delete: %v", err)
	}
	if n := f.count(t, `SELECT COUNT(*) FROM pods`) + f.count(t, `SELECT COUNT(*) FROM pod_members`); n != 0 {
		t.Errorf("a deleted pod left %d rows", n)
	}
	if _, err := f.s.Get(ctx, "ann", pod); !errors.Is(err, ErrNotFound) {
		t.Errorf("a deleted pod: %v", err)
	}
}

func TestAMemberMaySeeInviteAndLeaveAndNothingElse(t *testing.T) {
	f := newFixture(t, "ann", "bea", "cal", "dee")
	f.friends(t, "bea", "cal", "dee")
	pod := f.create(t, "ann", "Thursday")
	f.addMember(t, pod, "bea", "dee")

	if d, err := f.s.Get(ctx, "bea", pod); err != nil || d.Role != RoleMember || !d.LinkExpiresAt.IsZero() {
		t.Errorf("a member's view: %+v %v", d, err)
	}
	if _, _, err := f.s.MintLink(ctx, "ann", pod); err != nil {
		t.Fatal(err)
	}
	if d, _ := f.s.Get(ctx, "bea", pod); !d.LinkExpiresAt.IsZero() {
		t.Errorf("a member sees the link's expiry: %v", d.LinkExpiresAt)
	}
	if out, err := f.s.Invite(ctx, "bea", pod, "cal"); err != nil || out != OutcomeSent {
		t.Errorf("a member invites a friend: %v %v", out, err)
	}
	// Refused: everything that is the owner's.
	_, _, err := f.s.MintLink(ctx, "bea", pod)
	wantErr(t, "mint", err, ErrNotOwner)
	wantErr(t, "turn off", f.s.ClearLink(ctx, "bea", pod), ErrNotOwner)
	_, err = f.s.Rename(ctx, "bea", pod, "Mine now")
	wantErr(t, "rename", err, ErrNotOwner)
	_, err = f.s.RemoveMember(ctx, "bea", pod, "dee", false)
	wantErr(t, "remove", err, ErrNotOwner)
	wantErr(t, "hand over", f.s.HandOver(ctx, "bea", pod, "dee"), ErrNotOwner)
	wantErr(t, "delete", f.s.Delete(ctx, "bea", pod), ErrNotOwner)
	// Cancelling is the inviter's or the owner's: dee did not send it.
	wantErr(t, "cancel another's invite", f.s.CancelInvite(ctx, "dee", pod, "cal"), ErrNotInviter)
	if err := f.s.CancelInvite(ctx, "bea", pod, "cal"); err != nil {
		t.Errorf("the inviter cancels: %v", err)
	}
	// Leaving is anyone's.
	if r, err := f.s.RemoveMember(ctx, "bea", pod, "bea", false); err != nil || r.PodDeleted || r.NewOwnerID != "" {
		t.Errorf("leave: %+v %v", r, err)
	}
	if got := strings.Join(f.members(t, pod), ","); got != "ann,dee" {
		t.Errorf("members after bea left: %s", got)
	}
	if f.owner(t, pod) != "ann" {
		t.Error("a member leaving changed the owner")
	}
}

func TestANonMemberGetsNotFoundForEverything(t *testing.T) {
	f := newFixture(t, "ann", "out", "cal")
	f.friends(t, "out", "cal")
	pod := f.create(t, "ann", "Private")
	_, err := f.s.Get(ctx, "out", pod)
	wantErr(t, "get", err, ErrNotFound)
	_, err = f.s.Invite(ctx, "out", pod, "cal")
	wantErr(t, "invite", err, ErrNotFound)
	_, _, err = f.s.MintLink(ctx, "out", pod)
	wantErr(t, "mint", err, ErrNotFound)
	wantErr(t, "turn off", f.s.ClearLink(ctx, "out", pod), ErrNotFound)
	_, err = f.s.Rename(ctx, "out", pod, "X")
	wantErr(t, "rename", err, ErrNotFound)
	_, err = f.s.RemoveMember(ctx, "out", pod, "ann", false)
	wantErr(t, "remove", err, ErrNotFound)
	_, err = f.s.RemoveMember(ctx, "out", pod, "out", false)
	wantErr(t, "leave", err, ErrNotFound)
	wantErr(t, "hand over", f.s.HandOver(ctx, "out", pod, "ann"), ErrNotFound)
	wantErr(t, "delete", f.s.Delete(ctx, "out", pod), ErrNotFound)
	wantErr(t, "cancel", f.s.CancelInvite(ctx, "out", pod, "cal"), ErrNotFound)
	// A pod that does not exist answers the same, and so does a non-uuid.
	for _, id := range []string{"6d7c1e1e-0000-4000-8000-000000000000", "../etc", ""} {
		_, err := f.s.Get(ctx, "ann", id)
		wantErr(t, "get "+id, err, ErrNotFound)
	}
}

func TestHandOverRefusals(t *testing.T) {
	f := newFixture(t, "ann", "bea", "out")
	pod := f.create(t, "ann", "P")
	f.addMember(t, pod, "bea")
	wantErr(t, "to yourself", f.s.HandOver(ctx, "ann", pod, "ann"), ErrSelf)
	wantErr(t, "to a non-member", f.s.HandOver(ctx, "ann", pod, "out"), ErrNoMember)
	f.block(t, "bea", "ann")
	wantErr(t, "to a member with a block", f.s.HandOver(ctx, "ann", pod, "bea"), ErrNoMember)
	if f.owner(t, pod) != "ann" {
		t.Error("a refused hand-over changed the owner")
	}
}

// --- the owner passing on (§2) -----------------------------------------------

func TestTheOwnerLeavingPassesThePodToWhoJoinedFirst(t *testing.T) {
	f := newFixture(t, "ann", "bea", "cal", "dee")
	pod := f.create(t, "ann", "P")
	// cal joined before bea and dee.
	f.addMember(t, pod, "cal", "bea", "dee")
	r, err := f.s.RemoveMember(ctx, "ann", pod, "ann", false)
	if err != nil || r.NewOwnerID != "cal" || r.PodDeleted {
		t.Fatalf("the owner leaves: %+v %v", r, err)
	}
	if f.owner(t, pod) != "cal" {
		t.Errorf("owner = %s, want cal, who joined first", f.owner(t, pod))
	}
	if d, _ := f.s.Get(ctx, "bea", pod); d.Owner == nil || d.Owner.UserID != "cal" {
		t.Errorf("the pod's page shows owner %+v, want cal", d.Owner)
	}
	// And again: bea is next.
	if r, _ := f.s.RemoveMember(ctx, "cal", pod, "cal", false); r.NewOwnerID != "bea" {
		t.Errorf("the second hand-over went to %q, want bea", r.NewOwnerID)
	}
	if _, err := f.s.RemoveMember(ctx, "dee", pod, "dee", false); err != nil {
		t.Fatal(err)
	}
	// The last member leaving deletes the pod.
	r, err = f.s.RemoveMember(ctx, "bea", pod, "bea", false)
	if err != nil || !r.PodDeleted {
		t.Fatalf("the last member leaves: %+v %v", r, err)
	}
	if n := f.count(t, `SELECT COUNT(*) FROM pods`); n != 0 {
		t.Errorf("%d pods left after the last member left", n)
	}
}

func TestForgetAccountPassesOnOrDeletesOwnedPodsAndRemovesEveryRow(t *testing.T) {
	f := newFixture(t, "ann", "bea", "cal")
	shared := f.create(t, "ann", "Shared")
	f.addMember(t, shared, "bea", "cal")
	alone := f.create(t, "ann", "Alone")
	theirs := f.create(t, "bea", "Theirs")
	f.addMember(t, theirs, "ann")
	f.friends(t, "ann", "cal")
	if _, err := f.s.Invite(ctx, "ann", theirs, "cal"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.s.Invite(ctx, "bea", alone, "cal"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("bea is not in Alone: %v", err)
	}
	if _, err := f.d.Exec(`INSERT INTO pod_invites (pod_id, user_id, invited_by, created_at) VALUES (?, 'ann', 'cal', ?)`,
		shared, ms(f.now)); err != nil {
		t.Fatal(err)
	}
	if _, err := f.d.Exec(`INSERT INTO pod_suggestion_dismissals (user_id, group_key, dismissed_at) VALUES ('ann', 'k', 1), ('bea', 'k', 1)`); err != nil {
		t.Fatal(err)
	}

	tx, err := f.d.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	c, err := ForgetAccount(ctx, tx, "ann", f.now)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	want := AccountCounts{PodsPassedOn: 1, PodsDeleted: 1, Members: 3, Invites: 2, Dismissals: 1}
	if c != want {
		t.Errorf("counts = %+v, want %+v", c, want)
	}
	if f.owner(t, shared) != "bea" {
		t.Errorf("Shared passed to %s, want bea, who joined first", f.owner(t, shared))
	}
	if n := f.count(t, `SELECT COUNT(*) FROM pods WHERE id = ?`, alone); n != 0 {
		t.Error("a pod whose last member was deleted is still there")
	}
	if f.owner(t, theirs) != "bea" || strings.Join(f.members(t, theirs), ",") != "bea" {
		t.Errorf("a pod ann did not own: owner %s, members %v", f.owner(t, theirs), f.members(t, theirs))
	}
	for q, want := range map[string]int{
		`SELECT COUNT(*) FROM pod_members WHERE user_id = 'ann'`:                       0,
		`SELECT COUNT(*) FROM pod_invites WHERE user_id = 'ann' OR invited_by = 'ann'`: 0,
		`SELECT COUNT(*) FROM pod_suggestion_dismissals WHERE user_id = 'ann'`:         0,
		`SELECT COUNT(*) FROM pod_suggestion_dismissals WHERE user_id = 'bea'`:         1,
	} {
		if n := f.count(t, q); n != want {
			t.Errorf("%s = %d, want %d", q, n, want)
		}
	}
}

// --- names (§8) -------------------------------------------------------------

func TestNames(t *testing.T) {
	for _, tc := range []struct {
		in, want string
		err      error
	}{
		{"  Thursday night  ", "Thursday night", nil},
		{strings.Repeat("é", MaxNameLen), strings.Repeat("é", MaxNameLen), nil},
		{"", "", ErrNameInvalid},
		{"   ", "", ErrNameInvalid},
		{strings.Repeat("a", MaxNameLen+1), "", ErrNameInvalid},
		{"two\nlines", "", ErrNameInvalid},
		{"tab\there", "", ErrNameInvalid},
		{"line sep", "", ErrNameInvalid},
		{"bell\x07", "", ErrNameInvalid},
		{"xX_slût_Xx", "", moderation.ErrNameNotAllowed},
	} {
		got, err := CleanName(tc.in)
		if got != tc.want || !errors.Is(err, tc.err) {
			t.Errorf("CleanName(%q) = %q, %v; want %q, %v", tc.in, got, err, tc.want, tc.err)
		}
	}
	f := newFixture(t, "ann")
	_, err := f.s.Create(ctx, "ann", "xX_slût_Xx")
	wantErr(t, "create with a listed word", err, moderation.ErrNameNotAllowed)
	pod := f.create(t, "ann", "Fine")
	_, err = f.s.Rename(ctx, "ann", pod, "xX_slût_Xx")
	wantErr(t, "rename to a listed word", err, moderation.ErrNameNotAllowed)
	_, err = f.s.AdminRename(ctx, pod, "")
	wantErr(t, "an admin's empty name", err, ErrNameInvalid)
	if name, err := f.s.AdminRename(ctx, pod, " Renamed "); err != nil || name != "Renamed" {
		t.Errorf("admin rename: %q %v", name, err)
	}
	_, err = f.s.AdminRename(ctx, "6d7c1e1e-0000-4000-8000-000000000000", "X")
	wantErr(t, "admin rename of nothing", err, ErrNotFound)
	if err := f.s.AdminDelete(ctx, pod); err != nil {
		t.Errorf("admin delete: %v", err)
	}
	wantErr(t, "admin delete twice", f.s.AdminDelete(ctx, pod), ErrNotFound)
}

// --- friend invites (§3a) -----------------------------------------------------

func TestInviteAcceptDeclineAndExpiry(t *testing.T) {
	f := newFixture(t, "ann", "bea", "cal", "str")
	f.friends(t, "ann", "bea", "cal")
	f.friends(t, "bea", "cal")
	pod := f.create(t, "ann", "P")

	_, err := f.s.Invite(ctx, "ann", pod, "ann")
	wantErr(t, "yourself", err, ErrSelf)
	_, err = f.s.Invite(ctx, "ann", pod, "str")
	wantErr(t, "a stranger", err, ErrNotFriend)
	if out, err := f.s.Invite(ctx, "ann", pod, "bea"); err != nil || out != OutcomeSent {
		t.Fatalf("invite bea: %v %v", out, err)
	}
	if out, _ := f.s.Invite(ctx, "ann", pod, "bea"); out != OutcomePending {
		t.Errorf("invite again: %v, want pending", out)
	}
	o, err := f.s.Mine(ctx, "bea")
	if err != nil || len(o.Invites) != 1 || o.Invites[0].PodName != "P" || o.Invites[0].From.DisplayName != "Ann" {
		t.Fatalf("bea's invites: %+v %v", o, err)
	}
	s, err := f.s.Accept(ctx, "bea", pod)
	if err != nil || s.ID != pod || s.Members != 2 || s.Role != RoleMember {
		t.Fatalf("accept: %+v %v", s, err)
	}
	if n := f.count(t, `SELECT COUNT(*) FROM pod_invites`); n != 0 {
		t.Errorf("an accepted invite is still stored: %d", n)
	}
	_, err = f.s.Invite(ctx, "ann", pod, "bea")
	wantErr(t, "a member", err, ErrAlreadyMember)
	_, err = f.s.Accept(ctx, "bea", pod)
	wantErr(t, "accept twice", err, ErrNoInvite)

	// Decline is silent: off cal's list, still pending to the pod.
	if _, err := f.s.Invite(ctx, "bea", pod, "cal"); err != nil {
		t.Fatal(err)
	}
	if err := f.s.Decline(ctx, "cal", pod); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if o, _ := f.s.Mine(ctx, "cal"); len(o.Invites) != 0 {
		t.Errorf("a declined invite is still on cal's list: %+v", o.Invites)
	}
	if d, _ := f.s.Get(ctx, "bea", pod); len(d.Invites) != 1 || d.Invites[0].To.UserID != "cal" ||
		d.Invites[0].InvitedBy.UserID != "bea" {
		t.Errorf("the pod's pending invites after a decline: %+v", d.Invites)
	}
	if out, _ := f.s.Invite(ctx, "ann", pod, "cal"); out != OutcomePending {
		t.Errorf("inviting a decliner again: %v, want pending (the decline is silent)", out)
	}
	wantErr(t, "decline what isn't there", f.s.Decline(ctx, "str", pod), ErrNoInvite)

	// Expiry: 30 days on, the invite is gone from every read and can't be
	// accepted, and the sweep deletes it.
	f.now = f.now.Add(InviteTTL)
	if d, _ := f.s.Get(ctx, "ann", pod); len(d.Invites) != 0 {
		t.Errorf("an expired invite is still pending: %+v", d.Invites)
	}
	_, err = f.s.Accept(ctx, "cal", pod)
	wantErr(t, "accept an expired invite", err, ErrNoInvite)
	if n := f.count(t, `SELECT COUNT(*) FROM pod_invites`); n != 0 {
		t.Errorf("the read's sweep left %d expired invites", n)
	}
	if out, err := f.s.Invite(ctx, "ann", pod, "cal"); err != nil || out != OutcomeSent {
		t.Errorf("a fresh invite after expiry: %v %v", out, err)
	}
}

// --- the link (§3b) ------------------------------------------------------------

func TestLinkPreviewAndJoin(t *testing.T) {
	f := newFixture(t, "ann", "bea", "new")
	pod := f.create(t, "ann", "Thursday")
	f.addMember(t, pod, "bea")
	token, _, err := f.s.MintLink(ctx, "ann", pod)
	if err != nil {
		t.Fatal(err)
	}
	if n := f.count(t, `SELECT COUNT(*) FROM pods WHERE invite_hash = ?`, LinkHash(token)); n != 1 {
		t.Error("the link's hash is not what is stored")
	}
	if n := f.count(t, `SELECT COUNT(*) FROM pods WHERE invite_hash = ?`, token); n != 0 {
		t.Error("the token itself is stored")
	}
	p, err := f.s.Preview(ctx, "new", token)
	if err != nil || p.Name != "Thursday" || p.OwnerName != "Ann" || p.Members != 2 || p.Member || p.PodID != "" {
		t.Fatalf("preview: %+v %v", p, err)
	}
	s, err := f.s.JoinByLink(ctx, "new", token)
	if err != nil || s.ID != pod || s.Members != 3 || s.Role != RoleMember {
		t.Fatalf("join: %+v %v", s, err)
	}
	if p, _ := f.s.Preview(ctx, "new", token); !p.Member || p.PodID != pod {
		t.Errorf("a member's preview: %+v", p)
	}
	if s, err := f.s.JoinByLink(ctx, "new", token); err != nil || s.Members != 3 {
		t.Errorf("joining twice is joined: %+v %v", s, err)
	}
	if got := strings.Join(f.members(t, pod), ","); got != "ann,bea,new" {
		t.Errorf("members: %s", got)
	}
}

// TestEveryDeadLinkAnswersTheSame is §3b's "a probe learns nothing": an
// unknown, malformed, rotated, expired or turned-off link, a full pod,
// and a join refused for a block, in either direction, all answer
// ErrDeadLink, for the preview and the join alike.
func TestEveryDeadLinkAnswersTheSame(t *testing.T) {
	f := newFixture(t, "ann", "bea", "new")
	pod := f.create(t, "ann", "P")
	f.addMember(t, pod, "bea")
	live := func() string {
		tok, _, err := f.s.MintLink(ctx, "ann", pod)
		if err != nil {
			t.Fatal(err)
		}
		return tok
	}
	check := func(name, tok string) {
		t.Helper()
		_, err := f.s.Preview(ctx, "new", tok)
		wantErr(t, name+" preview", err, ErrDeadLink)
		_, err = f.s.JoinByLink(ctx, "new", tok)
		wantErr(t, name+" join", err, ErrDeadLink)
		if strings.Contains(strings.Join(f.members(t, pod), ","), "new") {
			t.Fatalf("%s let someone in", name)
		}
	}
	check("unknown", "AAAAAAAAAAAAAAAAAAAAAA")
	check("malformed", "not a token")
	check("empty", "")

	old := live()
	_ = live()
	check("rotated", old)

	tok := live()
	if err := f.s.ClearLink(ctx, "ann", pod); err != nil {
		t.Fatal(err)
	}
	check("turned off", tok)

	tok = live()
	f.now = f.now.Add(LinkTTL)
	check("expired", tok)
	if _, err := f.s.Sweep(ctx); err != nil {
		t.Fatal(err)
	}
	if n := f.count(t, `SELECT COUNT(*) FROM pods WHERE invite_hash IS NOT NULL`); n != 0 {
		t.Error("the sweep left an expired link's hash")
	}

	tok = live()
	f.block(t, "bea", "new")
	check("blocked by a member", tok)
	if err := f.soc.Unblock(ctx, "bea", "new"); err != nil {
		t.Fatal(err)
	}
	f.block(t, "new", "bea")
	check("blocking a member", tok)
	if err := f.soc.Unblock(ctx, "new", "bea"); err != nil {
		t.Fatal(err)
	}

	for i := len(f.members(t, pod)); i < MaxMembers; i++ {
		id := fmt.Sprintf("m%02d", i)
		f.user(t, id)
		f.addMember(t, pod, id)
	}
	check("full", tok)

	// The same link works again once there is room.
	if _, err := f.s.RemoveMember(ctx, "ann", pod, "m11", false); err != nil {
		t.Fatal(err)
	}
	if _, err := f.s.JoinByLink(ctx, "new", tok); err != nil {
		t.Errorf("join once there is room: %v", err)
	}
}

func TestRemovalCanReplaceTheLink(t *testing.T) {
	f := newFixture(t, "ann", "bea", "cal")
	pod := f.create(t, "ann", "P")
	f.addMember(t, pod, "bea", "cal")
	// No live link: nothing is made.
	if r, err := f.s.RemoveMember(ctx, "ann", pod, "cal", true); err != nil || r.LinkToken != "" {
		t.Errorf("remove with no link: %+v %v", r, err)
	}
	old, _, _ := f.s.MintLink(ctx, "ann", pod)
	r, err := f.s.RemoveMember(ctx, "ann", pod, "bea", true)
	if err != nil || r.LinkToken == "" || r.LinkToken == old || !r.LinkExpiresAt.Equal(f.now.Add(LinkTTL)) {
		t.Fatalf("remove with a new link: %+v %v", r, err)
	}
	_, err = f.s.JoinByLink(ctx, "bea", old)
	wantErr(t, "the removed person's old link", err, ErrDeadLink)
	if _, err := f.s.JoinByLink(ctx, "bea", r.LinkToken); err != nil {
		t.Errorf("the new link: %v", err)
	}
	_, err = f.s.RemoveMember(ctx, "ann", pod, "nobody", false)
	wantErr(t, "remove a non-member", err, ErrNoMember)
	if n := f.count(t, `SELECT COUNT(*) FROM users`); n != 3 {
		t.Errorf("removal stored something: %d users", n)
	}
}

// --- blocks (§6) ------------------------------------------------------------------

func TestABlockBeforeJoiningKeepsThePersonOut(t *testing.T) {
	f := newFixture(t, "ann", "bea", "cal", "dee")
	f.friends(t, "ann", "cal", "dee")
	pod := f.create(t, "ann", "P")
	f.addMember(t, pod, "bea")

	// cal, ann's friend, has a block with bea, another member: refused,
	// without naming who.
	f.block(t, "cal", "bea")
	_, err := f.s.Invite(ctx, "ann", pod, "cal")
	wantErr(t, "invite someone with a block with a member", err, ErrInviteRefused)
	// A block between the inviter and the friend ends the friendship.
	f.block(t, "ann", "dee")
	_, err = f.s.Invite(ctx, "ann", pod, "dee")
	wantErr(t, "invite someone you blocked", err, ErrNotFriend)

	// A block made after the invite: the invite leaves the invitee's
	// list, and accepting it is refused as if it were not there.
	if err := f.soc.Unblock(ctx, "ann", "dee"); err != nil {
		t.Fatal(err)
	}
	f.friends(t, "ann", "dee")
	if _, err := f.s.Invite(ctx, "ann", pod, "dee"); err != nil {
		t.Fatal(err)
	}
	f.block(t, "bea", "dee")
	if o, _ := f.s.Mine(ctx, "dee"); len(o.Invites) != 0 {
		t.Errorf("dee still sees an invite to a pod with someone who blocked them: %+v", o.Invites)
	}
	_, err = f.s.Accept(ctx, "dee", pod)
	wantErr(t, "accept past a block", err, ErrNoInvite)
	if got := strings.Join(f.members(t, pod), ","); got != "ann,bea" {
		t.Errorf("members: %s", got)
	}
	// bea and the others never see dee's invite; ann, who has no block,
	// still does.
	if d, _ := f.s.Get(ctx, "bea", pod); len(d.Invites) != 0 {
		t.Errorf("bea sees an invite to someone she blocked: %+v", d.Invites)
	}
	if d, _ := f.s.Get(ctx, "ann", pod); len(d.Invites) != 1 {
		t.Errorf("ann's view of the pending invite: %+v", d.Invites)
	}
}

func TestABlockAfterJoiningHidesEachFromTheOtherAndRemovesNeither(t *testing.T) {
	f := newFixture(t, "ann", "bea", "cal", "dee")
	pod := f.create(t, "ann", "P")
	f.addMember(t, pod, "bea", "cal", "dee")

	// Two members: bea blocks cal.
	f.block(t, "bea", "cal")
	if got := strings.Join(f.members(t, pod), ","); got != "ann,bea,cal,dee" {
		t.Fatalf("a block removed someone: %s", got)
	}
	for viewer, want := range map[string]string{
		"bea": "ann,bea,dee",
		"cal": "ann,cal,dee",
		"dee": "ann,bea,cal,dee",
		"ann": "ann,bea,cal,dee",
	} {
		d, err := f.s.Get(ctx, viewer, pod)
		if err != nil || ids(d.Members) != want {
			t.Errorf("%s sees %s (%v), want %s", viewer, ids(d.Members), err, want)
		}
	}
	if o, _ := f.s.Mine(ctx, "bea"); len(o.Pods) != 1 || o.Pods[0].Members != 3 {
		t.Errorf("bea's count: %+v", o.Pods)
	}

	// The owner blocks dee: the owner keeps one row, marked Blocked, with
	// nothing about them but who they are; dee no longer sees the owner.
	f.block(t, "ann", "dee")
	d, _ := f.s.Get(ctx, "ann", pod)
	if ids(d.Members) != "ann,bea,cal,dee(blocked)" {
		t.Fatalf("the owner's list: %s", ids(d.Members))
	}
	if last := d.Members[3]; last.AvatarURL != "" || !last.JoinedAt.IsZero() || last.Sharing || last.DisplayName != "Dee" {
		t.Errorf("the blocked row carries more than its name: %+v", last)
	}
	if d, _ := f.s.Get(ctx, "dee", pod); d.Owner != nil || ids(d.Members) != "bea,cal,dee" {
		t.Errorf("dee's view: owner %+v, members %s", d.Owner, ids(d.Members))
	}
	// The owner can still remove them.
	if _, err := f.s.RemoveMember(ctx, "ann", pod, "dee", false); err != nil {
		t.Errorf("the owner removes a member they blocked: %v", err)
	}

	// A member who blocks the owner is hidden from the owner like anyone
	// else, and can leave.
	f.block(t, "cal", "ann")
	if d, _ := f.s.Get(ctx, "ann", pod); ids(d.Members) != "ann,bea" {
		t.Errorf("the owner sees a member who blocked them: %s", ids(d.Members))
	}
	if _, err := f.s.RemoveMember(ctx, "cal", pod, "cal", false); err != nil {
		t.Errorf("leave after blocking the owner: %v", err)
	}
}

// --- limits (§5) -------------------------------------------------------------------

func TestTenPodsAPerson(t *testing.T) {
	f := newFixture(t, "ann", "bea")
	f.friends(t, "bea", "ann")
	for i := 0; i < MaxPodsPerPerson; i++ {
		f.now = f.now.Add(CreateEvery)
		f.create(t, "ann", fmt.Sprintf("P%d", i))
	}
	_, err := f.s.Create(ctx, "ann", "One more")
	wantErr(t, "an eleventh pod", err, ErrTooManyPods)
	theirs := f.create(t, "bea", "Theirs")
	if _, err := f.s.Invite(ctx, "bea", theirs, "ann"); err != nil {
		t.Fatal(err)
	}
	_, err = f.s.Accept(ctx, "ann", theirs)
	wantErr(t, "accept into an eleventh", err, ErrTooManyPods)
	tok, _, _ := f.s.MintLink(ctx, "bea", theirs)
	_, err = f.s.JoinByLink(ctx, "ann", tok)
	wantErr(t, "join an eleventh", err, ErrTooManyPods)
}

func TestTwelveMembersAPod(t *testing.T) {
	f := newFixture(t, "ann", "fri")
	f.friends(t, "ann", "fri")
	pod := f.create(t, "ann", "P")
	if _, err := f.s.Invite(ctx, "ann", pod, "fri"); err != nil {
		t.Fatal(err)
	}
	for i := 1; i < MaxMembers; i++ {
		id := fmt.Sprintf("m%02d", i)
		f.user(t, id)
		f.addMember(t, pod, id)
	}
	_, err := f.s.Accept(ctx, "fri", pod)
	wantErr(t, "accept into a full pod", err, ErrPodFull)
	if err := f.s.CancelInvite(ctx, "ann", pod, "fri"); err != nil {
		t.Fatal(err)
	}
	_, err = f.s.Invite(ctx, "ann", pod, "fri")
	wantErr(t, "invite into a full pod", err, ErrPodFull)
}

func TestInviteLimits(t *testing.T) {
	people := []string{"ann"}
	for i := 0; i < MaxPendingInvites+1; i++ {
		people = append(people, fmt.Sprintf("f%02d", i))
	}
	f := newFixture(t, people...)
	f.friends(t, "ann", people[1:]...)
	pod := f.create(t, "ann", "P")
	other := f.create(t, "ann", "Q")

	// 10 an hour, counted per sender across pods.
	for i := 0; i < InvitesPerHour; i++ {
		if _, err := f.s.Invite(ctx, "ann", pod, people[1+i]); err != nil {
			t.Fatalf("invite %d: %v", i, err)
		}
	}
	_, err := f.s.Invite(ctx, "ann", other, people[1+InvitesPerHour])
	wantErr(t, "the eleventh this hour", err, ErrInviteRateLimited)
	// Already pending is not a new invite, so not counted.
	if out, err := f.s.Invite(ctx, "ann", pod, people[1]); err != nil || out != OutcomePending {
		t.Errorf("pending: %v %v", out, err)
	}
	// 20 pending in a pod.
	for i := InvitesPerHour; i < MaxPendingInvites; i++ {
		if i%InvitesPerHour == 0 {
			f.now = f.now.Add(time.Hour)
		}
		if _, err := f.s.Invite(ctx, "ann", pod, people[1+i]); err != nil {
			t.Fatalf("invite %d: %v", i, err)
		}
	}
	f.now = f.now.Add(time.Hour)
	_, err = f.s.Invite(ctx, "ann", pod, people[1+MaxPendingInvites])
	wantErr(t, "the twenty-first pending", err, ErrTooManyInvites)
	if _, err := f.s.Invite(ctx, "ann", other, people[1+MaxPendingInvites]); err != nil {
		t.Errorf("another pod has room: %v", err)
	}
	f.s.Forget("ann")
	if len(f.s.invites) != 0 {
		t.Errorf("Forget left the invite window: %v", f.s.invites)
	}
}

func TestMakingPodsIsABurstOfThreeThenOneAnHour(t *testing.T) {
	f := newFixture(t, "ann")
	for i := 0; i < CreateBurst; i++ {
		if _, err := f.s.Create(ctx, "ann", fmt.Sprintf("P%d", i)); err != nil {
			t.Fatalf("pod %d: %v", i, err)
		}
	}
	_, err := f.s.Create(ctx, "ann", "Fourth")
	wantErr(t, "a fourth at once", err, ErrCreateRateLimited)
	f.now = f.now.Add(CreateEvery - time.Minute)
	_, err = f.s.Create(ctx, "ann", "Fourth")
	wantErr(t, "before the hour", err, ErrCreateRateLimited)
	f.now = f.now.Add(time.Minute)
	if _, err := f.s.Create(ctx, "ann", "Fourth"); err != nil {
		t.Errorf("after an hour: %v", err)
	}
	_, err = f.s.Create(ctx, "ann", "Fifth")
	wantErr(t, "a fifth in the same hour", err, ErrCreateRateLimited)
	f.s.Forget("ann")
	if _, err := f.s.Create(ctx, "ann", "Fifth"); err != nil {
		t.Errorf("Forget drops the allowance: %v", err)
	}
}

// --- the sweep (§9) -----------------------------------------------------------------

func TestSweepDeletesWhatHasExpired(t *testing.T) {
	f := newFixture(t, "ann", "bea")
	f.friends(t, "ann", "bea")
	pod := f.create(t, "ann", "P")
	if _, err := f.s.Invite(ctx, "ann", pod, "bea"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := f.s.MintLink(ctx, "ann", pod); err != nil {
		t.Fatal(err)
	}
	if _, err := f.d.Exec(`INSERT INTO pod_suggestion_dismissals (user_id, group_key, dismissed_at) VALUES ('ann', 'old', ?), ('ann', 'new', ?)`,
		ms(t0.Add(-DismissTTL)), ms(t0)); err != nil {
		t.Fatal(err)
	}
	c, err := f.s.Sweep(ctx)
	if err != nil || c != (SweepCounts{Dismissals: 1}) {
		t.Errorf("sweep at once: %+v %v", c, err)
	}
	f.now = f.now.Add(InviteTTL)
	c, err = f.s.Sweep(ctx)
	if err != nil || c != (SweepCounts{Invites: 1, Links: 1}) {
		t.Errorf("sweep after 30 days: %+v %v", c, err)
	}
	if n := f.count(t, `SELECT COUNT(*) FROM pod_suggestion_dismissals`); n != 1 {
		t.Errorf("the sweep took a dismissal younger than 180 days: %d left", n)
	}
}

func TestRunSweeperSweepsAtStart(t *testing.T) {
	f := newFixture(t, "ann")
	pod := f.create(t, "ann", "P")
	if _, _, err := f.s.MintLink(ctx, "ann", pod); err != nil {
		t.Fatal(err)
	}
	f.now = f.now.Add(LinkTTL + time.Minute)
	runCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() { RunSweeper(runCtx, f.s, nil, time.Hour); close(done) }()
	deadline := time.Now().Add(5 * time.Second)
	for f.count(t, `SELECT COUNT(*) FROM pods WHERE invite_hash IS NOT NULL`) != 0 {
		if time.Now().After(deadline) {
			t.Fatal("the sweeper did not clear the expired link at start")
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	<-done
}
