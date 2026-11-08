package lobby

// schedule.go: scheduled tables in the lobby (ADR 0215 §1, §3, §6, §7;
// slice 2 of #3321). The plans and their lists are internal/schedule's;
// this file is what a plan does to a table.
//
//   - RunSchedule ticks every ScheduleTickEvery. Each tick sweeps the
//     plans past their keep window, opens the plans whose doors are due,
//     runs the holds of the open ones forward, and keeps every plan's
//     table in step with its plan.
//   - Doors-open makes an ordinary lobby table through CreateWith: the
//     host is its creator and its named host (HostDiscordID), so the
//     open-table cap and #lfg read it as theirs, and nobody is seated.
//     The create skips the open-table cap and the create rate limit (the
//     plan was counted when it was made, §12) but not the load cap: a
//     full server is tried again every tick until the grace ends, and
//     then the plan is missed. A plan more than an hour past its start
//     (after a restart) is missed instead of opened.
//   - THE HOLDS ARE IN MEMORY ON THE TABLE (gameEntry.sched), copied from
//     the plan's list after every change, and the list is the authority:
//     a restart loses nothing, because the first tick copies them again.
//     A held seat counts as taken for everyone but its holder: a join by
//     invite link, a bot, the #lfg post's seat count and the start
//     (seatFreeLocked, lfgTable, Start). The plan's number of players is
//     the table's seat count while the plan is kept.
//   - Taking a held seat is a join that the hold authorises instead of an
//     invite (ClaimHeldSeat, POST /scheduled/{id}/claim), or a join by the
//     invite link, which a hold also lets through. The next advance marks
//     the person seated.
//
// LOCKS. l.sched.op serialises the scheduler with the schedule routes'
// table work, so an advance never interleaves with a claim or a leave.
// It is taken before l.mu and never while holding it; the schedule store
// has its own mutex inside each call.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/schedule"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/tablesetups"
)

// ScheduleTickEvery is how often RunSchedule looks at the plans. Doors
// open, holds lapse and seats pass down the waitlist at most this late.
const ScheduleTickEvery = time.Minute

// scheduleTickTimeout bounds one tick's database work.
const scheduleTickTimeout = 30 * time.Second

// The `code` of the two refusals a hold causes.
const (
	// CodeSeatsReserved: every free seat is held for someone else.
	CodeSeatsReserved = "seats_reserved"
	// CodeSeatsHeld: the start would drop holds, and was not confirmed.
	CodeSeatsHeld = "seats_held"
)

// ErrNoHeldSeat is a claim by someone who holds no seat at the table.
var ErrNoHeldSeat = errors.New("lobby: you don't hold a seat at this table")

// SeatsReservedError refuses a seat when every free one is held for
// somebody else (ADR 0215 §3): "The other seats are reserved until
// 19:40". Until is when the last hold runs out; the HTTP answer carries
// it as reserved_until, for the client to show in the viewer's zone.
type SeatsReservedError struct {
	Until time.Time
}

func (e *SeatsReservedError) Error() string {
	return "The other seats are reserved until " + e.Until.UTC().Format("15:04") + " UTC"
}

// SeatsHeldError refuses a start while reserved seats are still empty,
// until the host confirms it (§3): "2 reserved seats are still empty.
// Start without them?"
type SeatsHeldError struct {
	Held int
}

func (e *SeatsHeldError) Error() string {
	if e.Held == 1 {
		return "1 reserved seat is still empty. Start without it?"
	}
	return fmt.Sprintf("%d reserved seats are still empty. Start without them?", e.Held)
}

// scheduleState is the scheduler's lock and clock.
type scheduleState struct {
	op  sync.Mutex
	mu  sync.Mutex
	now func() time.Time
}

// SetScheduleClock replaces the schedule clock. Tests only; nil is
// time.Now.
func (l *Lobby) SetScheduleClock(now func() time.Time) {
	l.sched.mu.Lock()
	defer l.sched.mu.Unlock()
	l.sched.now = now
}

func (l *Lobby) scheduleNow() time.Time {
	l.sched.mu.Lock()
	defer l.sched.mu.Unlock()
	if l.sched.now != nil {
		return l.sched.now().UTC()
	}
	return time.Now().UTC()
}

// entrySchedule is a scheduled table's plan, as its table keeps it.
type entrySchedule struct {
	planID  string
	start   time.Time
	players int
	holds   []seatHold
}

// seatHold is one seat held for a person until until.
type seatHold struct {
	userID, discordID string
	until             time.Time
}

func (h seatHold) of(userID uuid.UUID, discordID string) bool {
	return (userID != uuid.Nil && h.userID == userID.String()) || (discordID != "" && h.discordID == discordID)
}

// seatCapLocked is how many seats e has: the plan's players while a
// plan is kept, game.MaxPlayers otherwise. Callers hold l.mu.
func seatCapLocked(e *gameEntry) int {
	if e.sched != nil && e.sched.players > 0 && e.sched.players < game.MaxPlayers {
		return e.sched.players
	}
	return game.MaxPlayers
}

// heldLocked is how many seats are held at e. Callers hold l.mu.
func heldLocked(e *gameEntry) int {
	if e.sched == nil {
		return 0
	}
	return len(e.sched.holds)
}

// holdIndexLocked is the hold of the person (userID, discordID) at e,
// or -1. Callers hold l.mu.
func holdIndexLocked(e *gameEntry, userID uuid.UUID, discordID string) int {
	if e.sched == nil {
		return -1
	}
	for i, h := range e.sched.holds {
		if h.of(userID, discordID) {
			return i
		}
	}
	return -1
}

// seatFreeLocked is the seat check of a join and a bot (§3): whether e
// has a seat for this person, and the index of their own hold, -1 for
// none. A bot or an agent passes uuid.Nil and "". Callers hold l.mu.
func seatFreeLocked(e *gameEntry, userID uuid.UUID, discordID string) (int, error) {
	taken, cap := len(e.meta.Players), seatCapLocked(e)
	hold := holdIndexLocked(e, userID, discordID)
	if hold >= 0 {
		if taken >= cap {
			return hold, ErrGameFull
		}
		return hold, nil
	}
	held := heldLocked(e)
	if taken+held < cap {
		return -1, nil
	}
	if held > 0 && taken < cap {
		until := time.Time{}
		for _, h := range e.sched.holds {
			if h.until.After(until) {
				until = h.until
			}
		}
		return -1, &SeatsReservedError{Until: until}
	}
	return -1, ErrGameFull
}

// takeHoldLocked drops hold i at e, once its person sat down. Callers
// hold l.mu.
func takeHoldLocked(e *gameEntry, i int) {
	if e.sched == nil || i < 0 || i >= len(e.sched.holds) {
		return
	}
	e.sched.holds = append(e.sched.holds[:i:i], e.sched.holds[i+1:]...)
}

// seatedLocked reports whether the person (userID, discordID) holds a
// seat at e. Callers hold l.mu.
func seatedLocked(e *gameEntry, userID, discordID string) bool {
	for _, s := range e.meta.Players {
		if s.IsBot || s.IsAgent {
			continue
		}
		if (userID != "" && s.UserID == userID) || (discordID != "" && s.DiscordID == discordID) {
			return true
		}
	}
	return false
}

// personSeatedLocked reports whether any person, not a bot or an agent,
// sits at e. Callers hold l.mu.
func personSeatedLocked(e *gameEntry) bool {
	for _, s := range e.meta.Players {
		if !s.IsBot && !s.IsAgent {
			return true
		}
	}
	return false
}

// scheduleTable is what the plan's advance needs to know about table id.
func (l *Lobby) scheduleTable(id uuid.UUID) schedule.Table {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.games[id]
	if !ok || e.room == nil || e.practice {
		return schedule.Table{}
	}
	t := schedule.Table{
		Live:    !e.meta.Archived() && e.room.Game.CurrentState() == game.StateLobby,
		Taken:   len(e.meta.Players),
		Users:   map[string]bool{},
		Discord: map[string]bool{},
	}
	for _, s := range e.meta.Players {
		if s.IsBot || s.IsAgent {
			continue
		}
		if s.UserID != "" {
			t.Users[s.UserID] = true
		}
		if s.DiscordID != "" {
			t.Discord[s.DiscordID] = true
		}
	}
	return t
}

// attachSchedule copies plan p and the holds on its list onto its table
// id. A hold whose person already sat down is left off, so a claim that
// raced the advance is not held twice.
func (l *Lobby) attachSchedule(id uuid.UUID, p schedule.Plan, rs []schedule.RSVP) {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.games[id]
	if !ok || e.practice {
		return
	}
	next := &entrySchedule{planID: p.ID, start: p.StartsAt, players: p.Players}
	if p.Status == schedule.StatusOpen {
		for _, r := range rs {
			if r.State == schedule.StateHeld && !seatedLocked(e, r.UserID, r.DiscordID) {
				next.holds = append(next.holds, seatHold{userID: r.UserID, discordID: r.DiscordID, until: r.HeldUntil})
			}
		}
	}
	if !sameSchedule(e.sched, next) {
		e.sched = next
		// ADR 0215 §6: the #lfg post counts the holds as taken.
		l.lfgChangedLocked(e)
	}
}

func sameSchedule(a, b *entrySchedule) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.planID != b.planID || !a.start.Equal(b.start) || a.players != b.players || len(a.holds) != len(b.holds) {
		return false
	}
	for i := range a.holds {
		if a.holds[i].userID != b.holds[i].userID || a.holds[i].discordID != b.holds[i].discordID ||
			!a.holds[i].until.Equal(b.holds[i].until) {
			return false
		}
	}
	return true
}

// detachSchedulesExcept lets go of the plan on every table not in keep:
// its plan was swept, deleted with its host's account, or lost its table.
func (l *Lobby) detachSchedulesExcept(keep map[uuid.UUID]bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for id, e := range l.games {
		if e.sched != nil && !keep[id] {
			e.sched = nil
			l.lfgChangedLocked(e)
		}
	}
}

// forgetScheduleOf is account deletion's in-memory half (ADR 0215,
// privacy section): the deleted person's holds go from every table, and
// the tables of the plans they hosted are plans no more. The rows went
// in the deletion's transaction.
func (l *Lobby) forgetScheduleOf(userID uuid.UUID, discordID string, games []string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	hosted := map[string]bool{}
	for _, g := range games {
		hosted[g] = true
	}
	for id, e := range l.games {
		if e.sched == nil {
			continue
		}
		if hosted[id.String()] {
			e.sched = nil
			l.lfgChangedLocked(e)
			continue
		}
		if i := holdIndexLocked(e, userID, discordID); i >= 0 {
			takeHoldLocked(e, i)
			l.lfgChangedLocked(e)
		}
	}
}

// ClaimHeldSeat is "Take your seat" (ADR 0215 §3): the person sits down
// in the seat held for them, with no invite token. The hold is the
// authority, keyed by account or Discord id. ErrNoHeldSeat when they
// hold none.
func (l *Lobby) ClaimHeldSeat(id uuid.UUID, playerName string, identity DiscordIdentity, userID uuid.UUID) (GameMeta, uuid.UUID, error) {
	return l.join(id, "", playerName, identity, userID, nil, true)
}

// RunSchedule ticks the scheduler every ScheduleTickEvery until ctx
// ends, once at once: started at boot after RestoreFromDisk, so the
// first tick puts the holds back on the restored tables. c is the
// lobby's Config, for the bots and the last setup; c.Schedule nil (no
// database) runs nothing.
func (l *Lobby) RunSchedule(ctx context.Context, c Config) {
	if c.Schedule == nil {
		return
	}
	l.TickSchedule(c)
	t := time.NewTicker(ScheduleTickEvery)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			l.TickSchedule(c)
		}
	}
}

// TickSchedule runs the scheduler once (ADR 0215 §1, §3): the sweep, the
// doors-opens that are due, every open plan's holds, and every plan's
// table kept in step. Exported for tests, which drive it with
// SetScheduleClock as TickIdle is driven.
func (l *Lobby) TickSchedule(c Config) {
	if c.Schedule == nil {
		return
	}
	log := c.logger()
	now := l.scheduleNow()
	l.sched.op.Lock()
	defer l.sched.op.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), scheduleTickTimeout)
	defer cancel()

	if n, err := c.Schedule.Sweep(ctx, now); err != nil {
		log.Error("sweeping old scheduled tables failed", "err", err)
	} else if n > 0 {
		log.Info("swept old scheduled tables", "plans", n)
	}
	plans, err := c.Schedule.Active(ctx, now)
	if err != nil {
		// Leave every table as it is: the next tick tries again.
		log.Error("reading scheduled tables failed", "err", err)
		return
	}
	keep := map[uuid.UUID]bool{}
	for _, it := range plans {
		p := it.Plan
		switch p.Status {
		case schedule.StatusScheduled:
			if id := l.openPlanOp(ctx, c, p, now); id != uuid.Nil {
				keep[id] = true
			}
		case schedule.StatusOpen:
			if id, perr := uuid.Parse(p.GameID); perr == nil {
				keep[id] = true
				l.advancePlanOp(ctx, c, p, now)
			}
		case schedule.StatusSettled:
			if id, perr := uuid.Parse(p.GameID); perr == nil {
				keep[id] = true
				l.attachSchedule(id, p, it.RSVPs)
			}
		}
	}
	l.detachSchedulesExcept(keep)
}

// advancePlanNow runs plan id's holds forward at once, after a route
// changed its list or its table: a leave, a removal, a reservation or a
// claim on an open plan.
func (l *Lobby) advancePlanNow(ctx context.Context, c Config, id string) {
	if c.Schedule == nil {
		return
	}
	l.sched.op.Lock()
	defer l.sched.op.Unlock()
	p, _, err := c.Schedule.Get(ctx, id)
	if err != nil || p.Status != schedule.StatusOpen {
		return
	}
	l.advancePlanOp(ctx, c, p, l.scheduleNow())
}

// openPlanOp opens plan p's doors at now (§1), or marks it missed. It
// answers the new table, uuid.Nil when none opened. Callers hold
// l.sched.op.
func (l *Lobby) openPlanOp(ctx context.Context, c Config, p schedule.Plan, now time.Time) uuid.UUID {
	log := c.logger()
	if !now.Before(p.StartsAt.Add(schedule.MissedAfter)) {
		l.missPlanOp(ctx, c, p, now, "too late to open")
		return uuid.Nil
	}
	host, err := uuid.Parse(p.HostUserID)
	if err != nil {
		l.missPlanOp(ctx, c, p, now, "no host")
		return uuid.Nil
	}
	meta, err := l.CreateWith(p.Name, host, p.HostDiscordID)
	if err != nil {
		if !now.Before(p.GraceEndsAt()) {
			l.missPlanOp(ctx, c, p, now, err.Error())
		} else {
			log.Warn("scheduled table could not open yet; trying again", "plan_id", p.ID, "err", err)
		}
		return uuid.Nil
	}
	opened, rs, err := c.Schedule.Open(ctx, now, p.ID, meta.ID.String())
	if err != nil {
		// The plan stays scheduled, so the next tick tries again; the
		// table made for it would be an orphan.
		log.Error("recording a scheduled table's doors-open failed", "plan_id", p.ID, "err", err)
		_ = l.Delete(meta.ID)
		return uuid.Nil
	}
	metrics.GameCreated()
	l.attachSchedule(meta.ID, opened, rs)
	applyPlanOptions(ctx, c, opened, meta.ID)
	held := 0
	for _, r := range rs {
		if r.State == schedule.StateHeld {
			held++
		}
	}
	log.Info("scheduled table opened", "plan_id", p.ID, "game_id", meta.ID.String(), "holds", held)
	return meta.ID
}

func (l *Lobby) missPlanOp(ctx context.Context, c Config, p schedule.Plan, now time.Time, why string) {
	if _, _, err := c.Schedule.Miss(ctx, now, p.ID); err != nil {
		c.logger().Error("marking a scheduled table missed failed", "plan_id", p.ID, "err", err)
		return
	}
	c.logger().Warn("scheduled table missed", "plan_id", p.ID, "why", why)
}

// advancePlanOp runs open plan p's holds forward to now against its
// table, copies the result onto the table, and fills the empty seats
// with bots when the plan settles and the host asked for them. Callers
// hold l.sched.op.
func (l *Lobby) advancePlanOp(ctx context.Context, c Config, p schedule.Plan, now time.Time) {
	id, err := uuid.Parse(p.GameID)
	if err != nil {
		return
	}
	after, rs, out, err := c.Schedule.Advance(ctx, now, p.ID, l.scheduleTable(id))
	if err != nil {
		c.logger().Error("running a scheduled table's holds failed", "plan_id", p.ID, "err", err)
		return
	}
	l.attachSchedule(id, after, rs)
	if out.Settled {
		filled := 0
		if out.FillBots > 0 {
			filled = fillScheduledBots(ctx, c, after, id, out.FillBots)
		}
		c.logger().Info("scheduled table's holds ended", "plan_id", p.ID, "game_id", id.String(),
			"seated", len(out.Seated), "lapsed", len(out.Lapsed), "bots", filled)
	}
}

// applyPlanOptions applies a plan's create options to its new table, as
// POST /games applies them (§1): Planechase, the last setup's settings,
// the voice room, #lfg and results. The last setup's bots are not
// seated: the seats are held. A server without a feature ignores the
// ask, as the create form's does.
func applyPlanOptions(ctx context.Context, c Config, p schedule.Plan, id uuid.UUID) {
	log := c.logger()
	o := p.Options
	if o.Planechase {
		on := true
		if _, err := c.Lobby.UpdateSettings(id, uuid.Nil, game.SettingsPatch{Planechase: &on}); err != nil {
			log.Warn("planechase on a scheduled table failed", "game_id", id.String(), "err", err)
		}
	}
	if o.Setup == schedule.SetupLast {
		if host, err := uuid.Parse(p.HostUserID); err == nil {
			if rec, gerr := c.tableSetups().Get(ctx, host); gerr == nil {
				if _, skip := applySetupSettings(c, uuid.Nil, id, rec.Setup.Settings); skip != nil {
					log.Warn("a scheduled table's last setup was skipped", "game_id", id.String(), "reason", skip.Reason)
				}
			}
		}
	}
	if o.VoiceRoom {
		if _, err := c.Lobby.SetVoiceRoom(id, true); err != nil && !errors.Is(err, ErrVoiceRoomsOff) {
			log.Warn("voice room on a scheduled table failed", "game_id", id.String(), "err", err)
		}
	}
	if o.LFG {
		if _, err := c.Lobby.SetLFG(id, true); err != nil && !errors.Is(err, ErrLFGOff) {
			log.Warn("lfg post for a scheduled table failed", "game_id", id.String(), "err", err)
		}
	}
	if o.Results {
		if _, err := c.Lobby.SetResults(id, true); err != nil && !errors.Is(err, ErrResultsOff) {
			log.Warn("results opt-in on a scheduled table failed", "game_id", id.String(), "err", err)
		}
	}
}

// fillScheduledBots seats n bots at a settled plan's table (§3): the
// host's last setup's bots first, then the default tier with the
// catalog's decks in turn. It answers how many sat down.
func fillScheduledBots(ctx context.Context, c Config, p schedule.Plan, id uuid.UUID, n int) int {
	var saved []tablesetups.Bot
	if host, err := uuid.Parse(p.HostUserID); err == nil {
		if rec, gerr := c.tableSetups().Get(ctx, host); gerr == nil {
			saved = rec.Setup.Bots
		}
	}
	added := 0
	for i := 0; i < n; i++ {
		b := defaultScheduleBot(c, i)
		if i < len(saved) {
			b = saved[i]
		}
		name := strings.TrimSpace(b.Name)
		if name == "" {
			name = fmt.Sprintf("Bot %d", i+1)
		}
		if reason := seatSetupBot(ctx, c, id, name, b); reason != "" {
			c.logger().Warn("a bot for a scheduled table was skipped", "game_id", id.String(), "reason", reason)
			continue
		}
		added++
	}
	return added
}

// defaultScheduleBot is the i'th bot of the default kind: the heuristic
// tier where it is offered (else the first tier), on the catalog's
// decks in turn.
func defaultScheduleBot(c Config, i int) tablesetups.Bot {
	var b tablesetups.Bot
	if c.Bots != nil {
		tiers := c.Bots.Tiers()
		for _, t := range tiers {
			if t == "heuristic" {
				b.Tier = t
			}
		}
		if b.Tier == "" && len(tiers) > 0 {
			b.Tier = tiers[0]
		}
	}
	if c.BotDecks != nil {
		if decks := c.BotDecks.List(); len(decks) > 0 {
			b.DeckID = decks[i%len(decks)].ID
		}
	}
	return b
}
