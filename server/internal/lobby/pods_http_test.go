package lobby

// Tests for pods_http.go (ADR 0220 slice 2, #4880): the caller rule,
// each role's allowed and refused routes, the ways in, the one answer
// for every dead link, blocks before and after joining, the admin routes
// and account deletion handing a pod on.

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/accountdelete"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/pods"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/social"
)

type podPerson struct {
	id  uuid.UUID
	tok string
}

type podStack struct {
	userStack
	social *social.Store
	pods   *pods.Store
	db     *db.DB
	ann    podPerson
	bea    podPerson
	cal    podPerson
	dee    podPerson
}

func newPodStack(t *testing.T) podStack {
	t.Helper()
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	var st podStack
	st.userStack = newUserStackWithDB(t, nil, func(c *Config, d *db.DB) {
		c.AdminToken = "shared-admin-token"
		c.Pods = pods.NewStore(d, c.Social)
		c.Accounts = accountdelete.NewSQLStore(d)
		st.social, st.pods, st.db = c.Social, c.Pods, d
	})
	for i, p := range []*podPerson{&st.ann, &st.bea, &st.cal, &st.dee} {
		name := []string{"Ann", "Bea", "Cal", "Dee"}[i]
		snowflake := "d-" + strings.ToLower(name)
		u := dmTarget(t, st.userStack, snowflake, name)
		p.id, p.tok = u.ID, identityToken(t, st.auth, u.ID, snowflake, name)
	}
	return st
}

func (s podStack) call(t *testing.T, method, path, tok string, body any) (int, string) {
	t.Helper()
	resp := do(t, s.srv, method, path, tok, body)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(raw)
}

func (s podStack) befriend(t *testing.T, a podPerson, others ...podPerson) {
	t.Helper()
	for _, b := range others {
		if _, err := s.social.Request(context.Background(), a.id.String(), b.id.String()); err != nil {
			t.Fatal(err)
		}
		if err := s.social.Accept(context.Background(), b.id.String(), a.id.String()); err != nil {
			t.Fatal(err)
		}
		s.social.Forget(a.id.String())
	}
}

func (s podStack) block(t *testing.T, who, whom podPerson) {
	t.Helper()
	if code, raw := s.call(t, "PUT", "/me/blocks/"+whom.id.String(), who.tok, nil); code != http.StatusOK {
		t.Fatalf("block: %d %s", code, raw)
	}
}

// makePod is POST /pods by owner, and answers the pod's id.
func (s podStack) makePod(t *testing.T, owner podPerson, name string) string {
	t.Helper()
	code, raw := s.call(t, "POST", "/pods", owner.tok, map[string]string{"name": name})
	if code != http.StatusCreated {
		t.Fatalf("POST /pods: %d %s", code, raw)
	}
	var out podSummary
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	s.pods.Forget(owner.id.String())
	return out.ID
}

// link is a fresh invite link's token.
func (s podStack) link(t *testing.T, owner podPerson, pod string) string {
	t.Helper()
	code, raw := s.call(t, "POST", "/pods/"+pod+"/link", owner.tok, nil)
	if code != http.StatusCreated {
		t.Fatalf("POST link: %d %s", code, raw)
	}
	var out podLinkResponse
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if out.Path != "#/pods/join/"+out.Token || out.ExpiresAt == 0 {
		t.Errorf("link response: %+v", out)
	}
	return out.Token
}

func (s podStack) join(t *testing.T, who podPerson, token string) {
	t.Helper()
	if code, raw := s.call(t, "POST", "/pods/join/"+token, who.tok, nil); code != http.StatusOK {
		t.Fatalf("join: %d %s", code, raw)
	}
}

func (s podStack) pod(t *testing.T, who podPerson, pod string) podResponse {
	t.Helper()
	code, raw := s.call(t, "GET", "/pods/"+pod, who.tok, nil)
	if code != http.StatusOK {
		t.Fatalf("GET /pods/%s: %d %s", pod, code, raw)
	}
	var out podResponse
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func memberNames(r podResponse) string {
	var out []string
	for _, m := range r.Members {
		n := m.DisplayName
		if m.Blocked {
			n += "(blocked)"
		}
		out = append(out, n)
	}
	return strings.Join(out, ",")
}

func TestPodRoutesNeedASignedInPerson(t *testing.T) {
	s := newPodStack(t)
	pod := s.makePod(t, s.ann, "Thursday")
	admin := podAdminToken(t, s.srv)
	guest := playerTokenWithUser(t, s.auth, uuid.New(), uuid.New(), uuid.Nil, "Guest")
	for _, route := range [][2]string{
		{"GET", "/me/pods"}, {"POST", "/pods"}, {"GET", "/pods/" + pod}, {"PATCH", "/pods/" + pod},
		{"DELETE", "/pods/" + pod}, {"PUT", "/pods/" + pod + "/owner"}, {"POST", "/pods/" + pod + "/invites"},
		{"DELETE", "/pods/" + pod + "/invites/" + s.bea.id.String()}, {"POST", "/me/pod-invites/" + pod + "/accept"},
		{"POST", "/me/pod-invites/" + pod + "/decline"}, {"POST", "/pods/" + pod + "/link"},
		{"DELETE", "/pods/" + pod + "/link"}, {"GET", "/pods/join/AAAAAAAAAAAAAAAAAAAAAA"},
		{"POST", "/pods/join/AAAAAAAAAAAAAAAAAAAAAA"}, {"DELETE", "/pods/" + pod + "/members/me"},
	} {
		if code, _ := s.call(t, route[0], route[1], "", nil); code != http.StatusUnauthorized {
			t.Errorf("%s %s with no credential: %d, want 401", route[0], route[1], code)
		}
		for name, tok := range map[string]string{"the admin token": admin, "a guest seat": guest} {
			code, raw := s.call(t, route[0], route[1], tok, map[string]string{})
			if code != http.StatusForbidden || !strings.Contains(raw, "pods need an account") {
				t.Errorf("%s %s as %s: %d %s, want 403", route[0], route[1], name, code, raw)
			}
		}
	}
}

func TestPodRoutesWithNoDatabaseAnswer503(t *testing.T) {
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	s := newUserStackWithDB(t, nil, func(c *Config, _ *db.DB) { c.AdminToken = "shared-admin-token" })
	u := dmTarget(t, s, "d-ann", "Ann")
	tok := identityToken(t, s.auth, u.ID, "d-ann", "Ann")
	resp := do(t, s.srv, "GET", "/me/pods", tok, nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("GET /me/pods with no pod store: %d, want 503", resp.StatusCode)
	}
	resp = do(t, s.srv, "DELETE", "/admin/pods/"+uuid.NewString(), podAdminToken(t, s.srv), nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("DELETE /admin/pods with no pod store: %d, want 503", resp.StatusCode)
	}
}

// podAdminToken is the shared admin token's session.
func podAdminToken(t *testing.T, srv *httptest.Server) string {
	t.Helper()
	resp := postJSON(t, srv, "/admin/login", "", adminLoginRequest{Token: "shared-admin-token"})
	defer resp.Body.Close()
	var sess sessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&sess); err != nil || sess.Token == "" {
		t.Fatalf("admin login: %d %v", resp.StatusCode, err)
	}
	return sess.Token
}

func TestPodOwnerAndMemberRoutes(t *testing.T) {
	s := newPodStack(t)
	s.befriend(t, s.ann, s.bea, s.cal)
	s.befriend(t, s.bea, s.dee)
	pod := s.makePod(t, s.ann, "  Thursday  ")

	// Ann invites Bea, who sees it and accepts.
	code, raw := s.call(t, "POST", "/pods/"+pod+"/invites", s.ann.tok, map[string]string{"user_id": s.bea.id.String()})
	if code != http.StatusOK || !strings.Contains(raw, `"outcome":"sent"`) || !strings.Contains(raw, `"display_name":"Bea"`) {
		t.Fatalf("invite: %d %s", code, raw)
	}
	code, raw = s.call(t, "GET", "/me/pods", s.bea.tok, nil)
	var mine myPodsResponse
	_ = json.Unmarshal([]byte(raw), &mine)
	if code != http.StatusOK || len(mine.Invites) != 1 || mine.Invites[0].PodName != "Thursday" ||
		mine.Invites[0].InvitedBy.DisplayName != "Ann" || mine.Invites[0].InvitedBy.UserID != "" ||
		len(mine.Suggestions) != 0 || mine.MaxPods != 10 || mine.MaxMembers != 12 {
		t.Fatalf("Bea's /me/pods: %d %s", code, raw)
	}
	if code, raw := s.call(t, "POST", "/me/pod-invites/"+pod+"/accept", s.bea.tok, nil); code != http.StatusOK ||
		!strings.Contains(raw, `"role":"member"`) || !strings.Contains(raw, `"member_count":2`) {
		t.Fatalf("accept: %d %s", code, raw)
	}

	// The owner's view carries the link state; the member's does not.
	if r := s.pod(t, s.ann, pod); r.Role != "owner" || r.Link == nil || r.Link.Live || r.Owner == nil ||
		r.Owner.DisplayName != "Ann" || memberNames(r) != "Ann,Bea" || r.MemberCount != 2 {
		t.Errorf("the owner's view: %+v", r)
	}
	token := s.link(t, s.ann, pod)
	if r := s.pod(t, s.ann, pod); r.Link == nil || !r.Link.Live {
		t.Errorf("the owner's link state: %+v", r.Link)
	}
	r := s.pod(t, s.bea, pod)
	if r.Role != "member" || r.Link != nil {
		t.Errorf("the member's view: %+v", r)
	}
	if strings.Contains(fmtJSON(t, r), token) {
		t.Error("the pod's page carries the link token")
	}
	for _, m := range r.Members {
		if m.Sharing == nil || *m.Sharing || m.JoinedAt == nil {
			t.Errorf("a member row: %+v; want joined_at and sharing:false", m)
		}
	}

	// Bea, a member, may invite her own friend and leave, and nothing
	// that is the owner's.
	if code, raw := s.call(t, "POST", "/pods/"+pod+"/invites", s.bea.tok, map[string]string{"user_id": s.dee.id.String()}); code != http.StatusOK {
		t.Errorf("a member invites a friend: %d %s", code, raw)
	}
	// Ann sees Dee's invite but not Dee's id: she did not send it, but she
	// owns the pod, so she can cancel it.
	if r := s.pod(t, s.ann, pod); len(r.Invites) != 1 || !r.Invites[0].CanCancel || r.Invites[0].UserID != s.dee.id.String() ||
		r.Invites[0].Mine || r.Invites[0].InvitedBy.DisplayName != "Bea" {
		t.Errorf("the owner's view of an invite: %+v", r.Invites)
	}
	for _, rt := range []struct {
		method, path string
		body         any
	}{
		{"PATCH", "/pods/" + pod, map[string]string{"name": "Mine"}},
		{"DELETE", "/pods/" + pod, nil},
		{"PUT", "/pods/" + pod + "/owner", map[string]string{"user_id": s.bea.id.String()}},
		{"POST", "/pods/" + pod + "/link", nil},
		{"DELETE", "/pods/" + pod + "/link", nil},
		{"DELETE", "/pods/" + pod + "/members/" + s.ann.id.String(), nil},
	} {
		if code, raw := s.call(t, rt.method, rt.path, s.bea.tok, rt.body); code != http.StatusForbidden ||
			!strings.Contains(raw, errNotPodOwner) {
			t.Errorf("a member's %s %s: %d %s, want 403", rt.method, rt.path, code, raw)
		}
	}

	// Cal, outside, gets 404 for every pod route, the same as for a pod
	// that does not exist.
	missing := uuid.NewString()
	for _, id := range []string{pod, missing} {
		for _, rt := range [][2]string{{"GET", "/pods/" + id}, {"PATCH", "/pods/" + id}, {"DELETE", "/pods/" + id},
			{"POST", "/pods/" + id + "/link"}, {"DELETE", "/pods/" + id + "/members/me"}} {
			var body any
			if rt[0] == "PATCH" {
				body = map[string]string{"name": "X"}
			}
			code, raw := s.call(t, rt[0], rt[1], s.cal.tok, body)
			if code != http.StatusNotFound || !strings.Contains(raw, errNoSuchPod) {
				t.Errorf("an outsider's %s %s: %d %s, want 404", rt[0], rt[1], code, raw)
			}
		}
	}

	// The owner renames, hands over and stays a member.
	if code, raw := s.call(t, "PATCH", "/pods/"+pod, s.ann.tok, map[string]string{"name": "Friday"}); code != http.StatusOK ||
		!strings.Contains(raw, `"name":"Friday"`) {
		t.Errorf("rename: %d %s", code, raw)
	}
	if code, raw := s.call(t, "PUT", "/pods/"+pod+"/owner", s.ann.tok, map[string]string{"user_id": s.bea.id.String()}); code != http.StatusNoContent {
		t.Fatalf("hand over: %d %s", code, raw)
	}
	if r := s.pod(t, s.ann, pod); r.Role != "member" || r.Owner == nil || r.Owner.DisplayName != "Bea" {
		t.Errorf("after the hand-over: %+v", r)
	}
	// Bea, now the owner, removes Ann; Ann can no longer read the pod.
	if code, raw := s.call(t, "DELETE", "/pods/"+pod+"/members/"+s.ann.id.String(), s.bea.tok, nil); code != http.StatusNoContent {
		t.Fatalf("remove: %d %s", code, raw)
	}
	if code, _ := s.call(t, "GET", "/pods/"+pod, s.ann.tok, nil); code != http.StatusNotFound {
		t.Errorf("a removed member reads the pod: %d", code)
	}
	if code, _ := s.call(t, "DELETE", "/pods/"+pod, s.bea.tok, nil); code != http.StatusNoContent {
		t.Errorf("delete: %d", code)
	}
}

func fmtJSON(t *testing.T, v any) string {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestPodLeaveHandsTheOwnerOnAndTheLastLeaveDeletes(t *testing.T) {
	s := newPodStack(t)
	pod := s.makePod(t, s.ann, "P")
	token := s.link(t, s.ann, pod)
	s.join(t, s.cal, token)
	s.join(t, s.bea, token)
	if code, raw := s.call(t, "DELETE", "/pods/"+pod+"/members/me", s.ann.tok, nil); code != http.StatusNoContent {
		t.Fatalf("the owner leaves: %d %s", code, raw)
	}
	if r := s.pod(t, s.bea, pod); r.Owner == nil || r.Owner.DisplayName != "Cal" || memberNames(r) != "Cal,Bea" {
		t.Errorf("after the owner left: owner %+v, members %s; want Cal, who joined first", r.Owner, memberNames(r))
	}
	for _, p := range []podPerson{s.bea, s.cal} {
		if code, raw := s.call(t, "DELETE", "/pods/"+pod+"/members/"+p.id.String(), p.tok, nil); code != http.StatusNoContent {
			t.Fatalf("leave by own id: %d %s", code, raw)
		}
	}
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM pods`).Scan(&n); err != nil || n != 0 {
		t.Errorf("%d pods after the last member left (%v)", n, err)
	}
}

func TestPodRemovalCanAnswerANewLink(t *testing.T) {
	s := newPodStack(t)
	pod := s.makePod(t, s.ann, "P")
	old := s.link(t, s.ann, pod)
	s.join(t, s.bea, old)
	code, raw := s.call(t, "DELETE", "/pods/"+pod+"/members/"+s.bea.id.String()+"?new_link=1", s.ann.tok, nil)
	var out map[string]podLinkResponse
	_ = json.Unmarshal([]byte(raw), &out)
	if code != http.StatusOK || out["link"].Token == "" || out["link"].Token == old {
		t.Fatalf("remove with a new link: %d %s", code, raw)
	}
	if code, _ := s.call(t, "POST", "/pods/join/"+old, s.bea.tok, nil); code != http.StatusNotFound {
		t.Errorf("the removed member rejoined on the old link: %d", code)
	}
}

// TestEveryDeadPodLinkGetsOneAnswer is §3b: an unknown, malformed,
// rotated, expired or turned-off link, a full pod and a join refused for
// a block, either way, answer the same status and the same body, for the
// preview and the join.
func TestEveryDeadPodLinkGetsOneAnswer(t *testing.T) {
	s := newPodStack(t)
	pod := s.makePod(t, s.ann, "P")
	s.join(t, s.bea, s.link(t, s.ann, pod))
	want := `{"error":"This invite link can't be used."}`
	check := func(name, token string) {
		t.Helper()
		for _, method := range []string{"GET", "POST"} {
			code, raw := s.call(t, method, "/pods/join/"+token, s.cal.tok, nil)
			if code != http.StatusNotFound || strings.TrimSpace(raw) != want {
				t.Errorf("%s, %s: %d %s; want 404 %s", name, method, code, raw, want)
			}
		}
	}
	check("unknown", "AAAAAAAAAAAAAAAAAAAAAA")
	check("malformed", "short")
	old := s.link(t, s.ann, pod)
	cur := s.link(t, s.ann, pod)
	check("rotated", old)
	if code, _ := s.call(t, "DELETE", "/pods/"+pod+"/link", s.ann.tok, nil); code != http.StatusNoContent {
		t.Fatal("turn off")
	}
	check("turned off", cur)

	cur = s.link(t, s.ann, pod)
	s.block(t, s.bea, s.cal)
	check("blocked by a member", cur)
	_, _ = s.call(t, "DELETE", "/me/blocks/"+s.cal.id.String(), s.bea.tok, nil)
	s.block(t, s.cal, s.bea)
	check("blocking a member", cur)
	_, _ = s.call(t, "DELETE", "/me/blocks/"+s.bea.id.String(), s.cal.tok, nil)

	// Full: fill the pod to 12 by hand.
	for i := 2; i < pods.MaxMembers; i++ {
		u := dmTarget(t, s.userStack, "d-m"+string(rune('a'+i)), "M"+string(rune('a'+i)))
		if _, err := s.db.Exec(`INSERT INTO pod_members (pod_id, user_id, joined_at) VALUES (?, ?, ?)`, pod, u.ID.String(), 10+i); err != nil {
			t.Fatal(err)
		}
	}
	check("full", cur)
	if _, err := s.db.Exec(`DELETE FROM pod_members WHERE pod_id = ? AND joined_at = 12`, pod); err != nil {
		t.Fatal(err)
	}

	// Expired: move the link's expiry into the past.
	if _, err := s.db.Exec(`UPDATE pods SET invite_expires_at = 1 WHERE id = ?`, pod); err != nil {
		t.Fatal(err)
	}
	check("expired", cur)

	// And a live one works for the same person, after all of that.
	cur = s.link(t, s.ann, pod)
	code, raw := s.call(t, "GET", "/pods/join/"+cur, s.cal.tok, nil)
	if code != http.StatusOK || !strings.Contains(raw, `"name":"P"`) || !strings.Contains(raw, `"owner_name":"Ann"`) ||
		!strings.Contains(raw, `"member":false`) || strings.Contains(raw, pod) {
		t.Errorf("the preview: %d %s", code, raw)
	}
	s.join(t, s.cal, cur)
	if code, raw := s.call(t, "GET", "/pods/join/"+cur, s.cal.tok, nil); code != http.StatusOK ||
		!strings.Contains(raw, `"member":true`) || !strings.Contains(raw, pod) {
		t.Errorf("a member's preview: %d %s", code, raw)
	}
}

func TestPodBlocksBeforeAndAfterJoining(t *testing.T) {
	s := newPodStack(t)
	s.befriend(t, s.ann, s.cal)
	pod := s.makePod(t, s.ann, "P")
	s.join(t, s.bea, s.link(t, s.ann, pod))
	s.join(t, s.dee, s.link(t, s.ann, pod))

	// Before: Cal has blocked Bea, a member. Ann's invite is refused in
	// words that name nobody.
	s.block(t, s.cal, s.bea)
	code, raw := s.call(t, "POST", "/pods/"+pod+"/invites", s.ann.tok, map[string]string{"user_id": s.cal.id.String()})
	if code != http.StatusForbidden || !strings.Contains(raw, errPodInviteRefused) ||
		strings.Contains(strings.ToLower(raw), "block") || strings.Contains(raw, "Bea") {
		t.Errorf("invite past a block: %d %s", code, raw)
	}

	// After: Bea blocks Dee. Both stay; each is gone from the other's view.
	s.block(t, s.bea, s.dee)
	if r := s.pod(t, s.bea, pod); memberNames(r) != "Ann,Bea" || r.MemberCount != 2 {
		t.Errorf("Bea's view: %s (%d)", memberNames(r), r.MemberCount)
	}
	if r := s.pod(t, s.dee, pod); memberNames(r) != "Ann,Dee" {
		t.Errorf("Dee's view: %s", memberNames(r))
	}
	if r := s.pod(t, s.ann, pod); memberNames(r) != "Ann,Bea,Dee" {
		t.Errorf("Ann's view: %s", memberNames(r))
	}

	// The owner blocks Dee: one row, marked blocked, with Remove's id and
	// nothing else.
	s.block(t, s.ann, s.dee)
	r := s.pod(t, s.ann, pod)
	if memberNames(r) != "Ann,Bea,Dee(blocked)" {
		t.Fatalf("the owner's list: %s", memberNames(r))
	}
	if last := r.Members[2]; last.UserID != s.dee.id.String() || last.JoinedAt != nil || last.Sharing != nil || last.AvatarURL != "" {
		t.Errorf("the blocked row: %+v", last)
	}
	if r := s.pod(t, s.dee, pod); r.Owner != nil || memberNames(r) != "Dee" {
		t.Errorf("Dee's view once the owner blocked her: owner %+v, members %s", r.Owner, memberNames(r))
	}
	if code, raw := s.call(t, "DELETE", "/pods/"+pod+"/members/"+s.dee.id.String(), s.ann.tok, nil); code != http.StatusNoContent {
		t.Errorf("the owner removes a member they blocked: %d %s", code, raw)
	}
}

func TestPodNamesAndLimitsOverHTTP(t *testing.T) {
	s := newPodStack(t)
	code, raw := s.call(t, "POST", "/pods", s.ann.tok, map[string]string{"name": "xX_slût_Xx"})
	var body struct{ Error, Code string }
	_ = json.Unmarshal([]byte(raw), &body)
	if code != http.StatusBadRequest || body.Code != CodeNameNotAllowed || strings.Contains(raw, "slût") {
		t.Errorf("a listed name: %d %s", code, raw)
	}
	for _, name := range []string{"", "   ", strings.Repeat("x", 41), "two\nlines"} {
		if code, raw := s.call(t, "POST", "/pods", s.ann.tok, map[string]string{"name": name}); code != http.StatusBadRequest ||
			!strings.Contains(raw, errPodNameInvalid) {
			t.Errorf("name %q: %d %s", name, code, raw)
		}
	}
	if code, raw := s.call(t, "POST", "/pods", s.ann.tok, map[string]string{"name": "P", "suggestion": "abc"}); code != http.StatusNotFound {
		t.Errorf("an unknown suggestion: %d %s", code, raw)
	}
	// A burst of three, then one an hour.
	for i := 0; i < pods.CreateBurst; i++ {
		if code, raw := s.call(t, "POST", "/pods", s.bea.tok, map[string]string{"name": "B"}); code != http.StatusCreated {
			t.Fatalf("pod %d: %d %s", i, code, raw)
		}
	}
	if code, raw := s.call(t, "POST", "/pods", s.bea.tok, map[string]string{"name": "B"}); code != http.StatusTooManyRequests ||
		!strings.Contains(raw, errTooManyNewPods) {
		t.Errorf("a fourth pod at once: %d %s", code, raw)
	}
	// Ten pods: the eleventh is refused in the ADR's words.
	for i := 0; i < pods.MaxPodsPerPerson; i++ {
		s.makePod(t, s.cal, "C")
	}
	pod := s.makePod(t, s.ann, "Room")
	if code, raw := s.call(t, "POST", "/pods/join/"+s.link(t, s.ann, pod), s.cal.tok, nil); code != http.StatusConflict ||
		!strings.Contains(raw, "You're in 10 pods. Leave one to join another.") {
		t.Errorf("an eleventh pod by link: %d %s", code, raw)
	}
}

func TestAdminPodRoutes(t *testing.T) {
	s := newPodStack(t)
	pod := s.makePod(t, s.ann, "Bad name")
	admin := podAdminToken(t, s.srv)
	if code, _ := s.call(t, "PUT", "/admin/pods/"+pod+"/name", s.ann.tok, map[string]string{"name": "X"}); code != http.StatusForbidden {
		t.Errorf("a non-admin's rename: %d", code)
	}
	if code, raw := s.call(t, "PUT", "/admin/pods/"+pod+"/name", admin, map[string]string{"name": " Renamed "}); code != http.StatusOK ||
		!strings.Contains(raw, `"name":"Renamed"`) {
		t.Errorf("admin rename: %d %s", code, raw)
	}
	if r := s.pod(t, s.ann, pod); r.Name != "Renamed" {
		t.Errorf("the pod's name after an admin rename: %q", r.Name)
	}
	if code, _ := s.call(t, "DELETE", "/admin/pods/"+uuid.NewString(), admin, nil); code != http.StatusNotFound {
		t.Errorf("admin delete of nothing: %d", code)
	}
	if code, _ := s.call(t, "DELETE", "/admin/pods/"+pod, admin, nil); code != http.StatusNoContent {
		t.Errorf("admin delete: %d", code)
	}
	if code, _ := s.call(t, "GET", "/pods/"+pod, s.ann.tok, nil); code != http.StatusNotFound {
		t.Errorf("a pod an admin deleted: %d", code)
	}
}

func TestDeletingAnAccountHandsItsPodOn(t *testing.T) {
	s := newPodStack(t)
	pod := s.makePod(t, s.ann, "P")
	token := s.link(t, s.ann, pod)
	s.join(t, s.dee, token)
	s.join(t, s.bea, token)
	if code, raw := s.call(t, "DELETE", "/me/account", s.ann.tok, map[string]string{"confirm": DeleteAccountPhrase}); code != http.StatusNoContent {
		t.Fatalf("delete the account: %d %s", code, raw)
	}
	r := s.pod(t, s.bea, pod)
	if r.Owner == nil || r.Owner.DisplayName != "Dee" || memberNames(r) != "Dee,Bea" {
		t.Errorf("after the owner's account went: owner %+v, members %s; want Dee, who joined first", r.Owner, memberNames(r))
	}
}

// TestPodSuggestionRoutes is ADR 0220 §4 over HTTP: the suggestion on
// GET /me/pods, naming the others and no ids, Make this pod through
// POST /pods with its key (inviting people who are not friends), and Not
// now.
func TestPodSuggestionRoutes(t *testing.T) {
	s := newPodStack(t)
	now := time.Now()
	for i := 0; i < 3; i++ {
		gid := uuid.NewString()
		if _, err := s.db.Exec(`INSERT INTO games (id, name, state, created_at, ended_at) VALUES (?, 'T', 'ended', 1, ?)`,
			gid, now.Add(-time.Duration(i+1)*time.Hour).UnixMilli()); err != nil {
			t.Fatal(err)
		}
		for seat, u := range []podPerson{s.ann, s.bea, s.cal} {
			if _, err := s.db.Exec(`INSERT INTO seats (game_id, seat, player_id, user_id) VALUES (?, ?, ?, ?)`,
				gid, seat, "p", u.id.String()); err != nil {
				t.Fatal(err)
			}
		}
	}
	mine := func(who podPerson) myPodsResponse {
		t.Helper()
		code, raw := s.call(t, "GET", "/me/pods", who.tok, nil)
		if code != http.StatusOK {
			t.Fatalf("GET /me/pods: %d %s", code, raw)
		}
		var out myPodsResponse
		if err := json.Unmarshal([]byte(raw), &out); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(raw, s.bea.id.String()) || strings.Contains(raw, s.cal.id.String()) {
			t.Errorf("a suggestion names no account id: %s", raw)
		}
		return out
	}
	got := mine(s.ann)
	if len(got.Suggestions) != 1 || got.Suggestions[0].Games != 3 || got.Suggestions[0].Days != 30 ||
		len(got.Suggestions[0].People) != 2 || got.Suggestions[0].People[0].DisplayName != "Bea" {
		t.Fatalf("ann's suggestions: %+v", got.Suggestions)
	}
	if n := len(mine(s.dee).Suggestions); n != 0 {
		t.Errorf("dee sees %d", n)
	}
	key := got.Suggestions[0].Key

	// Not now hides it from Bea only.
	if code, raw := s.call(t, "POST", "/me/pod-suggestions/"+key+"/dismiss", s.bea.tok, nil); code != http.StatusNoContent {
		t.Fatalf("dismiss: %d %s", code, raw)
	}
	if n := len(mine(s.bea).Suggestions); n != 0 {
		t.Errorf("bea still sees %d", n)
	}
	if n := len(mine(s.ann).Suggestions); n != 1 {
		t.Errorf("ann sees %d", n)
	}
	if code, _ := s.call(t, "POST", "/me/pod-suggestions/"+key+"/dismiss", s.dee.tok, nil); code != http.StatusNotFound {
		t.Errorf("dee dismissing: %d", code)
	}
	if code, _ := s.call(t, "POST", "/me/pod-suggestions/"+key+"/dismiss", "", nil); code != http.StatusUnauthorized && code != http.StatusForbidden {
		t.Errorf("signed out dismissing: %d", code)
	}

	// Make this pod: ann owns it, Bea and Cal get invites without being friends.
	if code, raw := s.call(t, "POST", "/pods", s.dee.tok, map[string]string{"name": "X", "suggestion": key}); code != http.StatusNotFound {
		t.Errorf("dee making ann's suggestion: %d %s", code, raw)
	}
	code, raw := s.call(t, "POST", "/pods", s.ann.tok, map[string]string{"name": "Thursday", "suggestion": key})
	if code != http.StatusCreated {
		t.Fatalf("make from suggestion: %d %s", code, raw)
	}
	for _, who := range []podPerson{s.bea, s.cal} {
		g := mine(who)
		if len(g.Invites) != 1 || g.Invites[0].PodName != "Thursday" || g.Invites[0].InvitedBy.DisplayName != "Ann" {
			t.Errorf("invites: %+v", g.Invites)
		}
	}
	if n := len(mine(s.dee).Invites); n != 0 {
		t.Errorf("dee has %d invites", n)
	}
}
