package lobby

// bugreport.go — the in-app "report a bug" surface. The client's
// bug button POSTs here; the server renders a markdown issue body
// and files it in the project's GitHub repo through the BugReporter
// seam. The GitHub token never reaches the client (ADR 0017): the
// browser talks only to this endpoint, the server talks to GitHub.

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
)

// BugReporter files a bug report as an issue in the project tracker.
// Satisfied by *github.Client; narrow so handler tests can record
// submissions without talking to GitHub (same seam pattern as
// GameEvictor / StateBroadcaster). Nil in Config disables the
// surface: POST /bugreport 503s and GET /bugreport/config reports
// enabled:false so the client hides the button — the same
// convention as the Discord OAuth and avatar-cache seams.
type BugReporter interface {
	CreateIssue(ctx context.Context, title, body string, labels []string) (url string, number int, err error)
}

const (
	// bugTitleMax / bugDescMax bound user-typed fields. GitHub's own
	// limits are far larger; these keep issues scannable and stop a
	// stuck key from filing a novel.
	bugTitleMax = 200
	bugDescMax  = 5000

	// bugFieldMax bounds each client-supplied context string. The
	// context fields are echoes of server-issued values (game ID,
	// phase name, …), so anything longer is garbage by definition.
	bugFieldMax = 64

	// bugTitlePrefix marks issues that arrived via the in-app
	// button, so the tracker can tell them from hand-written ones
	// at a glance.
	bugTitlePrefix = "[in-app] "
)

// bugReportContext is the client-collected game context attached to
// a report. Every field is optional — a report filed from the lobby
// has no turn to speak of. Values are clipped, never trusted: they
// only ever land inside a markdown code span in the issue body.
type bugReportContext struct {
	GameID     string `json:"game_id,omitempty"`
	Turn       int    `json:"turn,omitempty"`
	Phase      string `json:"phase,omitempty"`
	Step       string `json:"step,omitempty"`
	Seq        uint64 `json:"seq,omitempty"`
	Connection string `json:"connection,omitempty"`
}

type bugReportRequest struct {
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Context     *bugReportContext `json:"context,omitempty"`
}

// bugReportConfig mirrors /auth/discord/config: an unauthenticated
// enabled-flag probe the client uses to decide whether to render
// the report button at all. Leaks nothing but the bool.
func bugReportConfig(c Config, w http.ResponseWriter, _ *http.Request) error {
	return writeJSON(w, http.StatusOK, map[string]bool{"enabled": c.BugReporter != nil})
}

// bugReport handles POST /bugreport. Session-gated (any role —
// spectators hit bugs too) and rate-limited at the mux. 201 with
// {url, number} on success so the client can link straight to the
// filed issue.
func bugReport(c Config, w http.ResponseWriter, r *http.Request) error {
	if c.BugReporter == nil {
		return httpError(http.StatusServiceUnavailable, "bug reporting not configured on this server")
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}

	var req bugReportRequest
	if err := decodeJSON(w, r, &req); err != nil {
		return err
	}
	title := strings.TrimSpace(req.Title)
	switch {
	case title == "":
		return httpError(http.StatusBadRequest, "title is required")
	case len(title) > bugTitleMax:
		return httpError(http.StatusBadRequest, fmt.Sprintf("title exceeds %d characters", bugTitleMax))
	case len(req.Description) > bugDescMax:
		return httpError(http.StatusBadRequest, fmt.Sprintf("description exceeds %d characters", bugDescMax))
	}

	body := renderBugIssueBody(p, strings.TrimSpace(req.Description), req.Context, r.UserAgent(), time.Now().UTC())
	url, number, err := c.BugReporter.CreateIssue(r.Context(), bugTitlePrefix+title, body, []string{"bug"})
	if err != nil {
		// 502: the report was well-formed, the upstream filing
		// failed. The client shows a "try again / tell the admin"
		// message; err carries no secrets (github.Client redacts).
		return httpError(http.StatusBadGateway, fmt.Sprintf("filing the issue failed: %s", err))
	}
	return writeJSON(w, http.StatusCreated, map[string]any{"url": url, "number": number})
}

// renderBugIssueBody assembles the markdown issue body: the
// reporter's description first, then a metadata block. Pure —
// covered directly by unit tests.
//
// Deliberately NOT included: any game state (hands, libraries,
// decklists). The game ID is enough for the operator to pull the
// authoritative replay via GET /games/{id}/replay; embedding state
// in the issue would route hidden information around the wire
// visibility filter (see ADR 0017).
func renderBugIssueBody(p auth.Principal, desc string, bctx *bugReportContext, userAgent string, now time.Time) string {
	var b strings.Builder
	if desc == "" {
		desc = "_no description provided_"
	}
	b.WriteString(desc)
	b.WriteString("\n\n---\n\n**Reported from the app**\n\n")
	fmt.Fprintf(&b, "- Reporter: %s (%s)\n", reporterName(p), p.Role)
	if bctx != nil && bctx.GameID != "" {
		fmt.Fprintf(&b, "- Game: `%s`", clip(bctx.GameID, bugFieldMax))
		if bctx.Turn > 0 {
			fmt.Fprintf(&b, " — turn %d, %s/%s", bctx.Turn, clip(bctx.Phase, bugFieldMax), clip(bctx.Step, bugFieldMax))
		}
		if bctx.Seq > 0 {
			fmt.Fprintf(&b, " (seq %d)", bctx.Seq)
		}
		b.WriteString("\n")
		if bctx.Connection != "" {
			fmt.Fprintf(&b, "- Connection: %s\n", clip(bctx.Connection, bugFieldMax))
		}
		fmt.Fprintf(&b, "- Replay: `GET /games/%s/replay` (admin)\n", clip(bctx.GameID, bugFieldMax))
	}
	fmt.Fprintf(&b, "- Server time: %s\n", now.Format(time.RFC3339))
	if userAgent != "" {
		fmt.Fprintf(&b, "- User agent: %s\n", clip(userAgent, 200))
	}
	return b.String()
}

// reporterName picks the friendliest available identity: seat name,
// then Discord display name, then the role as a last resort.
func reporterName(p auth.Principal) string {
	switch {
	case p.Name != "":
		return p.Name
	case p.DiscordGlobalName != "":
		return p.DiscordGlobalName
	case p.DiscordUsername != "":
		return p.DiscordUsername
	default:
		return string(p.Role)
	}
}

// clip bounds a client-supplied string and strips newlines so a
// crafted value can't fake extra markdown list rows in the issue.
func clip(s string, n int) string {
	s = strings.NewReplacer("\n", " ", "\r", " ").Replace(s)
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}
