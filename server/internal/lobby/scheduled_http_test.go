package lobby

// Tests for scheduled tables (ADR 0215 slice 2, #3321): the routes'
// callers, the host's own actions, the caps, the RSVP order with its
// waitlist, and the scheduler driven by an injected clock: doors-open,
// the seat holds in join, a bot, #lfg and the start's confirm, the grace,
// the waitlist's promotion, a missed plan, the no-show close, retention
// and account deletion.

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/accountdelete"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/capacity"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/schedule"
)

type schedStack struct {
	podStack
	cfg *Config
	mu  sync.Mutex
	now time.Time
	eve podPerson
}

func newSchedStack(t *testing.T, extra func(*Config)) *schedStack {
	t.Helper()
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	st := &schedStack{now: time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)}
	st.userStack = newUserStackWithDB(t, nil, func(c *Config, d *db.DB) {
		c.AdminToken = "shared-admin-token"
		c.Schedule = schedule.NewStore(d)
		c.Accounts = accountdelete.NewSQLStore(d)
		if extra != nil {
			extra(c)
		}
		st.cfg = c
	})
	st.social = st.cfg.Social
	clock := func() time.Time {
		st.mu.Lock()
		defer st.mu.Unlock()
		return st.now
	}
	st.lobby.SetScheduleClock(clock)
	st.lobby.SetIdleClock(clock)
	for i, p := range []*podPerson{&st.ann, &st.bea, &st.cal, &st.dee, &st.eve} {
		name := []string{"Ann", "Bea", "Cal", "Dee", "Eve"}[i]
		snowflake := "d-" + strings.ToLower(name)
		u := dmTarget(t, st.userStack, snowflake, name)
		p.id, p.tok = u.ID, identityToken(t, st.auth, u.ID, snowflake, name)
	}
	return st
}

func (s *schedStack) set(at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.now = at
}

func (s *schedStack) tick() { s.lobby.TickSchedule(*s.cfg) }

func (s *schedStack) view(t *testing.T, raw string) scheduledView {
	t.Helper()
	var v scheduledView
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	return v
}

// schedule is POST /scheduled by host, a day ahead, and answers the plan.
func (s *schedStack) schedule(t *testing.T, host podPerson, body map[string]any) scheduledView {
	t.Helper()
	if body == nil {
		body = map[string]any{}
	}
	if _, ok := body["starts_at"]; !ok {
		body["starts_at"] = s.now.Add(24 * time.Hour).UnixMilli()
	}
	if _, ok := body["zone"]; !ok {
		body["zone"] = "Europe/London"
	}
	code, raw := s.call(t, "POST", "/scheduled", host.tok, body)
	if code != http.StatusCreated {
		t.Fatalf("POST /scheduled: %d %s", code, raw)
	}
	return s.view(t, raw)
}

func (s *schedStack) reserve(t *testing.T, plan string, who ...podPerson) {
	t.Helper()
	for _, p := range who {
		if code, raw := s.call(t, "PUT", "/scheduled/"+plan+"/rsvp", p.tok, nil); code != http.StatusOK {
			t.Fatalf("reserve: %d %s", code, raw)
		}
		// Each later reservation is a millisecond later, so the order is
		// the order of the calls.
		s.set(s.now.Add(time.Millisecond))
	}
}

// list is the plan's list as the host sees it: "Name:state ...".
func (s *schedStack) list(t *testing.T, plan string) string {
	t.Helper()
	code, raw := s.call(t, "GET", "/scheduled/"+plan, s.ann.tok, nil)
	if code != http.StatusOK {
		t.Fatalf("GET plan: %d %s", code, raw)
	}
	var out []string
	for _, r := range s.view(t, raw).List {
		out = append(out, r.DisplayName+":"+r.State)
	}
	return strings.Join(out, " ")
}

func (s *schedStack) plan(t *testing.T, id string) schedule.Plan {
	t.Helper()
	p, _, err := s.cfg.Schedule.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("Get plan: %v", err)
	}
	return p
}

func TestScheduleRoutesNeedAnAccount(t *testing.T) {
	s := newSchedStack(t, nil)
	admin := adminToken(t, s.srv)
	body := map[string]any{"starts_at": s.now.Add(24 * time.Hour).UnixMilli(), "zone": "UTC"}
	if code, _ := s.call(t, "POST", "/scheduled", admin, body); code != http.StatusForbidden {
		t.Errorf("the admin token scheduling: %d, want 403", code)
	}
	if code, _ := s.call(t, "POST", "/scheduled", "", body); code != http.StatusUnauthorized {
		t.Errorf("nobody scheduling: %d, want 401", code)
	}
	if code, _ := s.call(t, "GET", "/scheduled", "", nil); code != http.StatusUnauthorized {
		t.Errorf("nobody listing: %d, want 401", code)
	}
	p := s.schedule(t, s.ann, map[string]any{"name": "Friday pod"})
	if p.Name != "Friday pod" || p.Host != "Ann" || p.Players != 4 || p.LengthMinutes != 180 || !p.IsHost ||
		p.OpenSeats != 3 || p.MovesLeft == nil || *p.MovesLeft != 3 {
		t.Errorf("the new plan: %+v", p)
	}
	if p.DoorsOpenAt != p.StartsAt-15*60*1000 || p.HeldUntil != p.StartsAt+10*60*1000 {
		t.Errorf("doors %d, held until %d, start %d", p.DoorsOpenAt, p.HeldUntil, p.StartsAt)
	}
	if code, _ := s.call(t, "PUT", "/scheduled/"+p.ID+"/rsvp", admin, nil); code != http.StatusForbidden {
		t.Errorf("the admin token reserving: %d, want 403", code)
	}
	// The plan's page is public by id, with counts and no names.
	code, raw := s.call(t, "GET", "/scheduled/"+p.ID, "", nil)
	if code != http.StatusOK {
		t.Fatalf("a stranger's GET: %d %s", code, raw)
	}
	if v := s.view(t, raw); v.List != nil || v.You != nil || v.IsHost || v.MovesLeft != nil || v.Name != "Friday pod" {
		t.Errorf("a stranger sees %+v", v)
	}
	if code, _ := s.call(t, "GET", "/scheduled/"+uuid.NewString(), "", nil); code != http.StatusNotFound {
		t.Errorf("an unknown plan: %d", code)
	}
	if code, _ := s.call(t, "GET", "/scheduled/not-a-uuid", "", nil); code != http.StatusNotFound {
		t.Errorf("a malformed id: %d", code)
	}
}

func TestScheduleOffWithNoDatabase(t *testing.T) {
	s := newSchedStack(t, func(c *Config) { c.Schedule = nil })
	if code, _ := s.call(t, "GET", "/scheduled", s.ann.tok, nil); code != http.StatusServiceUnavailable {
		t.Errorf("GET /scheduled with no store: %d, want 503", code)
	}
	if code, _ := s.call(t, "GET", "/scheduled/"+uuid.NewString(), "", nil); code != http.StatusServiceUnavailable {
		t.Errorf("GET /scheduled/{id} with no store: %d, want 503", code)
	}
	s.tick() // nothing to do, and no panic
}

func TestTheRSVPOrderAndTheWaitlist(t *testing.T) {
	s := newSchedStack(t, nil)
	p := s.schedule(t, s.ann, map[string]any{"players": 3})
	s.reserve(t, p.ID, s.bea, s.cal, s.dee, s.eve)
	if got, want := s.list(t, p.ID), "Ann:seat Bea:seat Cal:seat Dee:waiting Eve:waiting"; got != want {
		t.Fatalf("list = %q, want %q", got, want)
	}

	// Someone on the list sees the names and their own place; a stranger
	// sees counts.
	_, raw := s.call(t, "GET", "/scheduled/"+p.ID, s.eve.tok, nil)
	v := s.view(t, raw)
	if v.You == nil || v.You.State != "waiting" || v.You.Position != 2 || len(v.List) != 5 || v.List[0].DiscordID != "" {
		t.Errorf("Eve's view: you %+v, %d rows, first %+v", v.You, len(v.List), v.List[0])
	}
	if v.OpenSeats != 0 || v.Waitlist != 2 {
		t.Errorf("counts: %d open, %d waiting", v.OpenSeats, v.Waitlist)
	}
	_, raw = s.call(t, "GET", "/scheduled/"+p.ID, s.ann.tok, nil)
	if v := s.view(t, raw); v.List[1].DiscordID != "d-bea" {
		t.Errorf("the host sees no Discord id to remove with: %+v", v.List[1])
	}

	// The list of plans shows counts, with yours first.
	other := s.schedule(t, s.cal, nil)
	code, raw := s.call(t, "GET", "/scheduled", s.eve.tok, nil)
	var all scheduledListResponse
	if err := json.Unmarshal([]byte(raw), &all); code != http.StatusOK || err != nil || len(all.Plans) != 2 {
		t.Fatalf("GET /scheduled: %d %s", code, raw)
	}
	if all.Plans[0].ID != p.ID || all.Plans[0].List != nil || all.Plans[1].ID != other.ID {
		t.Errorf("Eve's list: %+v", all.Plans)
	}

	// Cal leaves: Dee moves up at once.
	if code, raw := s.call(t, "DELETE", "/scheduled/"+p.ID+"/rsvp", s.cal.tok, nil); code != http.StatusOK {
		t.Fatalf("leave: %d %s", code, raw)
	}
	if got, want := s.list(t, p.ID), "Ann:seat Bea:seat Dee:seat Eve:waiting"; got != want {
		t.Errorf("after Cal left: %q, want %q", got, want)
	}
	if code, _ := s.call(t, "DELETE", "/scheduled/"+p.ID+"/rsvp", s.cal.tok, nil); code != http.StatusNotFound {
		t.Errorf("leaving twice: %d", code)
	}
	if code, _ := s.call(t, "DELETE", "/scheduled/"+p.ID+"/rsvp", s.ann.tok, nil); code != http.StatusConflict {
		t.Errorf("the host leaving: %d, want 409", code)
	}

	// Only the host removes; the removed person can't come back.
	if code, _ := s.call(t, "DELETE", "/scheduled/"+p.ID+"/rsvps/d-dee", s.bea.tok, nil); code != http.StatusForbidden {
		t.Errorf("a removal by a guest of the list: %d, want 403", code)
	}
	if code, raw := s.call(t, "DELETE", "/scheduled/"+p.ID+"/rsvps/d-dee", s.ann.tok, nil); code != http.StatusOK {
		t.Fatalf("the host's removal: %d %s", code, raw)
	}
	if got, want := s.list(t, p.ID), "Ann:seat Bea:seat Dee:removed Eve:seat"; got != want {
		t.Errorf("after the removal: %q, want %q", got, want)
	}
	code, raw = s.call(t, "PUT", "/scheduled/"+p.ID+"/rsvp", s.dee.tok, nil)
	if code != http.StatusForbidden || !strings.Contains(raw, errPlanRemoved) {
		t.Errorf("Dee again: %d %s", code, raw)
	}
	// A removed person is not on the list, and sees only counts.
	_, raw = s.call(t, "GET", "/scheduled/"+p.ID, s.dee.tok, nil)
	if v := s.view(t, raw); v.List != nil {
		t.Errorf("a removed person sees the list: %+v", v.List)
	}

	// A person the host blocked gets the same answer (ADR 0213 §3).
	s.block(t, s.ann, s.cal)
	code, raw = s.call(t, "PUT", "/scheduled/"+p.ID+"/rsvp", s.cal.tok, nil)
	if code != http.StatusForbidden || !strings.Contains(raw, errPlanRemoved) {
		t.Errorf("a blocked person reserving: %d %s", code, raw)
	}
}

func TestTheHostsChangesAndTheCaps(t *testing.T) {
	s := newSchedStack(t, nil)
	p := s.schedule(t, s.ann, nil)
	s.schedule(t, s.ann, nil)
	s.schedule(t, s.ann, nil)
	code, raw := s.call(t, "POST", "/scheduled", s.ann.tok, map[string]any{"starts_at": s.now.Add(time.Hour).UnixMilli(), "zone": "UTC"})
	if code != http.StatusConflict || !strings.Contains(raw, errTooManyPlans) {
		t.Errorf("a fourth upcoming plan: %d %s", code, raw)
	}
	for name, body := range map[string]map[string]any{
		"too soon":    {"starts_at": s.now.Add(10 * time.Minute).UnixMilli(), "zone": "UTC"},
		"too far":     {"starts_at": s.now.Add(31 * 24 * time.Hour).UnixMilli(), "zone": "UTC"},
		"no start":    {"zone": "UTC"},
		"bad zone":    {"starts_at": s.now.Add(time.Hour).UnixMilli(), "zone": "Nowhere/Land"},
		"one player":  {"starts_at": s.now.Add(time.Hour).UnixMilli(), "zone": "UTC", "players": 1},
		"bad length":  {"starts_at": s.now.Add(time.Hour).UnixMilli(), "zone": "UTC", "length_minutes": 30},
		"bad setup":   {"starts_at": s.now.Add(time.Hour).UnixMilli(), "zone": "UTC", "setup": "best"},
		"listed name": {"starts_at": s.now.Add(time.Hour).UnixMilli(), "zone": "UTC", "name": "retarded"},
	} {
		if code, raw := s.call(t, "POST", "/scheduled", s.bea.tok, body); code != http.StatusBadRequest {
			t.Errorf("%s: %d %s, want 400", name, code, raw)
		}
	}

	// Moves: the host only, three at most, never into the next 30 minutes.
	move := func(who podPerson, d time.Duration) (int, string) {
		return s.call(t, "PATCH", "/scheduled/"+p.ID, who.tok, map[string]any{"starts_at": s.now.Add(d).UnixMilli()})
	}
	if code, _ := move(s.bea, 48*time.Hour); code != http.StatusForbidden {
		t.Errorf("a move by someone else: %d", code)
	}
	if code, _ := move(s.ann, 20*time.Minute); code != http.StatusBadRequest {
		t.Errorf("a move into the next 30 minutes: %d", code)
	}
	for i := 1; i <= 3; i++ {
		code, raw := move(s.ann, time.Duration(24+i)*time.Hour)
		if code != http.StatusOK || *s.view(t, raw).MovesLeft != 3-i {
			t.Fatalf("move %d: %d %s", i, code, raw)
		}
	}
	if code, raw := move(s.ann, 72*time.Hour); code != http.StatusConflict || !strings.Contains(raw, errTooManyMoves) {
		t.Errorf("a fourth move: %d %s", code, raw)
	}
	// Other changes are not moves.
	code, raw = s.call(t, "PATCH", "/scheduled/"+p.ID, s.ann.tok, map[string]any{"name": "Renamed", "players": 3, "fill_bots": true})
	if v := s.view(t, raw); code != http.StatusOK || v.Name != "Renamed" || v.Players != 3 || !v.Options.FillBots {
		t.Errorf("a rename: %d %s", code, raw)
	}

	// Cancel: someone else no; the host yes; an admin cancels anyone's.
	if code, _ := s.call(t, "DELETE", "/scheduled/"+p.ID, s.bea.tok, nil); code != http.StatusForbidden {
		t.Errorf("a cancel by someone else: %d", code)
	}
	if code, raw := s.call(t, "DELETE", "/scheduled/"+p.ID, s.ann.tok, nil); code != http.StatusOK || s.view(t, raw).Status != "cancelled" {
		t.Errorf("the host's cancel: %d %s", code, raw)
	}
	if code, _ := s.call(t, "PUT", "/scheduled/"+p.ID+"/rsvp", s.bea.tok, nil); code != http.StatusConflict {
		t.Errorf("reserving a cancelled plan: %d", code)
	}
	beas := s.schedule(t, s.bea, nil)
	if code, raw := s.call(t, "DELETE", "/scheduled/"+beas.ID, adminToken(t, s.srv), nil); code != http.StatusOK {
		t.Errorf("an admin's cancel: %d %s", code, raw)
	}
	// A cancelled plan frees its host's place.
	s.schedule(t, s.ann, nil)
}

// doorsOpen schedules a three-seat table for Ann with Bea and Cal in
// the seats and Dee waiting, and opens its doors.
func (s *schedStack) doorsOpen(t *testing.T, opts map[string]any) (schedule.Plan, uuid.UUID) {
	t.Helper()
	body := map[string]any{"players": 3}
	for k, v := range opts {
		body[k] = v
	}
	p := s.schedule(t, s.ann, body)
	s.reserve(t, p.ID, s.bea, s.cal, s.dee)
	plan := s.plan(t, p.ID)
	// Not yet: a minute before doors-open.
	s.set(plan.DoorsOpenAt().Add(-time.Minute))
	s.tick()
	if got := s.plan(t, p.ID); got.Status != schedule.StatusScheduled {
		t.Fatalf("before doors-open: %s", got.Status)
	}
	s.set(plan.DoorsOpenAt())
	s.tick()
	plan = s.plan(t, p.ID)
	if plan.Status != schedule.StatusOpen || plan.GameID == "" {
		t.Fatalf("at doors-open: %s %q", plan.Status, plan.GameID)
	}
	return plan, uuid.MustParse(plan.GameID)
}

func (s *schedStack) claim(t *testing.T, plan string, who podPerson) (int, sessionResponse, string) {
	t.Helper()
	code, raw := s.call(t, "POST", "/scheduled/"+plan+"/claim", who.tok, nil)
	var out sessionResponse
	_ = json.Unmarshal([]byte(raw), &out)
	return code, out, raw
}

func TestDoorsOpenHoldTheSeatsUntilTheGraceEnds(t *testing.T) {
	s := newSchedStack(t, nil)
	plan, gid := s.doorsOpen(t, nil)
	meta := mustGet(t, s.lobby, gid)
	if meta.Name != plan.Name || meta.CreatedBy != s.ann.id || len(meta.Players) != 0 {
		t.Errorf("the table: %q by %s, %d seated; the host is not seated at doors-open", meta.Name, meta.CreatedBy, len(meta.Players))
	}
	if got, want := s.list(t, plan.ID), "Ann:held Bea:held Cal:held Dee:waiting"; got != want {
		t.Errorf("at doors-open: %q, want %q", got, want)
	}

	// Holds in join: a stranger with the invite link is refused while
	// every seat is held, with the time they run out.
	_, _, err := s.lobby.Join(gid, meta.InviteToken, "Stranger")
	var reserved *SeatsReservedError
	if !errors.As(err, &reserved) || !reserved.Until.Equal(plan.GraceEndsAt()) {
		t.Fatalf("a stranger's join: %v, want the seats reserved until %v", err, plan.GraceEndsAt())
	}
	resp := postJSON(t, s.srv, "/games/"+gid.String()+"/join", "", joinRequest{InviteToken: meta.InviteToken, Name: "Stranger"})
	var body struct {
		Code  string `json:"code"`
		Until int64  `json:"reserved_until"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusConflict || body.Code != CodeSeatsReserved || body.Until != plan.GraceEndsAt().UnixMilli() {
		t.Errorf("the join route: %d %+v", resp.StatusCode, body)
	}
	// A bot is refused the same way.
	if _, _, err := s.lobby.AddBot(gid, "Bot", "random", "", "deck", botDeck(5)); !errors.As(err, &reserved) {
		t.Errorf("a bot at a held table: %v", err)
	}
	// #lfg counts the holds as taken, and posts nothing while every free
	// seat is held.
	if lt, ok := s.lobby.lfgTable(gid); !ok || lt.post.Seated != 3 || lt.post.Seats != 3 || lt.eligible {
		t.Errorf("#lfg: seated %d of %d, eligible %v", lt.post.Seated, lt.post.Seats, lt.eligible)
	}

	// Take your seat: Bea, with no invite. Dee holds nothing.
	code, seat, raw := s.claim(t, plan.ID, s.bea)
	if code != http.StatusOK || seat.Token == "" || seat.PlayerID == uuid.Nil {
		t.Fatalf("Bea's claim: %d %s", code, raw)
	}
	if code, _, _ := s.claim(t, plan.ID, s.bea); code != http.StatusConflict {
		t.Errorf("a second claim: %d, want 409", code)
	}
	if code, _, raw := s.claim(t, plan.ID, s.dee); code != http.StatusConflict || !strings.Contains(raw, ErrNoHeldSeat.Error()) {
		t.Errorf("Dee's claim with no hold: %d %s", code, raw)
	}
	// Ann, the host, takes seat 1 by the invite link: her hold lets her
	// through, and she hosts.
	if _, pid, err := s.lobby.JoinAs(gid, meta.InviteToken, "", DiscordIdentity{ID: "d-ann", GlobalName: "Ann"}, s.ann.id); err != nil {
		t.Fatalf("Ann by the link: %v", err)
	} else if mustGet(t, s.lobby, gid).HostPlayerID != pid {
		t.Errorf("Ann sat down but does not host")
	}
	s.tick()
	if got, want := s.list(t, plan.ID), "Ann:seated Bea:seated Cal:held Dee:waiting"; got != want {
		t.Errorf("after two claims: %q, want %q", got, want)
	}

	// The grace ends: Cal's hold passes to Dee, for ten more minutes.
	s.set(plan.GraceEndsAt())
	s.tick()
	if got, want := s.list(t, plan.ID), "Ann:seated Bea:seated Cal:lapsed Dee:held"; got != want {
		t.Errorf("at the grace's end: %q, want %q", got, want)
	}
	_, raw = s.call(t, "GET", "/scheduled/"+plan.ID, s.dee.tok, nil)
	if v := s.view(t, raw); v.You == nil || v.You.HeldUntil != plan.GraceEndsAt().Add(schedule.WaitlistHold).UnixMilli() {
		t.Errorf("Dee's hold: %+v", v.You)
	}
	if _, _, err := s.lobby.Join(gid, meta.InviteToken, "Stranger"); !errors.As(err, &reserved) {
		t.Errorf("a stranger while Dee holds the last seat: %v", err)
	}
	// Dee's ten minutes run out: the plan settles and the seat is open.
	s.set(plan.GraceEndsAt().Add(schedule.WaitlistHold))
	s.tick()
	if got := s.plan(t, plan.ID); got.Status != schedule.StatusSettled {
		t.Errorf("after the last hold: %s, want settled", got.Status)
	}
	if _, _, err := s.lobby.Join(gid, meta.InviteToken, "Stranger"); err != nil {
		t.Errorf("a stranger once the holds are over: %v", err)
	}
	// The plan's three seats are the table's.
	if _, _, err := s.lobby.Join(gid, meta.InviteToken, "Fourth"); !errors.Is(err, ErrGameFull) {
		t.Errorf("a fourth seat at a three-player table: %v", err)
	}
}

func TestStartingWithSeatsHeldNeedsTheHostsConfirm(t *testing.T) {
	s := newSchedStack(t, nil)
	plan, gid := s.doorsOpen(t, nil)
	var seats []sessionResponse
	for _, who := range []podPerson{s.ann, s.bea} {
		code, seat, raw := s.claim(t, plan.ID, who)
		if code != http.StatusOK {
			t.Fatalf("claim: %d %s", code, raw)
		}
		if _, err := s.lobby.SetDeck(gid, seat.PlayerID, "Deck", botDeck(60)); err != nil {
			t.Fatal(err)
		}
		seats = append(seats, seat)
	}
	resp := postJSON(t, s.srv, "/games/"+gid.String()+"/start", seats[0].Token, nil)
	var body struct {
		Code string `json:"code"`
		Held int    `json:"held"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusConflict || body.Code != CodeSeatsHeld || body.Held != 1 {
		t.Fatalf("a start with Cal's seat held: %d %+v", resp.StatusCode, body)
	}
	if got := mustGet(t, s.lobby, gid).State; got != "lobby" {
		t.Fatalf("the refused start started the game: %s", got)
	}
	resp = postJSON(t, s.srv, "/games/"+gid.String()+"/start", seats[0].Token, map[string]bool{"without_held": true})
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("the confirmed start: %d", resp.StatusCode)
	}
	s.tick()
	if got := s.plan(t, plan.ID); got.Status != schedule.StatusSettled {
		t.Errorf("after the start: %s, want settled", got.Status)
	}
	if got, want := s.list(t, plan.ID), "Ann:seated Bea:seated Cal:lapsed Dee:waiting"; got != want {
		t.Errorf("after the start: %q, want %q", got, want)
	}
}

func TestFillWithBotsWhenTheLastHoldEnds(t *testing.T) {
	idx := buildMinimalDeckIndex(t)
	host := newFakeBotHost()
	s := newSchedStack(t, func(c *Config) {
		c.Cards, c.Bots, c.BotDecks = idx, host, fakeDeckSource{}
		c.Lobby.SetCardIndex(idx)
		c.Lobby.SetBotHost(host)
	})
	plan, gid := s.doorsOpen(t, map[string]any{"fill_bots": true})
	if code, _, raw := s.claim(t, plan.ID, s.ann); code != http.StatusOK {
		t.Fatalf("Ann's claim: %d %s", code, raw)
	}
	s.set(plan.GraceEndsAt())
	s.tick() // Bea and Cal lapse; Dee holds one seat for ten minutes
	if n := len(mustGet(t, s.lobby, gid).Players); n != 1 {
		t.Errorf("bots before the last hold ended: %d seated", n)
	}
	s.set(plan.GraceEndsAt().Add(schedule.WaitlistHold))
	s.tick()
	meta := mustGet(t, s.lobby, gid)
	bots := 0
	for _, p := range meta.Players {
		if p.IsBot {
			bots++
		}
	}
	if len(meta.Players) != 3 || bots != 2 {
		t.Errorf("after the last hold: %d seated, %d bots; want Ann and two bots", len(meta.Players), bots)
	}
}

func TestAPlanThatCannotOpenIsMissed(t *testing.T) {
	s := newSchedStack(t, nil)
	p := s.schedule(t, s.ann, nil)
	plan := s.plan(t, p.ID)
	// The server is full: the open is tried every tick until the grace
	// ends, and then the plan is missed.
	s.lobby.SetCapacity(capacity.Limits{MaxLoad: 1, TableWeight: 5}, nil)
	s.set(plan.DoorsOpenAt())
	s.tick()
	if got := s.plan(t, p.ID); got.Status != schedule.StatusScheduled {
		t.Errorf("a full server at doors-open: %s, want still scheduled", got.Status)
	}
	s.set(plan.GraceEndsAt())
	s.tick()
	if got := s.plan(t, p.ID); got.Status != schedule.StatusMissed || got.GameID != "" {
		t.Errorf("a full server at the grace's end: %s %q, want missed", got.Status, got.GameID)
	}

	// After a restart: a plan more than an hour past its start is missed
	// rather than opened.
	s.lobby.SetCapacity(capacity.Limits{}, nil)
	s.set(time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC))
	late := s.plan(t, s.schedule(t, s.bea, nil).ID)
	s.set(late.StartsAt.Add(schedule.MissedAfter))
	s.tick()
	if got := s.plan(t, late.ID); got.Status != schedule.StatusMissed {
		t.Errorf("an hour past its start: %s, want missed", got.Status)
	}
}

func TestAScheduledTableNobodyCameToIsClosed(t *testing.T) {
	s := newSchedStack(t, nil)
	plan, gid := s.doorsOpen(t, nil)
	s.set(plan.StartsAt.Add(schedule.NoShowAfter - time.Minute))
	s.tick()
	s.lobby.TickIdle()
	if mustGet(t, s.lobby, gid).Archived() {
		t.Fatal("closed before an hour after the start")
	}
	s.set(plan.StartsAt.Add(schedule.NoShowAfter))
	s.lobby.TickIdle()
	if !mustGet(t, s.lobby, gid).Archived() {
		t.Error("a scheduled table nobody came to is still open an hour after its start")
	}

	// One person at the table keeps it open.
	s2 := newSchedStack(t, nil)
	plan2, gid2 := s2.doorsOpen(t, nil)
	if code, _, raw := s2.claim(t, plan2.ID, s2.bea); code != http.StatusOK {
		t.Fatalf("claim: %d %s", code, raw)
	}
	s2.set(plan2.StartsAt.Add(schedule.NoShowAfter))
	s2.lobby.TickIdle()
	if mustGet(t, s2.lobby, gid2).Archived() {
		t.Error("a scheduled table with a person seated was closed")
	}
}

func TestPlansAreSweptADayAfterTheirStart(t *testing.T) {
	s := newSchedStack(t, nil)
	plan, gid := s.doorsOpen(t, nil)
	if code, _, raw := s.claim(t, plan.ID, s.ann); code != http.StatusOK {
		t.Fatalf("claim: %d %s", code, raw)
	}
	s.set(plan.StartsAt.Add(schedule.Keep))
	s.tick()
	if code, _ := s.call(t, "GET", "/scheduled/"+plan.ID, "", nil); code != http.StatusNotFound {
		t.Errorf("a plan a day after its start: %d, want 404", code)
	}
	// The table is an ordinary table again: four seats, no plan.
	if lt, ok := s.lobby.lfgTable(gid); !ok || lt.post.Seats != 4 {
		t.Errorf("after the sweep the table has %d seats, want 4", lt.post.Seats)
	}
}

func TestAccountDeletionTakesPlansAndRSVPs(t *testing.T) {
	s := newSchedStack(t, nil)
	// Cal hosts a plan Bea reserved on, opened with Bea's seat held.
	calsPlan, gid := func() (schedule.Plan, uuid.UUID) {
		p := s.schedule(t, s.cal, map[string]any{"players": 2})
		s.reserve(t, p.ID, s.bea, s.dee)
		plan := s.plan(t, p.ID)
		s.set(plan.DoorsOpenAt())
		s.tick()
		plan = s.plan(t, p.ID)
		return plan, uuid.MustParse(plan.GameID)
	}()
	upcoming := s.schedule(t, s.ann, nil)
	s.reserve(t, upcoming.ID, s.bea)

	// Bea deletes her account: her RSVPs go, her hold with them, and Dee
	// gets the seat.
	if code, raw := s.call(t, "DELETE", "/me/account", s.bea.tok, map[string]string{"confirm": DeleteAccountPhrase}); code != http.StatusNoContent {
		t.Fatalf("Bea's deletion: %d %s", code, raw)
	}
	if got, want := s.list(t, upcoming.ID), "Ann:seat"; got != want {
		t.Errorf("Ann's plan after Bea left: %q, want %q", got, want)
	}
	s.tick()
	_, rs, _ := s.cfg.Schedule.Get(context.Background(), calsPlan.ID)
	states := map[string]string{}
	for _, r := range rs {
		states[r.DiscordID] = string(r.State)
	}
	if _, ok := states["d-bea"]; ok || states["d-dee"] != "held" {
		t.Errorf("Cal's plan after Bea's deletion: %v", states)
	}

	// Cal deletes his: his plans go, and his table is a plain table.
	if code, raw := s.call(t, "DELETE", "/me/account", s.cal.tok, map[string]string{"confirm": DeleteAccountPhrase}); code != http.StatusNoContent {
		t.Fatalf("Cal's deletion: %d %s", code, raw)
	}
	if _, _, err := s.cfg.Schedule.Get(context.Background(), calsPlan.ID); !errors.Is(err, schedule.ErrNotFound) {
		t.Errorf("Cal's plan after his deletion: %v", err)
	}
	if lt, _ := s.lobby.lfgTable(gid); lt.post.Seats != 4 || lt.post.Seated != 0 {
		t.Errorf("Cal's table after his deletion: %d of %d", lt.post.Seated, lt.post.Seats)
	}
}
