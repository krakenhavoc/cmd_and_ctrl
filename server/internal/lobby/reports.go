package lobby

// reports.go: "report player" (#2984, ADR 0075 section 2.6).
//
// A seated player files a report against another seat from the table:
// who, why in a sentence, and the last few chat lines as the table saw
// them. It goes to the admin Reports list and nowhere else; nothing is
// automatically done to the reported player, and the reporter is not
// told what happened. The report is a tip to the operator, who decides.
// The operator is told one is waiting by the cmdctrl_player_reports_total
// counter and the PlayerReportFiled alert on it (ADR 0123 section 7,
// #3113), which carries a count and nothing else: no name, no reason.
//
//	POST   /games/{id}/reports    a seat at that table; rate limited per reporter
//	GET    /admin/reports         admin only
//	DELETE /admin/reports/{id}    admin only: dismiss
//
// What a report stores, and for how long, is on the privacy page: the
// two seats' names and account ids, the reason, and up to
// reportExcerptLines recent chat lines (everyone's, with the seat names
// they were written under). It is removed ReportRetention after it is
// filed, the list is capped at ReportMaxRows (oldest first), and account
// deletion removes the reports a person filed or is the subject of.
//
// Not reportable: a bot (nobody typed its name or its lines) and the
// reporter's own seat. A SPECTATOR is reportable since #3113 (spectators.go):
// the request names "spectator_id" in place of "player_id", the report
// goes through the same limit into the same list, and the admin reads the
// reported line as "<name> (spectator)". The spectator must have a socket
// open at the table when the report is filed.

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

const (
	// ReportRetention is how long a report is kept. The privacy notice
	// states it (privacy.ts reportDays; privacy_facts_test.go holds it).
	ReportRetention = 90 * 24 * time.Hour
	// ReportMaxRows caps the stored reports. Past it the oldest go, so a
	// flood of reports cannot grow the table without bound.
	ReportMaxRows = 500
	// MaxReportReasonLen is the longest reason, in runes.
	MaxReportReasonLen = 300
	// reportExcerptLines is how many recent chat lines a report carries.
	reportExcerptLines = 10
	// reportBurst and reportEvery are the per-reporter rate limit: a
	// burst of reportBurst, then one more every reportEvery. The privacy
	// notice does not state them; the report form does, in its refusal.
	reportBurst = 3
	reportEvery = 5 * time.Minute
	// ReportSweepInterval is how often RunReportSweeper prunes.
	ReportSweepInterval = 6 * time.Hour
)

var (
	// ErrReportSelf: you cannot report your own seat.
	ErrReportSelf = errors.New("you can't report your own seat")
	// ErrReportTarget: the reported seat is not a person at this table.
	ErrReportTarget = errors.New("that player isn't a person at this table")
	// ErrReportReasonEmpty: a report needs a reason.
	ErrReportReasonEmpty = errors.New("say in a few words what the problem is")
)

// ReportInput is what FileReport needs, all of it already authenticated.
type ReportInput struct {
	GameID           uuid.UUID
	ReporterPlayerID uuid.UUID
	ReportedPlayerID uuid.UUID
	// ReportedSpectator, when set, is the reported party in place of
	// ReportedPlayerID: a spectator the hub reports as watching (#3113).
	ReportedSpectator *ws.LiveSpectator
	Reason            string
}

// FileReport stores a report and returns it. The reporter must hold a
// seat at the table; the reported seat must be a human at it.
func (l *Lobby) FileReport(in ReportInput) (ReportRecord, error) {
	reason := strings.TrimSpace(in.Reason)
	if reason == "" {
		return ReportRecord{}, ErrReportReasonEmpty
	}
	if r := []rune(reason); len(r) > MaxReportReasonLen {
		reason = string(r[:MaxReportReasonLen])
	}
	if in.ReportedSpectator == nil && in.ReporterPlayerID == in.ReportedPlayerID {
		return ReportRecord{}, ErrReportSelf
	}

	l.mu.Lock()
	entry, ok := l.games[in.GameID]
	if !ok {
		l.mu.Unlock()
		return ReportRecord{}, ErrGameNotFound
	}
	reporter, rok := findSeat(entry.meta.Players, in.ReporterPlayerID)
	var (
		reported SeatInfo
		tok      bool
	)
	if in.ReportedSpectator == nil {
		reported, tok = findSeat(entry.meta.Players, in.ReportedPlayerID)
	}
	name, room := entry.meta.Name, entry.room
	removed := rok && room.Game.SeatRemoved(reporter.PlayerID)
	l.mu.Unlock()

	if !rok || removed {
		return ReportRecord{}, ErrPlayerNotInGame
	}
	if in.ReportedSpectator == nil && (!tok || reported.IsBot) {
		return ReportRecord{}, ErrReportTarget
	}

	var excerpt []ReportLine
	for _, c := range room.RecentChat(reportExcerptLines) {
		excerpt = append(excerpt, ReportLine{Author: c.AuthorName, Text: c.Text, At: c.Timestamp})
	}
	rec := ReportRecord{
		ID:               uuid.New(),
		CreatedAt:        time.Now().UTC().Truncate(time.Millisecond),
		GameID:           in.GameID,
		GameName:         name,
		ReporterPlayerID: reporter.PlayerID,
		ReporterUserID:   reporter.UserID,
		ReporterName:     reporter.Name,
		ReportedPlayerID: reported.PlayerID,
		ReportedUserID:   reported.UserID,
		ReportedName:     reported.Name,
		Reason:           reason,
		Excerpt:          excerpt,
	}
	if in.ReportedSpectator != nil {
		rec.ReportedPlayerID, rec.ReportedUserID, rec.ReportedName = reportedSpectator(*in.ReportedSpectator)
	}
	ctx, cancel := storeCtx()
	defer cancel()
	if err := l.store.AddReport(ctx, rec); err != nil {
		return ReportRecord{}, err
	}
	// Counted once the row is stored, so the operator's alert (ADR 0123
	// section 7) means a report an admin can open, never a rejected one.
	metrics.PlayerReported()
	l.pruneReports(ctx, rec.CreatedAt)
	return rec, nil
}

// Reports lists the stored reports, newest first, after pruning.
func (l *Lobby) Reports() ([]ReportRecord, error) {
	ctx, cancel := storeCtx()
	defer cancel()
	l.pruneReports(ctx, time.Now().UTC())
	return l.store.Reports(ctx)
}

// DismissReport removes one report. ErrStoreNotFound if there is none.
func (l *Lobby) DismissReport(id uuid.UUID) error {
	ctx, cancel := storeCtx()
	defer cancel()
	return l.store.DeleteReport(ctx, id)
}

// PruneReports removes reports past ReportRetention and past
// ReportMaxRows, and returns how many it removed.
func (l *Lobby) PruneReports(now time.Time) (int, error) {
	ctx, cancel := storeCtx()
	defer cancel()
	return l.store.PruneReports(ctx, now.Add(-ReportRetention), ReportMaxRows)
}

func (l *Lobby) pruneReports(ctx context.Context, now time.Time) {
	if _, err := l.store.PruneReports(ctx, now.Add(-ReportRetention), ReportMaxRows); err != nil {
		slog.Default().Warn("prune player reports failed", "err", err)
	}
}

// RunReportSweeper prunes old reports until ctx is done: once at start
// and then every interval. The prune also runs whenever a report is filed
// or the list is read; the sweep is what keeps the 90-day promise on a
// server nobody files or reads on. interval <= 0 means ReportSweepInterval.
func RunReportSweeper(ctx context.Context, l *Lobby, log *slog.Logger, interval time.Duration) {
	if interval <= 0 {
		interval = ReportSweepInterval
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		if n, err := l.PruneReports(time.Now().UTC()); err != nil && log != nil {
			log.Error("prune player reports failed", "err", err)
		} else if n > 0 && log != nil {
			log.Info("pruned old player reports", "removed", n)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

// --- HTTP ---------------------------------------------------------------

type reportRequest struct {
	// PlayerID is the reported seat's player id. SpectatorID names a
	// spectator instead (GET /games/{id}/spectators); exactly one is set.
	PlayerID    string `json:"player_id,omitempty"`
	SpectatorID string `json:"spectator_id,omitempty"`
	Reason      string `json:"reason"`
}

type reportFiled struct {
	ID string `json:"id"`
}

// reportPlayer handles POST /games/{id}/reports.
//
//	401  anonymous
//	403  the caller does not hold a seat at this table
//	404  no such table
//	400  no reason, or a malformed player id
//	404  the reported spectator is not watching this table
//	422  your own seat, a bot, or a player who is not at this table
//	429  too many reports (per reporter; see the route's limiter)
//	201  {"id": "<report id>"}
func reportPlayer(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	// A seat's own session, bound to this table. An admin token, a
	// spectator and a player of another table are not reporters.
	if p.Role != auth.RolePlayer || p.GameID != id || p.PlayerID == uuid.Nil {
		return httpError(http.StatusForbidden, "only a player at this table can report another")
	}
	var body reportRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	in := ReportInput{GameID: id, ReporterPlayerID: p.PlayerID, Reason: body.Reason}
	switch {
	case body.SpectatorID != "" && body.PlayerID != "":
		return httpError(http.StatusBadRequest, "name a player_id or a spectator_id, not both")
	case body.SpectatorID != "":
		specID, perr := uuid.Parse(body.SpectatorID)
		if perr != nil {
			return httpError(http.StatusBadRequest, "spectator_id must be a spectator's id")
		}
		if c.SpectatorHub == nil {
			return ErrNoSpectatorHub
		}
		who, found := findSpectator(c.SpectatorHub.Spectators(id), specID)
		if !found {
			return httpError(http.StatusNotFound, ErrSpectatorNotWatching.Error())
		}
		in.ReportedSpectator = &who
	default:
		target, perr := uuid.Parse(body.PlayerID)
		if perr != nil {
			return httpError(http.StatusBadRequest, "player_id must be a seat's player id")
		}
		in.ReportedPlayerID = target
	}
	rec, err := c.Lobby.FileReport(in)
	if err != nil {
		return err
	}
	c.logger().Info("player report filed", "game_id", id, "report_id", rec.ID)
	return writeJSON(w, http.StatusCreated, reportFiled{ID: rec.ID.String()})
}

type adminReportParty struct {
	PlayerID string `json:"player_id"`
	UserID   string `json:"user_id,omitempty"`
	Name     string `json:"name"`
}

type adminReport struct {
	ID        string           `json:"id"`
	CreatedAt int64            `json:"created_at"` // Unix milliseconds
	GameID    string           `json:"game_id"`
	GameName  string           `json:"game_name"`
	Reporter  adminReportParty `json:"reporter"`
	Reported  adminReportParty `json:"reported"`
	Reason    string           `json:"reason"`
	Excerpt   []ReportLine     `json:"excerpt"`
}

type adminReportsResponse struct {
	Reports       []adminReport `json:"reports"`
	RetentionDays int           `json:"retention_days"`
}

// adminListReports handles GET /admin/reports.
func adminListReports(c Config, w http.ResponseWriter, _ *http.Request) error {
	recs, err := c.Lobby.Reports()
	if err != nil {
		c.logger().Error("list player reports failed", "err", err)
		return httpError(http.StatusInternalServerError, "could not load the reports; try again")
	}
	out := adminReportsResponse{Reports: make([]adminReport, 0, len(recs)), RetentionDays: int(ReportRetention / (24 * time.Hour))}
	for _, r := range recs {
		ex := r.Excerpt
		if ex == nil {
			ex = []ReportLine{}
		}
		out.Reports = append(out.Reports, adminReport{
			ID:        r.ID.String(),
			CreatedAt: r.CreatedAt.UnixMilli(),
			GameID:    r.GameID.String(),
			GameName:  r.GameName,
			Reporter:  adminReportParty{PlayerID: r.ReporterPlayerID.String(), UserID: r.ReporterUserID, Name: r.ReporterName},
			Reported:  adminReportParty{PlayerID: r.ReportedPlayerID.String(), UserID: r.ReportedUserID, Name: r.ReportedName},
			Reason:    r.Reason,
			Excerpt:   ex,
		})
	}
	return writeJSON(w, http.StatusOK, out)
}

// adminDismissReport handles DELETE /admin/reports/{id}.
func adminDismissReport(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return httpError(http.StatusBadRequest, "invalid report id")
	}
	if err := c.Lobby.DismissReport(id); err != nil {
		if errors.Is(err, ErrStoreNotFound) {
			return httpError(http.StatusNotFound, "no such report")
		}
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
