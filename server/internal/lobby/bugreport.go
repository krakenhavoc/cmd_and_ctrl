package lobby

// bugreport.go — the in-app "report a bug" surface. The client's bug
// button POSTs here; the server renders a markdown issue body and
// files it in the project's GitHub repo through the BugReporter seam.
// The GitHub token never reaches the client (ADR 0017): the browser
// talks only to this endpoint, the server talks to GitHub.
//
// A report also carries a KIND — bug, idea, or question (ADR 0017
// §8). The kind picks the tracker label and, just as importantly,
// picks which artifacts the report collects: a bug is worth a pinned
// replay, "the stack panel should be wider" is not. The mapping lives
// in bugKinds below and is server-authoritative — the client names a
// kind, never a label.
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
//   - a pinned copy of the PUBLIC game log (S31 sub-PR 0), same admin
//     posture as the replay. It is the artifact that makes a report
//     readable: the replay says what the state WAS at every frame, the
//     log says what HAPPENED, and "what happened" is what a bug report
//     is about.
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
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/bugstore"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
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
	//
	// Deliberately NOT varied by kind. Every issue this feature has
	// ever filed carries this exact string and people filter on it;
	// widening it to "[in-app bug] " / "[in-app idea] " would orphan
	// that convention on every open issue and give the kind two
	// sources of truth — one of which (the title) triage can't fix
	// with a click. The label is the kind (ADR 0017 §8).
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

// bugKind is a report kind: what the player is telling us, which
// decides the tracker label and the artifacts worth collecting.
//
// Label is looked up here rather than sent by the client. That is the
// whole reason this table exists: a client-supplied label string
// would let any session with a report button attach `good first
// issue` — or create labels — on the project tracker.
type bugKind struct {
	// Label is an EXISTING repo label. Adding a kind means picking
	// one of bug / enhancement / documentation / question / … — not
	// inventing a taxonomy GitHub then auto-creates on first use.
	Label string

	// Noun is how the issue body names the kind, so a report stays
	// self-describing if the label is ever stripped (or never landed
	// — see fileBugIssue).
	Noun string

	// Log inlines the reporter's client log in a <details> block.
	// Log keeps its ADR 0017 §7 posture wherever it is attached: it
	// holds only what that browser already saw.
	Log bool

	// Replay pins the game's replay JSONL. The expensive one — tens
	// of MiB per report, retained 90 days — so it is reserved for
	// the kind that is actually triaged frame by frame.
	Replay bool

	// GameLog pins the public game log: a few KiB of "what
	// happened", which is what answers a question as often as it
	// diagnoses a bug.
	GameLog bool
}

// bugKinds maps wire kind → policy. The wire word is the player's
// word ("idea"); the label is the tracker's word ("enhancement").
//
// Screenshots are attached for every kind — a picture is the cheapest
// thing in the feature and the most useful for a UI idea. Whatever is
// attached keeps ADR 0017's split: images are public (Camo has to
// reach them), pins are admin-only.
var bugKinds = map[string]bugKind{
	// A bug gets everything: the log says what the client did, the
	// pinned replay says what the state was, the game log says what
	// happened. That triple is the whole value of the feature.
	"bug": {Label: "bug", Noun: "bug", Log: true, Replay: true, GameLog: true},

	// An idea is about what the game SHOULD do, so none of the
	// forensic artifacts apply — a screenshot of the thing being
	// complained about is the evidence, and pinning a 30 MiB replay
	// to "the stack panel should be wider" is pure waste.
	"idea": {Label: "enhancement", Noun: "idea"},

	// A question sits between the two. "Why did my creature die?" is
	// answered from what HAPPENED — the public game log, a few KiB —
	// and from what the client did. It is not answered by
	// frame-by-frame hidden state, which is the replay's job, so the
	// expensive pin stays off.
	"question": {Label: "question", Noun: "question", Log: true, GameLog: true},
}

// defaultBugKind is what an omitted kind means. It has to be "bug":
// every client shipped before this field existed sends no kind, and
// those are bug reports. A 400 there would break the button for an
// in-flight tab mid-game, which is exactly when it gets used.
const defaultBugKind = "bug"

// resolveBugKind maps the request's kind to its policy. Empty →
// bug; anything else must be in the table.
//
// An unknown non-empty kind is a 400 rather than a silent fallback:
// this endpoint already rejects unknown JSON fields outright, and a
// kind we quietly rewrite to "bug" is a mislabelled issue nobody ever
// finds out about. Only our own client sets the field, and it can
// only emit these three.
func resolveBugKind(raw string) (string, bugKind, error) {
	name := strings.ToLower(strings.TrimSpace(raw))
	if name == "" {
		name = defaultBugKind
	}
	k, ok := bugKinds[name]
	if !ok {
		return "", bugKind{}, httpError(http.StatusBadRequest,
			fmt.Sprintf("kind must be one of %s", strings.Join(bugKindNames(), ", ")))
	}
	return name, k, nil
}

// bugKindNames lists the accepted kinds in a stable order for error
// messages (map iteration order would make the 400 body flap).
func bugKindNames() []string {
	names := make([]string, 0, len(bugKinds))
	for name := range bugKinds {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

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
	Title string `json:"title"`
	// Kind is "bug", "idea", or "question". Optional on the wire:
	// omitted means "bug", which is what every client that predates
	// the field is filing.
	Kind        string            `json:"kind,omitempty"`
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
		"attachments": c.BugStore.AttachmentsEnabled(),
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

	kindName, kind, err := resolveBugKind(req.Kind)
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
	report, err := buildBugArtifacts(c, p, kind, req.Context, images)
	if err != nil {
		return err
	}

	in := bugIssue{
		Principal: p,
		Kind:      kind,
		KindName:  kindName,
		Desc:      strings.TrimSpace(req.Description),
		Ctx:       req.Context,
		UserAgent: r.UserAgent(),
		Now:       time.Now().UTC(),
	}
	// The log rides along only for the kinds that can use it. A
	// client that sends one anyway (an older build, a curl) has it
	// dropped rather than the report refused — the server decides
	// what an issue carries.
	if kind.Log {
		in.Log = req.Log
	}
	if report != nil {
		in.ReportID = report.ID
		in.Images = report.Images()
		in.Replay = report.Replay()
		in.GameLog = report.GameLog()
	}

	url, number, label, err := fileBugIssue(r.Context(), c, bugTitlePrefix+title, renderBugIssueBody(in), kind.Label)
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
	if label != "" {
		// Echoed so the modal can say "filed as enhancement" — and so
		// it says nothing when the label didn't land, rather than
		// claiming a label the issue doesn't have.
		body["label"] = label
	}
	if report != nil {
		body["report_id"] = report.ID
	}
	return writeJSON(w, http.StatusCreated, body)
}

// statusCoder is implemented by upstream errors carrying an HTTP
// status — github.APIError does. Declared here rather than imported
// so the BugReporter seam stays one method wide and a test fake can
// satisfy it without pulling in the real client.
type statusCoder interface{ StatusCode() int }

// fileBugIssue files the issue and returns the label it actually
// landed with ("" if none did).
//
// A label must never cost us the report. If GitHub REJECTS the
// request because of the label — the repo doesn't have it, or the
// token may not apply it — the issue is re-filed unlabelled and the
// failure is logged at error level. An unlabelled issue costs one
// click in triage; a report that evaporated because someone renamed
// a label costs the whole report, and nobody ever learns it happened.
//
// The retry is confined to statuses that mean GitHub refused the
// request outright (403 permission, 422 validation), because those
// are the ones where the issue was definitively NOT created. A
// timeout or a 5xx is never retried: GitHub may well have created
// the issue already, and a duplicate is worse than a 502 the
// reporter can act on.
func fileBugIssue(ctx context.Context, c Config, title, body, label string) (string, int, string, error) {
	var labels []string
	if label != "" {
		labels = []string{label}
	}
	url, number, err := c.BugReporter.CreateIssue(ctx, title, body, labels)
	if err == nil {
		return url, number, label, nil
	}
	if len(labels) == 0 || !bugRequestRejected(err) {
		return "", 0, "", err
	}
	url, number, retryErr := c.BugReporter.CreateIssue(ctx, title, body, nil)
	if retryErr != nil {
		// Report the FIRST error: it describes the actual rejection,
		// and the unlabelled retry failing the same way adds nothing.
		return "", 0, "", err
	}
	logBugError(c, "labels rejected by GitHub — issue filed unlabelled",
		"label", label, "issue", url, "err", err)
	return url, number, "", nil
}

// bugRequestRejected reports whether err is GitHub refusing the
// request itself, as opposed to failing to deliver it.
func bugRequestRejected(err error) bool {
	var sc statusCoder
	if !errors.As(err, &sc) {
		return false
	}
	switch sc.StatusCode() {
	case http.StatusForbidden, http.StatusUnprocessableEntity:
		return true
	default:
		return false
	}
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

// buildBugArtifacts stores the images and pins whatever the kind
// asks for, returning nil when there is nothing to store or no store
// to store it in.
//
// The kind gates the pins BEFORE they are read, not after: an idea
// never causes the server to copy a replay at all, so the cost isn't
// paid and then thrown away.
func buildBugArtifacts(c Config, p auth.Principal, kind bugKind, bctx *bugReportContext, images [][]byte) (*bugArtifacts, error) {
	if !c.BugStore.Enabled() {
		if len(images) > 0 {
			return nil, httpError(http.StatusServiceUnavailable, "attachments are not configured on this server")
		}
		return nil, nil
	}
	var replaySrc string
	if kind.Replay {
		replaySrc = bugReplaySource(c, p, bctx)
	}
	var gameLog []byte
	if kind.GameLog {
		gameLog = bugGameLog(c, p, bctx)
	}
	if len(images) == 0 && replaySrc == "" && len(gameLog) == 0 {
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
	if len(gameLog) > 0 {
		if _, err := rep.PinGameLog(gameLog); err != nil && !errors.Is(err, bugstore.ErrNotFound) {
			logBugStoreWarning(c, "pin game log", err)
		}
	}
	return out, nil
}

// bugGameLog renders the game's public log (ADR 0033 §4) as text to
// pin alongside the replay, or nil when there is no game to render.
//
// Entitlement mirrors bugReplaySource exactly: admins anywhere,
// everyone else only for the game their session is bound to. As with
// the replay, that check is not what protects the contents — the
// pinned copy is admin-only to read — it stops a session in game A
// from making the server render game B's history.
//
// The UNFILTERED projection is what gets pinned. That is the whole
// point of an admin-only artifact: a report saying "the game let them
// cast that twice" is triaged against what the server believed, not
// against one seat's redacted view.
func bugGameLog(c Config, p auth.Principal, bctx *bugReportContext) []byte {
	if bctx == nil || bctx.GameID == "" || c.Lobby == nil {
		return nil
	}
	id, err := uuid.Parse(bctx.GameID)
	if err != nil {
		return nil
	}
	if p.Role != auth.RoleAdmin && p.GameID != id {
		return nil
	}
	room := c.Lobby.RoomOf(id)
	if room == nil || room.Game == nil {
		return nil
	}
	entries := protocol.ViewOfGame(room.Game).Log
	if len(entries) == 0 {
		return nil
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# public game log — game %s, last %d entries\n", id, len(entries))
	for _, e := range entries {
		fmt.Fprintf(&b, "%6d  t%-3d %-10s %s\n", e.Seq, e.Turn, e.Kind, e.Text)
	}
	return []byte(b.String())
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
	if !c.BugStore.AttachmentsEnabled() {
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

// bugPinnedGameLog handles GET /bugreport/{id}/gamelog — admin only.
//
// The contents are public information by construction, but the pin is
// the UNFILTERED projection and it is served under the same admin gate
// as the replay rather than inviting a per-viewer question this route
// has no seat to answer.
func bugPinnedGameLog(c Config, w http.ResponseWriter, r *http.Request) error {
	if !c.BugStore.Enabled() {
		return httpError(http.StatusServiceUnavailable, "bug report storage is not configured on this server")
	}
	f, info, err := c.BugStore.OpenGameLog(r.PathValue("id"))
	if err != nil {
		if errors.Is(err, bugstore.ErrNotFound) {
			return httpError(http.StatusNotFound, "no pinned game log for that report")
		}
		return httpError(http.StatusInternalServerError, "opening the pinned game log failed")
	}
	defer func() { _ = f.Close() }()
	name := r.PathValue("id") + ".gamelog.txt"
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	http.ServeContent(w, r, name, info.ModTime(), f)
	return nil
}

// bugIssue is the input to renderBugIssueBody — a struct rather than
// a positional argument list because the body now assembles from six
// independent sources and a seventh would have been unreadable.
type bugIssue struct {
	Principal auth.Principal
	Kind      bugKind
	KindName  string
	Desc      string
	Ctx       *bugReportContext
	Log       []bugLogEntry
	UserAgent string
	Now       time.Time
	ReportID  string
	Images    []bugstore.Image
	Replay    *bugstore.ReplayPin
	GameLog   *bugstore.GameLogPin
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
	if in.Kind.Noun != "" {
		// Also in the label — but written here too, because a label
		// can be stripped in triage or fail to apply at all
		// (fileBugIssue), and then this row is the only record of
		// what the reporter said they were filing.
		fmt.Fprintf(&b, "- Kind: %s (label `%s`)\n", in.Kind.Noun, in.Kind.Label)
	}
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
	case in.Kind.Replay && in.Ctx != nil && in.Ctx.GameID != "":
		// No pin: either storage is off or the game had produced no
		// replay yet. The live route is the fallback, with the caveat
		// that it disappears when the game is evicted. Kinds that
		// don't want a replay don't get the consolation prize either
		// — it would read as "we tried and failed".
		fmt.Fprintf(&b, "- Replay: `GET /games/%s/replay` (admin; not pinned — gone once the game is evicted)\n", clip(in.Ctx.GameID, bugFieldMax))
	}
	if in.GameLog != nil {
		// Referenced, never inlined — same rule as the replay. The log
		// is public information, but ADR 0017 §4 keeps GAME STATE out
		// of issue bodies as a category, and an admin-gated artifact
		// costs the operator one request.
		fmt.Fprintf(&b, "- Pinned game log: `GET /bugreport/%s/gamelog` (admin) — %d entries, %d KiB, what happened at the table\n",
			in.ReportID, in.GameLog.Lines-1, in.GameLog.Bytes>>10)
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

// logBugError reports a filing problem the report survived but an
// operator has to know about — today, only the unlabelled fallback.
// Error rather than Warn on purpose: this is the one failure mode in
// the feature that is otherwise invisible, because the reporter sees
// a perfectly successful "issue #123 filed".
func logBugError(c Config, msg string, args ...any) {
	if c.Log == nil {
		return
	}
	c.Log.Error("bugreport: "+msg, args...)
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
