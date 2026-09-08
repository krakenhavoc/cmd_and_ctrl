package lobby

// bugreport.go — the in-app "report a bug" surface. The client's bug
// button POSTs here; the server renders a markdown issue body and
// files it in the project's GitHub repo through the BugReporter seam.
// The GitHub token never reaches the client (ADR 0017): the browser
// talks only to this endpoint, the server talks to GitHub.
//
// A report can carry three things beyond the reporter's prose:
//
//   - the client-side log ring buffer (what the reporter's own browser
//     saw: actions sent, snapshot seqs, socket drops, JS errors). Safe
//     to inline in the issue because it is, by construction, only what
//     that browser already had.
//   - screenshots, stored by bugstore and linked absolutely so
//     GitHub's image proxy renders them inline.
//   - a pinned copy of the game's replay JSONL, kept behind admin auth
//     and referenced only by report ID.
//
// The split matters: ADR 0017 §4 kept ALL game state out of issues to
// stop hidden information leaking around the wire visibility filter.
// That invariant still holds — the raw replay is never in the body,
// and the client log holds no state the reporter couldn't already see
// (ADR 0017 §7 records the revision).

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/bugstore"
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

	// bugLogMaxEntries caps how many client log lines a report may
	// carry. The client's own ring buffer is the same size, so this
	// is a "don't trust the client" ceiling rather than a policy.
	bugLogMaxEntries = 200

	// bugLogLineMax clips one rendered log line. Frames are logged as
	// short summaries, so anything longer is either a pathological
	// chat message or a client that isn't ours.
	bugLogLineMax = 240

	// bugLogTotalMax caps the rendered log block. GitHub renders long
	// issue bodies fine but nobody reads 200 KiB of it, and the block
	// is inside a <details> anyway.
	bugLogTotalMax = 24 << 10

	// bugMultipartMax caps a whole multipart submission: the JSON
	// part plus every image. Sized as bugstore.MaxTotalImageBytes
	// plus headroom for the prose, the log, and part boundaries.
	bugMultipartMax = bugstore.MaxTotalImageBytes + (2 << 20)

	// bugMultipartMemory is how much of a multipart body is buffered
	// in RAM before parts spill to temp files.
	bugMultipartMemory = 1 << 20
)

// bugLogKinds is the allowlist of client log-entry kinds. Anything
// else is rendered as "info" rather than rejected: a log line is
// diagnostic colour, never worth failing a report over.
var bugLogKinds = map[string]bool{
	"sent":     true,
	"received": true,
	"error":    true,
	"info":     true,
	"console":  true,
}

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

// bugLogEntry is one line of the reporter's client-side log.
//
// At is epoch milliseconds from the REPORTER'S clock, not ours: the
// server formats it rather than accepting a preformatted string, so
// there is no client-supplied timestamp text to sanitise and a skewed
// clock can't forge a plausible-looking server time.
type bugLogEntry struct {
	At   int64  `json:"at"`
	Kind string `json:"kind"`
	Text string `json:"text"`
}

type bugReportRequest struct {
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Context     *bugReportContext `json:"context,omitempty"`
	Log         []bugLogEntry     `json:"log,omitempty"`
}

// bugReportConfig mirrors /auth/discord/config: an unauthenticated
// enabled-flag probe the client uses to decide whether to render
// the report button at all. Leaks nothing but the bools.
//
// attachments is reported separately from enabled: a deploy with a
// GitHub token but no data dir (or no public base URL) can still file
// text reports, and the modal hides its file picker rather than
// offering an upload that would silently do nothing.
func bugReportConfig(c Config, w http.ResponseWriter, _ *http.Request) error {
	return writeJSON(w, http.StatusOK, map[string]any{
		"enabled":     c.BugReporter != nil,
		"attachments": c.BugStore.Enabled(),
		"max_images":  bugstore.MaxImages,
		"max_bytes":   bugstore.MaxImageBytes,
	})
}

// bugReport handles POST /bugreport. Session-gated (any role —
// spectators hit bugs too) and rate-limited at the mux. 201 with
// {url, number, report_id} on success so the client can link straight
// to the filed issue.
//
// Accepts either application/json (prose + context + log) or
// multipart/form-data with a `report` part holding that same JSON and
// zero or more `image` file parts. Two shapes rather than one because
// the JSON path predates attachments and is the one the tests, the
// docs, and any curl-wielding operator already know.
func bugReport(c Config, w http.ResponseWriter, r *http.Request) error {
	if c.BugReporter == nil {
		return httpError(http.StatusServiceUnavailable, "bug reporting not configured on this server")
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}

	var req bugReportRequest
	var images [][]byte
	var err error
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		req, images, err = decodeBugMultipart(w, r)
	} else {
		err = decodeJSON(w, r, &req)
	}
	if err != nil {
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
	case len(images) > bugstore.MaxImages:
		return httpError(http.StatusBadRequest, fmt.Sprintf("at most %d images per report", bugstore.MaxImages))
	}

	// Artifacts are assembled before the issue is filed, because the
	// image URLs and report ID have to appear in the body. If filing
	// then fails, the report is discarded — otherwise every GitHub
	// outage would leave orphaned images on disk at live URLs.
	report, err := buildBugArtifacts(c, p, req.Context, images)
	if err != nil {
		return err
	}

	in := bugIssue{
		Principal: p,
		Desc:      strings.TrimSpace(req.Description),
		Ctx:       req.Context,
		Log:       req.Log,
		UserAgent: r.UserAgent(),
		Now:       time.Now().UTC(),
	}
	if report != nil {
		in.ReportID = report.ID
		in.Images = report.Images()
		in.Replay = report.Replay()
	}

	url, number, err := c.BugReporter.CreateIssue(r.Context(), bugTitlePrefix+title, renderBugIssueBody(in), []string{"bug"})
	if err != nil {
		report.discard()
		// 502: the report was well-formed, the upstream filing
		// failed. The client shows a "try again / tell the admin"
		// message; err carries no secrets (github.Client redacts).
		return httpError(http.StatusBadGateway, fmt.Sprintf("filing the issue failed: %s", err))
	}
	if report != nil {
		gameID := ""
		if req.Context != nil {
			gameID = req.Context.GameID
		}
		// A manifest that fails to write costs forensics later, not
		// the report the user just filed — log-and-continue.
		if err := report.Commit(gameID, url); err != nil {
			logBugStoreWarning(c, "write bug report manifest", err)
		}
		// Retention is enforced opportunistically here rather than on
		// a timer: there are only ever a handful of directories, and
		// a report is the only thing that grows the set.
		if _, err := c.BugStore.Prune(bugstore.DefaultRetention, bugstore.DefaultMaxStoreBytes); err != nil && !errors.Is(err, bugstore.ErrDisabled) {
			logBugStoreWarning(c, "prune bug reports", err)
		}
	}

	body := map[string]any{"url": url, "number": number}
	if report != nil {
		body["report_id"] = report.ID
	}
	return writeJSON(w, http.StatusCreated, body)
}

// bugArtifacts wraps *bugstore.Report so the handler can treat "no
// store configured" and "store configured" uniformly — a nil
// *bugArtifacts answers every call harmlessly.
type bugArtifacts struct{ *bugstore.Report }

func (b *bugArtifacts) discard() {
	if b != nil {
		b.Report.Discard()
	}
}

// buildBugArtifacts stores the images and pins the replay, returning
// nil when there is nothing to store or no store to store it in.
func buildBugArtifacts(c Config, p auth.Principal, bctx *bugReportContext, images [][]byte) (*bugArtifacts, error) {
	if !c.BugStore.Enabled() {
		if len(images) > 0 {
			return nil, httpError(http.StatusServiceUnavailable, "attachments are not configured on this server")
		}
		return nil, nil
	}
	replaySrc := bugReplaySource(c, p, bctx)
	if len(images) == 0 && replaySrc == "" {
		return nil, nil
	}

	rep, err := c.BugStore.NewReport()
	if err != nil {
		return nil, httpError(http.StatusInternalServerError, "preparing the report failed")
	}
	out := &bugArtifacts{Report: rep}
	for _, data := range images {
		if _, err := rep.AddImage(data); err != nil {
			out.discard()
			switch {
			case errors.Is(err, bugstore.ErrUnsupportedType):
				return nil, httpError(http.StatusBadRequest, "attachments must be PNG, JPEG, GIF, or WebP images")
			case errors.Is(err, bugstore.ErrTooLarge):
				return nil, httpError(http.StatusRequestEntityTooLarge, fmt.Sprintf("images must be under %d MiB each and %d MiB in total", bugstore.MaxImageBytes>>20, bugstore.MaxTotalImageBytes>>20))
			case errors.Is(err, bugstore.ErrTooMany):
				return nil, httpError(http.StatusBadRequest, fmt.Sprintf("at most %d images per report", bugstore.MaxImages))
			default:
				return nil, httpError(http.StatusInternalServerError, "storing the attachment failed")
			}
		}
	}
	if replaySrc != "" {
		// A missing or unreadable replay is not a reason to lose the
		// report: the issue simply says no replay was pinned.
		if _, err := rep.PinReplay(replaySrc); err != nil && !errors.Is(err, bugstore.ErrNotFound) {
			logBugStoreWarning(c, "pin replay", err)
		}
	}
	return out, nil
}

// bugReplaySource resolves the on-disk replay path to pin, or "" when
// there is nothing to pin or the reporter isn't entitled to it.
//
// Entitlement mirrors downloadReplay: admins anywhere, everyone else
// only for the game their session is bound to. The pinned copy is
// admin-only to READ, so this check isn't what protects the contents
// — it stops a session in game A from making us copy game B's log.
func bugReplaySource(c Config, p auth.Principal, bctx *bugReportContext) string {
	if bctx == nil || bctx.GameID == "" || c.Lobby == nil {
		return ""
	}
	id, err := uuid.Parse(bctx.GameID)
	if err != nil {
		return ""
	}
	if p.Role != auth.RoleAdmin && p.GameID != id {
		return ""
	}
	room := c.Lobby.RoomOf(id)
	if room == nil {
		return ""
	}
	return room.ReplayPath()
}

// decodeBugMultipart parses the multipart form: one `report` field
// carrying the same JSON the plain path accepts, plus `image` file
// parts.
func decodeBugMultipart(w http.ResponseWriter, r *http.Request) (bugReportRequest, [][]byte, error) {
	var req bugReportRequest
	r.Body = http.MaxBytesReader(w, r.Body, bugMultipartMax)
	if err := r.ParseMultipartForm(bugMultipartMemory); err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			return req, nil, httpError(http.StatusRequestEntityTooLarge, fmt.Sprintf("report exceeds the %d MiB limit", bugMultipartMax>>20))
		}
		return req, nil, httpError(http.StatusBadRequest, "invalid multipart body: "+err.Error())
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()

	raw := r.FormValue("report")
	if raw == "" {
		return req, nil, httpError(http.StatusBadRequest, "missing report field")
	}
	if err := decodeJSONString(raw, &req); err != nil {
		return req, nil, httpError(http.StatusBadRequest, "invalid report field: "+err.Error())
	}

	files := r.MultipartForm.File["image"]
	if len(files) > bugstore.MaxImages {
		return req, nil, httpError(http.StatusBadRequest, fmt.Sprintf("at most %d images per report", bugstore.MaxImages))
	}
	images := make([][]byte, 0, len(files))
	for _, fh := range files {
		if fh.Size > bugstore.MaxImageBytes {
			return req, nil, httpError(http.StatusRequestEntityTooLarge, fmt.Sprintf("each image must be under %d MiB", bugstore.MaxImageBytes>>20))
		}
		f, err := fh.Open()
		if err != nil {
			return req, nil, httpError(http.StatusBadRequest, "unreadable attachment")
		}
		// LimitReader as well as the Size check: Size comes from the
		// part header and a hostile client controls it.
		data, err := io.ReadAll(io.LimitReader(f, bugstore.MaxImageBytes+1))
		_ = f.Close()
		if err != nil {
			return req, nil, httpError(http.StatusBadRequest, "unreadable attachment")
		}
		if int64(len(data)) > bugstore.MaxImageBytes {
			return req, nil, httpError(http.StatusRequestEntityTooLarge, fmt.Sprintf("each image must be under %d MiB", bugstore.MaxImageBytes>>20))
		}
		images = append(images, data)
	}
	return req, images, nil
}

// bugAttachment handles GET /bugreport/att/{id}/{name} — the one
// unauthenticated route in this feature.
//
// It has to be unauthenticated: GitHub renders an issue image by
// fetching it through the Camo proxy, which presents no session and
// follows no login. The compensating controls live in bugstore
// (uuid-keyed paths, magic-byte allowlist, inert response headers);
// the capability is the report ID itself, which only ever appears
// inside a private-repo issue body.
func bugAttachment(c Config, w http.ResponseWriter, r *http.Request) error {
	if !c.BugStore.Enabled() {
		return httpError(http.StatusServiceUnavailable, "attachments are not configured on this server")
	}
	err := c.BugStore.ServeImage(w, r, r.PathValue("id"), r.PathValue("name"))
	switch {
	case err == nil:
		return nil
	case errors.Is(err, bugstore.ErrNotFound):
		return httpError(http.StatusNotFound, "attachment not found")
	default:
		return httpError(http.StatusInternalServerError, "serving the attachment failed")
	}
}

// bugPinnedReplay handles GET /bugreport/{id}/replay — admin only.
//
// This is the raw, UNFILTERED replay: opponents' hands and full
// library order. It is admin-gated for the same reason
// /games/{id}/replay is, and unlike that route there is no
// game-has-ended relaxation — a pinned replay has no live game to
// reason about, so the strict rule is the only safe one.
func bugPinnedReplay(c Config, w http.ResponseWriter, r *http.Request) error {
	if !c.BugStore.Enabled() {
		return httpError(http.StatusServiceUnavailable, "bug report storage is not configured on this server")
	}
	f, info, err := c.BugStore.OpenReplay(r.PathValue("id"))
	if err != nil {
		if errors.Is(err, bugstore.ErrNotFound) {
			return httpError(http.StatusNotFound, "no pinned replay for that report")
		}
		return httpError(http.StatusInternalServerError, "opening the pinned replay failed")
	}
	// ServeContent does NOT close its ReadSeeker, and it completes
	// before this handler returns — defer is the whole cleanup story.
	defer func() { _ = f.Close() }()
	name := r.PathValue("id") + ".jsonl"
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	http.ServeContent(w, r, name, info.ModTime(), f)
	return nil
}

// bugIssue is the input to renderBugIssueBody — a struct rather than
// a positional argument list because the body now assembles from six
// independent sources and a seventh would have been unreadable.
type bugIssue struct {
	Principal auth.Principal
	Desc      string
	Ctx       *bugReportContext
	Log       []bugLogEntry
	UserAgent string
	Now       time.Time
	ReportID  string
	Images    []bugstore.Image
	Replay    *bugstore.ReplayPin
}

// renderBugIssueBody assembles the markdown issue body: the
// reporter's description, then screenshots, then the client log in a
// collapsed block, then a metadata footer. Pure — covered directly by
// unit tests.
//
// Deliberately NOT included: the replay itself, or any GameView
// content. Only the report ID appears, and pulling the replay behind
// it takes admin credentials — so an issue is never a side channel
// around the wire visibility filter, even for a reporter who is also
// a repo collaborator (ADR 0017 §4, revised in §7).
func renderBugIssueBody(in bugIssue) string {
	var b strings.Builder
	desc := in.Desc
	if desc == "" {
		desc = "_no description provided_"
	}
	b.WriteString(desc)

	if len(in.Images) > 0 {
		b.WriteString("\n\n---\n\n**Screenshots**\n\n")
		for _, img := range in.Images {
			// Absolute URL, plain markdown image: GitHub proxies it
			// through Camo and renders it inline in the issue.
			fmt.Fprintf(&b, "![%s](%s)\n\n", img.Name, img.URL)
		}
		b.WriteString("<sub>Uploaded by the reporter. Served from the game server; expires after " +
			fmt.Sprintf("%d days", int(bugstore.DefaultRetention.Hours()/24)) + ".</sub>\n")
	}

	if block, n := renderBugLog(in.Log); n > 0 {
		fmt.Fprintf(&b, "\n\n---\n\n<details>\n<summary>Client log — last %d entries before the report</summary>\n\n```text\n%s```\n\n", n, block)
		b.WriteString("<sub>Captured in the reporter's browser; times are that browser's clock. " +
			"Contains only what that client already saw — no hidden game state.</sub>\n</details>\n")
	}

	b.WriteString("\n\n---\n\n**Reported from the app**\n\n")
	fmt.Fprintf(&b, "- Reporter: %s (%s)\n", reporterName(in.Principal), in.Principal.Role)
	if in.Ctx != nil && in.Ctx.GameID != "" {
		fmt.Fprintf(&b, "- Game: `%s`", clip(in.Ctx.GameID, bugFieldMax))
		if in.Ctx.Turn > 0 {
			fmt.Fprintf(&b, " — turn %d, %s/%s", in.Ctx.Turn, clip(in.Ctx.Phase, bugFieldMax), clip(in.Ctx.Step, bugFieldMax))
		}
		if in.Ctx.Seq > 0 {
			fmt.Fprintf(&b, " (seq %d)", in.Ctx.Seq)
		}
		b.WriteString("\n")
		if in.Ctx.Connection != "" {
			fmt.Fprintf(&b, "- Connection: %s\n", clip(in.Ctx.Connection, bugFieldMax))
		}
	}
	if in.ReportID != "" {
		fmt.Fprintf(&b, "- Report ID: `%s`\n", in.ReportID)
	}
	switch {
	case in.Replay != nil:
		fmt.Fprintf(&b, "- Pinned replay: `GET /bugreport/%s/replay` (admin) — %d lines, %d KiB, snapshotted at report time",
			in.ReportID, in.Replay.Lines, in.Replay.Bytes>>10)
		if in.Replay.Truncated {
			// Say so loudly: a truncated pin cannot be replayed from
			// turn one, which is the first thing an operator assumes.
			b.WriteString(" — **tail only** (source exceeded the pin cap)")
		}
		b.WriteString("\n")
	case in.Ctx != nil && in.Ctx.GameID != "":
		// No pin: either storage is off or the game had produced no
		// replay yet. The live route is the fallback, with the caveat
		// that it disappears when the game is evicted.
		fmt.Fprintf(&b, "- Replay: `GET /games/%s/replay` (admin; not pinned — gone once the game is evicted)\n", clip(in.Ctx.GameID, bugFieldMax))
	}
	fmt.Fprintf(&b, "- Server time: %s\n", in.Now.Format(time.RFC3339))
	if in.UserAgent != "" {
		fmt.Fprintf(&b, "- User agent: %s\n", clip(in.UserAgent, 200))
	}
	return b.String()
}

// renderBugLog formats the client log ring buffer as fixed-width
// lines, returning the block and how many entries it holds.
//
// Every field is client-supplied, and one of them (chat text) is
// typed by other players — so backticks are stripped rather than
// escaped. A stray fence inside a ``` block would end it early and
// let the rest render as markdown, which is how a "log line" turns
// into a forged metadata row.
func renderBugLog(entries []bugLogEntry) (string, int) {
	if len(entries) == 0 {
		return "", 0
	}
	if len(entries) > bugLogMaxEntries {
		// Keep the tail — the bug is at the end of the log.
		entries = entries[len(entries)-bugLogMaxEntries:]
	}
	var b strings.Builder
	n := 0
	for _, e := range entries {
		kind := e.Kind
		if !bugLogKinds[kind] {
			kind = "info"
		}
		line := fmt.Sprintf("%s  %-8s  %s\n", bugLogTime(e.At), kind, sanitizeLogText(e.Text))
		if b.Len()+len(line) > bugLogTotalMax {
			break
		}
		b.WriteString(line)
		n++
	}
	if n == 0 {
		return "", 0
	}
	return b.String(), n
}

// bugLogTime renders a client epoch-millis stamp as a UTC wall clock.
// Out-of-range values (a zero field, a clock set to 1970 or 2400)
// render as a placeholder rather than a misleading date.
func bugLogTime(ms int64) string {
	if ms <= 0 {
		return "--:--:--.---"
	}
	t := time.UnixMilli(ms).UTC()
	if t.Year() < 2000 || t.Year() > 2100 {
		return "--:--:--.---"
	}
	return t.Format("15:04:05.000")
}

// sanitizeLogText makes one client log line safe to drop inside a
// fenced code block: no backticks (fence escape), no control
// characters or newlines (forged extra lines), length-clipped.
func sanitizeLogText(s string) string {
	s = strings.Map(func(r rune) rune {
		switch {
		case r == '`':
			return '\''
		case r == '\n' || r == '\r' || r == '\t':
			return ' '
		case r < 0x20 || r == 0x7f:
			return -1
		default:
			return r
		}
	}, s)
	return clip(s, bugLogLineMax)
}

// logBugStoreWarning reports a non-fatal storage problem. Storage
// failures must never cost a report that GitHub already accepted, so
// every caller logs and carries on.
func logBugStoreWarning(c Config, what string, err error) {
	if c.Log == nil {
		return
	}
	c.Log.Warn("bugreport: "+what+" failed", "err", err)
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
