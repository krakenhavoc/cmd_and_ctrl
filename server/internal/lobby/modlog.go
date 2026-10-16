package lobby

// modlog.go is ADR 0156's 2026-10-16 amendment (#3325): the mod log.
// Each in-app player report and each admin action that acts on a person
// or a table (ending, restoring or deleting a table, removing a seat or a
// spectator, revoking an account's sessions) is posted to a private
// #mod-log channel on the community server, so the Discord's moderators
// see in one place what was reported and what was done.
//
// What an entry carries is what an admin already reads in the admin views,
// and less: the table's name, the name a seat or spectator had there, a
// report's reason, an account's display name and a link into the admin
// view. Never a Discord id, an account id, a chat excerpt or who filed a
// report (the Reports view keeps those). Nothing new is stored here: an
// entry is sent and forgotten. Discord keeps the message until a
// moderator deletes it, which the privacy page says.
//
// Posting never blocks and never fails the request that caused it: a
// background goroutine per entry, at most modLogMaxInFlight at once, a
// retry while Discord is rate-limiting or not answering, then a WARN.

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
)

const (
	// modLogMaxAttempts is how many transient failures in a row an entry
	// survives.
	modLogMaxAttempts = 3
	// modLogCallTimeout bounds one Discord call.
	modLogCallTimeout = 10 * time.Second
	// modLogMaxInFlight bounds the entries being sent at once. Reports
	// are rate limited per reporter and admin actions are admin only;
	// this is the second wall, so a burst of slow Discord calls cannot
	// pile up goroutines.
	modLogMaxInFlight = 8
)

// modLogAPI is the Discord half the mod log uses: discord.ModLog, or a
// fake in tests.
type modLogAPI interface {
	Post(ctx context.Context, client *http.Client, e discord.ModLogEntry) error
}

// ModLog posts entries to the mod log channel. Construct it with
// NewModLog and put it in Config.ModLog. A nil *ModLog is the feature
// off: every method is a no-op on it.
type ModLog struct {
	api    modLogAPI
	client *http.Client
	log    *slog.Logger
	// base is the site's origin, for the links into the admin view; ""
	// leaves the links out.
	base string
	// wait sleeps between attempts and now stamps an entry; a test
	// replaces them.
	wait func(time.Duration)
	now  func() time.Time

	slots chan struct{}
	wg    sync.WaitGroup
}

// NewModLog builds the mod log over the configured channel. base is the
// public origin (CMDCTRL_PUBLIC_BASE_URL); "" posts entries with no link.
func NewModLog(m discord.ModLog, base string, log *slog.Logger) *ModLog {
	return newModLog(m, base, &http.Client{Timeout: modLogCallTimeout}, log)
}

func newModLog(api modLogAPI, base string, client *http.Client, log *slog.Logger) *ModLog {
	if log == nil {
		log = slog.Default()
	}
	return &ModLog{
		api: api, client: client, log: log,
		base:  strings.TrimRight(base, "/"),
		wait:  time.Sleep,
		now:   time.Now,
		slots: make(chan struct{}, modLogMaxInFlight),
	}
}

// Wait blocks until every entry being sent has finished. A test's handle
// on the background work.
func (m *ModLog) Wait() {
	if m != nil {
		m.wg.Wait()
	}
}

// post sends one entry in the background.
func (m *ModLog) post(e discord.ModLogEntry) {
	if m == nil {
		return
	}
	if e.At.IsZero() {
		e.At = m.now()
	}
	select {
	case m.slots <- struct{}{}:
	default:
		m.log.Warn("mod log: too many entries being sent; this one is dropped", "title", e.Title)
		return
	}
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		defer func() { <-m.slots }()
		m.send(e)
	}()
}

func (m *ModLog) send(e discord.ModLogEntry) {
	for attempt := 1; ; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), modLogCallTimeout)
		err := m.api.Post(ctx, m.client, e)
		cancel()
		if err == nil {
			return
		}
		if !discord.Transient(err) || attempt >= modLogMaxAttempts {
			m.log.Warn("mod log: entry not posted", "title", e.Title, "attempt", attempt, "err", err)
			return
		}
		m.wait(retryWait(err))
	}
}

// link is an admin view's address, or "" with no origin configured.
func (m *ModLog) link(hash string) string {
	if m.base == "" {
		return ""
	}
	return m.base + "/" + hash
}

// --- what is logged ---------------------------------------------------

// Report logs a filed player report: the table, the reported seat's name
// and the reason. The reporter is not named, here or anywhere outside the
// admin Reports view, and the chat excerpt stays there too.
func (m *ModLog) Report(rec ReportRecord) {
	if m == nil {
		return
	}
	m.post(discord.ModLogEntry{
		Title: "Player report",
		Link:  m.link("#/admin/reports"),
		Fields: []discord.ModLogField{
			{Name: "Table", Value: rec.GameName},
			{Name: "Reported", Value: rec.ReportedName},
			{Name: "Reason", Value: rec.Reason},
		},
		Footer: "Filed from a table. Who filed it, and the chat, are in Admin → Reports.",
		At:     rec.CreatedAt,
	})
}

// TableArchived logs an admin ending (archiving) a table, or restoring
// one.
func (m *ModLog) TableArchived(p auth.Principal, id uuid.UUID, name string, archived bool) {
	if m == nil {
		return
	}
	title := "Table ended"
	if !archived {
		title = "Table restored"
	}
	m.post(discord.ModLogEntry{
		Title:  title,
		Link:   m.link("#/admin/games/" + id.String()),
		Fields: []discord.ModLogField{{Name: "Table", Value: name}},
		Actor:  modLogActor(p),
	})
}

// TableDeleted logs an admin deleting a table for good.
func (m *ModLog) TableDeleted(p auth.Principal, name string) {
	if m == nil {
		return
	}
	m.post(discord.ModLogEntry{
		Title:  "Table deleted",
		Fields: []discord.ModLogField{{Name: "Table", Value: name}},
		Actor:  modLogActor(p),
	})
}

// SeatRemoved logs an admin removing a seat, or a spectator, from a
// table. who is the name the seat or spectator had there.
func (m *ModLog) SeatRemoved(p auth.Principal, id uuid.UUID, table, who string, spectator, blocked bool) {
	if m == nil {
		return
	}
	title, label := "Seat removed", "Seat"
	if spectator {
		title, label = "Spectator removed", "Spectator"
	}
	blockedText := "no"
	if blocked {
		blockedText = "yes, from this table"
	}
	m.post(discord.ModLogEntry{
		Title: title,
		Link:  m.link("#/admin/games/" + id.String()),
		Fields: []discord.ModLogField{
			{Name: "Table", Value: table},
			{Name: label, Value: who},
			{Name: "Blocked", Value: blockedText},
		},
		Actor: modLogActor(p),
	})
}

// SessionsRevoked logs an admin signing an account out everywhere. name
// is its display name, "" when it could not be read.
func (m *ModLog) SessionsRevoked(p auth.Principal, userID uuid.UUID, name string, sockets int) {
	if m == nil {
		return
	}
	if name == "" {
		name = "an account"
	}
	m.post(discord.ModLogEntry{
		Title: "Sessions revoked",
		Link:  m.link("#/admin/accounts/" + userID.String()),
		Fields: []discord.ModLogField{
			{Name: "Account", Value: name},
			{Name: "Connections closed", Value: strconv.Itoa(sockets)},
		},
		Actor: modLogActor(p),
	})
}

// modLogActor names the admin behind p for an entry's footer: an
// allowlisted person by their display name, the shared token by what it
// is. The token is what the Discord bot's /c2-end uses, so an entry made
// through the bot says so.
func modLogActor(p auth.Principal) string {
	if p.UserID != uuid.Nil {
		if n := strings.TrimSpace(p.Name); n != "" {
			return n + " (admin)"
		}
		return "an admin"
	}
	return "the admin token (the Admin page, or the Discord bot's /c2-end)"
}
