package lobby

// scheduled_http.go is ADR 0215 slice 2 (#3321): scheduled tables with
// site RSVPs, over internal/schedule.
//
//	GET    /scheduled                          signed in: the upcoming plans, yours first
//	POST   /scheduled                          signed in: schedule a table
//	GET    /scheduled/{id}                     anyone with the id: the plan
//	PATCH  /scheduled/{id}                     host: rename, move, resize, re-option
//	DELETE /scheduled/{id}                     host or admin: cancel
//	PUT    /scheduled/{id}/rsvp                signed in: reserve a seat
//	DELETE /scheduled/{id}/rsvp                signed in: leave
//	DELETE /scheduled/{id}/rsvps/{discord_id}  host: remove a reservation
//	POST   /scheduled/{id}/claim               a holder: take your seat
//
// WHO. Scheduling, reserving and claiming need a signed-in person (an
// account that passed the gate, ADR 0215 §2). A guest, an agent seat and
// the admin token are credentials, not people, and get 403; an admin may
// still cancel any plan. With no database (c.Schedule nil) every route
// answers 503.
//
// WHO SEES NAMES (§11). GET /scheduled/{id} is public by id, because it
// is the Discord event's link: anyone sees the name, the host's display
// name, the times, the players and the counts. The host and the people
// on the list also see the list, in order, and the table it became; the
// host also sees each person's Discord id, which Remove takes. GET
// /scheduled shows counts only.
//
// TIMES are Unix milliseconds, UTC. The host's IANA zone rides along so
// the client can say "4:30 PM for the host" (§8).

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/capacity"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/moderation"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/schedule"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/ratelimit"
)

// The words of the refusals (ADR 0215 §3, §10, §12).
const (
	errScheduleOff         = "scheduled tables are not available on this server"
	errScheduleNeedAccount = "scheduled tables need an account; sign in with Discord"
	errNoSuchPlan          = "no such scheduled table"
	errNotPlanHost         = "only the host can do that"
	errPlanName            = "a table name is 1 to 80 characters on one line"
	errPlanTooSoon         = "The start must be at least 30 minutes from now."
	errPlanTooFar          = "A table can be scheduled at most 30 days ahead."
	errPlanLength          = "the length is 120, 180 or 240 minutes"
	errPlanPlayers         = "a table has 2 to 4 players"
	errPlanZone            = "unknown time zone; send an IANA name such as Europe/London"
	errPlanSetup           = `setup is "last" or nothing`
	errTooManyPlans        = "You have 3 upcoming scheduled tables. Cancel one to schedule another."
	errCalendarFull        = "The calendar is full"
	errPlanRateLimited     = "you're scheduling or moving tables quickly; try again shortly"
	errTooManyMoves        = "This table has been moved 3 times. Cancel it and schedule a new one instead."
	errDoorsOpen           = "The doors have opened: this is a table now, and the table's own controls manage it."
	errPlanClosed          = "This scheduled table is closed."
	errPlanIsHost          = "You're the host: you hold seat 1."
	errPlanRemoved         = "This table can't take your reservation"
	errNotOnPlanList       = "not on this table's list"
	errPlanNoDiscord       = "reserving a seat needs a Discord account linked to yours"
	errStartsAtRequired    = "starts_at is required (Unix milliseconds)"
	errNoDoorsYet          = "the doors haven't opened yet; your seat is held from 15 minutes before the start"
)

// scheduledPerson is one row of a plan's list, for the host and the
// people on it.
type scheduledPerson struct {
	DisplayName string `json:"display_name"`
	State       string `json:"state"`
	Host        bool   `json:"host,omitempty"`
	You         bool   `json:"you,omitempty"`
	HeldUntil   int64  `json:"held_until,omitempty"`
	// DiscordID is the host's only: Remove takes it.
	DiscordID string `json:"discord_id,omitempty"`
}

// scheduledYou is the caller's own place on the list. Position counts
// from 1 within the seats (seat, held, seated) or within the waitlist,
// and is 0 for any other state.
type scheduledYou struct {
	State     string `json:"state"`
	Position  int    `json:"position,omitempty"`
	HeldUntil int64  `json:"held_until,omitempty"`
}

// scheduledView is one plan as GET /scheduled/{id} and every write
// answer it.
type scheduledView struct {
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	Host          string           `json:"host"`
	HostZone      string           `json:"host_zone"`
	StartsAt      int64            `json:"starts_at"`
	EndsAt        int64            `json:"ends_at"`
	DoorsOpenAt   int64            `json:"doors_open_at"`
	HeldUntil     int64            `json:"held_until"`
	LengthMinutes int              `json:"length_minutes"`
	Players       int              `json:"players"`
	Status        string           `json:"status"`
	OpenSeats     int              `json:"open_seats"`
	Waitlist      int              `json:"waitlist"`
	Options       schedule.Options `json:"options"`
	IsHost        bool             `json:"is_host,omitempty"`
	// MovesLeft is the host's only.
	MovesLeft *int `json:"moves_left,omitempty"`
	// GameID is the table the plan became, for the host and the people
	// on the list.
	GameID string            `json:"game_id,omitempty"`
	You    *scheduledYou     `json:"you,omitempty"`
	List   []scheduledPerson `json:"list,omitempty"`
}

// scheduledListResponse is GET /scheduled.
type scheduledListResponse struct {
	Plans              []scheduledView `json:"plans"`
	MaxUpcomingPerHost int             `json:"max_upcoming_per_host"`
}

// scheduleBody is POST /scheduled.
type scheduleBody struct {
	Name          string `json:"name"`
	StartsAt      int64  `json:"starts_at"`
	LengthMinutes int    `json:"length_minutes,omitempty"`
	Zone          string `json:"zone"`
	Players       int    `json:"players,omitempty"`
	VoiceRoom     bool   `json:"voice_room,omitempty"`
	LFG           bool   `json:"lfg,omitempty"`
	Results       bool   `json:"results,omitempty"`
	Planechase    bool   `json:"planechase,omitempty"`
	Setup         string `json:"setup,omitempty"`
	FillBots      bool   `json:"fill_bots,omitempty"`
}

// schedulePatchBody is PATCH /scheduled/{id}: each field present is
// changed, each absent one kept.
type schedulePatchBody struct {
	Name          *string `json:"name,omitempty"`
	StartsAt      *int64  `json:"starts_at,omitempty"`
	LengthMinutes *int    `json:"length_minutes,omitempty"`
	Zone          *string `json:"zone,omitempty"`
	Players       *int    `json:"players,omitempty"`
	VoiceRoom     *bool   `json:"voice_room,omitempty"`
	LFG           *bool   `json:"lfg,omitempty"`
	Results       *bool   `json:"results,omitempty"`
	Planechase    *bool   `json:"planechase,omitempty"`
	Setup         *string `json:"setup,omitempty"`
	FillBots      *bool   `json:"fill_bots,omitempty"`
}

func (b schedulePatchBody) options() bool {
	return b.VoiceRoom != nil || b.LFG != nil || b.Results != nil || b.Planechase != nil || b.Setup != nil || b.FillBots != nil
}

// scheduleCaller is the signed-in person behind a schedule route, once
// the feature is known to be on.
func scheduleCaller(c Config, r *http.Request) (auth.Principal, error) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return auth.Principal{}, httpError(http.StatusInternalServerError, "missing principal")
	}
	if c.Schedule == nil {
		return auth.Principal{}, httpError(http.StatusServiceUnavailable, errScheduleOff)
	}
	if !isSignedInPerson(p) {
		return auth.Principal{}, httpError(http.StatusForbidden, errScheduleNeedAccount)
	}
	return p, nil
}

// scheduleError maps a schedule error onto the route's answer.
func scheduleError(c Config, w http.ResponseWriter, err error, what string) error {
	switch {
	case errors.Is(err, moderation.ErrNameNotAllowed):
		// writeLobbyError answers it 400 with code name_not_allowed.
		return err
	case errors.Is(err, schedule.ErrNotFound):
		return httpError(http.StatusNotFound, errNoSuchPlan)
	case errors.Is(err, schedule.ErrNotHost):
		return httpError(http.StatusForbidden, errNotPlanHost)
	case errors.Is(err, schedule.ErrNameInvalid):
		return httpError(http.StatusBadRequest, errPlanName)
	case errors.Is(err, schedule.ErrTooSoon):
		return httpError(http.StatusBadRequest, errPlanTooSoon)
	case errors.Is(err, schedule.ErrTooFar):
		return httpError(http.StatusBadRequest, errPlanTooFar)
	case errors.Is(err, schedule.ErrBadLength):
		return httpError(http.StatusBadRequest, errPlanLength)
	case errors.Is(err, schedule.ErrBadPlayers):
		return httpError(http.StatusBadRequest, errPlanPlayers)
	case errors.Is(err, schedule.ErrBadZone):
		return httpError(http.StatusBadRequest, errPlanZone)
	case errors.Is(err, schedule.ErrBadSetup):
		return httpError(http.StatusBadRequest, errPlanSetup)
	case errors.Is(err, schedule.ErrTooManyPlans):
		return httpError(http.StatusConflict, errTooManyPlans)
	case errors.Is(err, schedule.ErrCalendarFull):
		return httpError(http.StatusConflict, errCalendarFull)
	case errors.Is(err, schedule.ErrRateLimited):
		w.Header().Set("Retry-After", "30")
		return httpError(http.StatusTooManyRequests, errPlanRateLimited)
	case errors.Is(err, schedule.ErrTooManyMoves):
		return httpError(http.StatusConflict, errTooManyMoves)
	case errors.Is(err, schedule.ErrDoorsOpen):
		return httpError(http.StatusConflict, errDoorsOpen)
	case errors.Is(err, schedule.ErrClosed):
		return httpError(http.StatusConflict, errPlanClosed)
	case errors.Is(err, schedule.ErrIsHost):
		return httpError(http.StatusConflict, errPlanIsHost)
	case errors.Is(err, schedule.ErrRemoved):
		return httpError(http.StatusForbidden, errPlanRemoved)
	case errors.Is(err, schedule.ErrNotOnList):
		return httpError(http.StatusNotFound, errNotOnPlanList)
	case errors.Is(err, schedule.ErrNoDiscord):
		return httpError(http.StatusForbidden, errPlanNoDiscord)
	}
	c.logger().Error(what+" failed", "err", err)
	return httpError(http.StatusInternalServerError, "something went wrong; try again")
}

// mineIn is the caller's row on rs, or nil.
func mineIn(rs []schedule.RSVP, p auth.Principal, signedIn bool) *schedule.RSVP {
	if !signedIn {
		return nil
	}
	for i := range rs {
		r := &rs[i]
		if (p.UserID != uuid.Nil && r.UserID == p.UserID.String()) || (p.DiscordID != "" && r.DiscordID == p.DiscordID) {
			return r
		}
	}
	return nil
}

func msOrZero(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UnixMilli()
}

func seatState(s schedule.State) bool {
	return s == schedule.StateSeat || s == schedule.StateHeld || s == schedule.StateSeated
}

// viewOf builds a plan's view for viewer. names is false for GET
// /scheduled, which carries counts only.
func viewOf(c Config, p schedule.Plan, rs []schedule.RSVP, viewer auth.Principal, signedIn, names bool) scheduledView {
	v := scheduledView{
		ID: p.ID, Name: p.Name, Host: p.HostName, HostZone: p.Zone,
		StartsAt: p.StartsAt.UnixMilli(), EndsAt: p.EndsAt().UnixMilli(),
		DoorsOpenAt: p.DoorsOpenAt().UnixMilli(), HeldUntil: p.GraceEndsAt().UnixMilli(),
		LengthMinutes: p.LengthMinutes, Players: p.Players, Status: string(p.Status), Options: p.Options,
	}
	isHost := signedIn && p.HostUserID == viewer.UserID.String()
	mine := mineIn(rs, viewer, signedIn)
	seated, held, seats := 0, 0, 0
	seatPos, waitPos := 0, 0
	for _, r := range rs {
		switch r.State {
		case schedule.StateSeat:
			seats++
		case schedule.StateHeld:
			held++
		case schedule.StateSeated:
			seated++
		case schedule.StateWaiting:
			v.Waitlist++
		}
		if seatState(r.State) {
			seatPos++
		} else if r.State == schedule.StateWaiting {
			waitPos++
		}
		if mine != nil && r.DiscordID == mine.DiscordID {
			you := &scheduledYou{State: string(r.State), HeldUntil: msOrZero(r.HeldUntil)}
			switch {
			case seatState(r.State):
				you.Position = seatPos
			case r.State == schedule.StateWaiting:
				you.Position = waitPos
			}
			v.You = you
		}
	}
	switch p.Status {
	case schedule.StatusScheduled:
		v.OpenSeats = p.Players - seats
	case schedule.StatusOpen:
		taken := seated
		if id, err := uuid.Parse(p.GameID); err == nil {
			if t := c.Lobby.scheduleTable(id); t.Live {
				taken = t.Taken
			}
		}
		v.OpenSeats = p.Players - taken - held
	}
	v.OpenSeats = max(0, v.OpenSeats)
	onList := mine != nil && mine.OnList()
	if isHost {
		v.IsHost = true
		left := max(0, schedule.MaxMoves-p.Moves)
		v.MovesLeft = &left
	}
	if names && (isHost || onList) {
		v.GameID = p.GameID
		v.List = []scheduledPerson{}
		for _, r := range rs {
			if !r.OnList() && !isHost {
				continue
			}
			row := scheduledPerson{
				DisplayName: r.Name, State: string(r.State), Host: r.Source == schedule.SourceHost,
				You: mine != nil && r.DiscordID == mine.DiscordID, HeldUntil: msOrZero(r.HeldUntil),
			}
			if isHost {
				row.DiscordID = r.DiscordID
			}
			v.List = append(v.List, row)
		}
	}
	return v
}

// answerPlan reads plan id again and answers it for p.
func answerPlan(c Config, w http.ResponseWriter, r *http.Request, p auth.Principal, id string, status int) error {
	plan, rs, err := c.Schedule.Get(r.Context(), id)
	if err != nil {
		return scheduleError(c, w, err, "reading a scheduled table")
	}
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, status, viewOf(c, plan, rs, p, true, true))
}

// listScheduled is GET /scheduled.
func listScheduled(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := scheduleCaller(c, r)
	if err != nil {
		return err
	}
	all, err := c.Schedule.List(r.Context())
	if err != nil {
		return scheduleError(c, w, err, "listing scheduled tables")
	}
	out := scheduledListResponse{Plans: make([]scheduledView, 0, len(all)), MaxUpcomingPerHost: schedule.MaxUpcomingPerHost}
	for _, it := range all {
		out.Plans = append(out.Plans, viewOf(c, it.Plan, it.RSVPs, p, true, false))
	}
	// Your own plans and reservations first (§11), each part by start.
	sort.SliceStable(out.Plans, func(i, j int) bool {
		a := out.Plans[i].IsHost || (out.Plans[i].You != nil && out.Plans[i].You.State != string(schedule.StateRemoved))
		b := out.Plans[j].IsHost || (out.Plans[j].You != nil && out.Plans[j].You.State != string(schedule.StateRemoved))
		return a && !b
	})
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusOK, out)
}

// postScheduled is POST /scheduled. A plan is counted against the host's
// three upcoming plans and the server's fifty, and spends the create
// bucket of POST /games once every other check has passed (§12).
func postScheduled(limit *ratelimit.Limiter) lobbyHandler {
	return func(c Config, w http.ResponseWriter, r *http.Request) error {
		p, err := scheduleCaller(c, r)
		if err != nil {
			return err
		}
		var body scheduleBody
		if err := decodeJSON(w, r, &body); err != nil {
			return err
		}
		if body.StartsAt == 0 {
			return httpError(http.StatusBadRequest, errStartsAtRequired)
		}
		if strings.TrimSpace(body.Name) == "" {
			// A blank name is a surprise, as on POST /games (#2630).
			body.Name = SuggestTableName(p.Name)
		}
		key := callerKey(r)
		plan, err := c.Schedule.Create(r.Context(), c.Lobby.scheduleNow(), p.UserID.String(), schedule.Draft{
			Name: body.Name, StartsAt: time.UnixMilli(body.StartsAt), LengthMinutes: body.LengthMinutes,
			Zone: body.Zone, Players: body.Players,
			Options: schedule.Options{
				VoiceRoom: body.VoiceRoom, LFG: body.LFG, Results: body.Results, Planechase: body.Planechase,
				Setup: strings.TrimSpace(body.Setup), FillBots: body.FillBots,
			},
		}, func() bool { return limit.Allow(key) })
		if err != nil {
			return scheduleError(c, w, err, "scheduling a table")
		}
		c.logger().Info("table scheduled", "plan_id", plan.ID, "user_id", p.UserID.String(),
			"players", plan.Players, "starts_at", plan.StartsAt.Format(time.RFC3339))
		return answerPlan(c, w, r, p, plan.ID, http.StatusCreated)
	}
}

// getScheduled is GET /scheduled/{id}: public by id, so a member with no
// account can open the event's link. A session, when there is one, says
// who is asking.
func getScheduled(c Config, w http.ResponseWriter, r *http.Request) error {
	if c.Schedule == nil {
		return httpError(http.StatusServiceUnavailable, errScheduleOff)
	}
	plan, rs, err := c.Schedule.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		return scheduleError(c, w, err, "reading a scheduled table")
	}
	p, ok := optionalSession(c, r)
	signedIn := ok && isSignedInPerson(p)
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusOK, viewOf(c, plan, rs, p, signedIn, true))
}

// patchScheduled is PATCH /scheduled/{id}: the host's change before the
// doors open (§10). A new start is a move: at most three, at least 30
// minutes ahead, and it spends the create bucket.
func patchScheduled(limit *ratelimit.Limiter) lobbyHandler {
	return func(c Config, w http.ResponseWriter, r *http.Request) error {
		p, err := scheduleCaller(c, r)
		if err != nil {
			return err
		}
		var body schedulePatchBody
		if err := decodeJSON(w, r, &body); err != nil {
			return err
		}
		id := r.PathValue("id")
		patch := schedule.Patch{Name: body.Name, LengthMinutes: body.LengthMinutes, Zone: body.Zone, Players: body.Players}
		if body.StartsAt != nil {
			at := time.UnixMilli(*body.StartsAt)
			patch.StartsAt = &at
		}
		if body.options() {
			cur, _, gerr := c.Schedule.Get(r.Context(), id)
			if gerr != nil {
				return scheduleError(c, w, gerr, "reading a scheduled table")
			}
			o := cur.Options
			set := func(dst *bool, v *bool) {
				if v != nil {
					*dst = *v
				}
			}
			set(&o.VoiceRoom, body.VoiceRoom)
			set(&o.LFG, body.LFG)
			set(&o.Results, body.Results)
			set(&o.Planechase, body.Planechase)
			set(&o.FillBots, body.FillBots)
			if body.Setup != nil {
				o.Setup = strings.TrimSpace(*body.Setup)
			}
			patch.Options = &o
		}
		key := callerKey(r)
		plan, _, moved, err := c.Schedule.Update(r.Context(), c.Lobby.scheduleNow(), id, p.UserID.String(), patch,
			func() bool { return limit.Allow(key) })
		if err != nil {
			return scheduleError(c, w, err, "changing a scheduled table")
		}
		if moved {
			c.logger().Info("scheduled table moved", "plan_id", plan.ID, "moves", plan.Moves,
				"starts_at", plan.StartsAt.Format(time.RFC3339))
		}
		return answerPlan(c, w, r, p, plan.ID, http.StatusOK)
	}
}

// deleteScheduled is DELETE /scheduled/{id}: the host, or an admin, who
// can cancel any plan (§12); an admin's cancel of someone else's plan is
// a #mod-log line. After the doors open it is a table, closed as one.
func deleteScheduled(c Config, w http.ResponseWriter, r *http.Request) error {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	if c.Schedule == nil {
		return httpError(http.StatusServiceUnavailable, errScheduleOff)
	}
	admin := c.isAdmin(p)
	if !admin && !isSignedInPerson(p) {
		return httpError(http.StatusForbidden, errScheduleNeedAccount)
	}
	actor := ""
	if p.UserID != uuid.Nil {
		actor = p.UserID.String()
	}
	plan, _, err := c.Schedule.Cancel(r.Context(), c.Lobby.scheduleNow(), r.PathValue("id"), actor, admin)
	if err != nil {
		return scheduleError(c, w, err, "cancelling a scheduled table")
	}
	if plan.HostUserID != actor {
		logAdminAction(c.Log, "DELETE /scheduled/{id}", p, "plan_id", plan.ID)
		if c.ModLog != nil {
			c.ModLog.ScheduledCancelled(p, plan.Name)
		}
	}
	c.logger().Info("scheduled table cancelled", "plan_id", plan.ID)
	return answerPlan(c, w, r, p, plan.ID, http.StatusOK)
}

// putScheduledRSVP is PUT /scheduled/{id}/rsvp: Reserve a seat (§3). A
// person the host removed, or has blocked (ADR 0213 §3), is refused with
// the one answer that says neither.
func putScheduledRSVP(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := scheduleCaller(c, r)
	if err != nil {
		return err
	}
	id := r.PathValue("id")
	plan, _, err := c.Schedule.Get(r.Context(), id)
	if err != nil {
		return scheduleError(c, w, err, "reading a scheduled table")
	}
	me := p.UserID.String()
	if c.Social != nil && plan.HostUserID != me {
		blocked, berr := c.Social.IsBlocking(r.Context(), plan.HostUserID, me)
		if berr != nil {
			return scheduleError(c, w, berr, "checking a host's blocks")
		}
		if blocked {
			return httpError(http.StatusForbidden, errPlanRemoved)
		}
	}
	plan, _, err = c.Schedule.Reserve(r.Context(), c.Lobby.scheduleNow(), id, me)
	if err != nil {
		return scheduleError(c, w, err, "reserving a seat")
	}
	if plan.Status == schedule.StatusOpen {
		// A free seat at the open table can be held for them at once.
		c.Lobby.advancePlanNow(r.Context(), c, plan.ID)
	}
	return answerPlan(c, w, r, p, plan.ID, http.StatusOK)
}

// deleteScheduledRSVP is DELETE /scheduled/{id}/rsvp: Leave. After the
// doors open, a hold released this way passes down the waitlist at once.
func deleteScheduledRSVP(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := scheduleCaller(c, r)
	if err != nil {
		return err
	}
	plan, _, _, err := c.Schedule.Leave(r.Context(), c.Lobby.scheduleNow(), r.PathValue("id"), p.UserID.String())
	if err != nil {
		return scheduleError(c, w, err, "leaving a scheduled table")
	}
	if plan.Status == schedule.StatusOpen {
		c.Lobby.advancePlanNow(r.Context(), c, plan.ID)
	}
	return answerPlan(c, w, r, p, plan.ID, http.StatusOK)
}

// removeScheduledRSVP is DELETE /scheduled/{id}/rsvps/{discord_id}: the
// host removes a reservation from this occurrence (§3). That person
// cannot reserve it again.
func removeScheduledRSVP(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := scheduleCaller(c, r)
	if err != nil {
		return err
	}
	plan, _, _, err := c.Schedule.Remove(r.Context(), c.Lobby.scheduleNow(), r.PathValue("id"), p.UserID.String(),
		strings.TrimSpace(r.PathValue("discord_id")))
	if err != nil {
		return scheduleError(c, w, err, "removing a reservation")
	}
	if plan.Status == schedule.StatusOpen {
		c.Lobby.advancePlanNow(r.Context(), c, plan.ID)
	}
	return answerPlan(c, w, r, p, plan.ID, http.StatusOK)
}

// claimScheduledSeat is POST /scheduled/{id}/claim: Take your seat (§3).
// The seat held for the caller is theirs with no invite token; the
// answer is the seat's session, as POST /games/{id}/join's is.
func claimScheduledSeat(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := scheduleCaller(c, r)
	if err != nil {
		return err
	}
	plan, _, err := c.Schedule.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		return scheduleError(c, w, err, "reading a scheduled table")
	}
	switch plan.Status {
	case schedule.StatusScheduled:
		return httpError(http.StatusConflict, errNoDoorsYet)
	case schedule.StatusCancelled, schedule.StatusMissed:
		return httpError(http.StatusConflict, errPlanClosed)
	}
	gameID, err := uuid.Parse(plan.GameID)
	if err != nil {
		return httpError(http.StatusConflict, errPlanClosed)
	}
	// As a join: a newcomer is turned away at the connection cap before
	// a seat is taken that its socket would then be refused (ADR 0148).
	if c.Lobby.ConnectionsFull() {
		return capacity.ErrServerFull
	}
	identity := principalIdentity(p)
	name := ""
	if !identity.Populated() {
		name = p.Name
	}
	meta, playerID, err := c.Lobby.ClaimHeldSeat(gameID, name, identity, p.UserID)
	if err != nil {
		return err
	}
	c.Lobby.advancePlanNow(context.WithoutCancel(r.Context()), c, plan.ID)
	sp := withIdentity(auth.Principal{
		Role:     auth.RolePlayer,
		UserID:   p.UserID,
		GameID:   meta.ID,
		PlayerID: playerID,
		Name:     name,
	}, identity)
	tok, issued, err := issueFor(r.Context(), c, sp, &p)
	if err != nil {
		return err
	}
	setSessionCookie(c, w, tok, issued.ExpiresAt)
	meta.InviteToken = ""
	meta.SpectatorInvite = ""
	return writeJSON(w, http.StatusOK, sessionResponse{
		Token:     tok,
		ExpiresAt: issued.ExpiresAt,
		Principal: issued,
		Game:      &meta,
		PlayerID:  playerID,
	})
}
